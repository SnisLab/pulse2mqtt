package mqtt

import (
	"encoding/json"
	"testing"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"

	"pulse2mqtt/include/data"
	"pulse2mqtt/include/metrics"
	"pulse2mqtt/include/settings"
	"pulse2mqtt/include/vars"
)

type mockToken struct{}

func (mockToken) Wait() bool                     { return true }
func (mockToken) WaitTimeout(time.Duration) bool { return true }
func (mockToken) Done() <-chan struct{} {
	channel := make(chan struct{})
	close(channel)
	return channel
}
func (mockToken) Error() error { return nil }

type mockPublisher struct {
	connected    bool
	disconnected bool
	topic        string
	qos          byte
	retained     bool
	payload      interface{}
}

func (client *mockPublisher) IsConnected() bool { return client.connected }
func (client *mockPublisher) Publish(topic string, qos byte, retained bool, payload interface{}) paho.Token {
	client.topic = topic
	client.qos = qos
	client.retained = retained
	client.payload = payload
	return mockToken{}
}
func (client *mockPublisher) Disconnect(uint) { client.disconnected = true }

func TestBuildDiscoveryConfig(t *testing.T) {
	original := settings.Load
	t.Cleanup(func() { settings.Load = original })

	settings.Load.Service.Pulse.Node = 1
	settings.Load.Service.Pulse.BatteryProfile = "lfb_aa"
	settings.Load.Service.Mqtt.Topics.Data = "pulse2mqtt/data"
	settings.Load.Service.Mqtt.Topics.Metrics = "pulse2mqtt/metrics"
	settings.Load.Service.HomeAssistant.DiscoveryPrefix = "homeassistant"
	settings.Load.Service.HomeAssistant.DeviceID = "pulse node 1"
	settings.Load.Service.HomeAssistant.DeviceName = "Tibber Pulse"

	payload, err := buildDiscoveryConfig()
	if err != nil {
		t.Fatalf("could not build discovery config: %v", err)
	}

	var config discoveryConfig
	if err := json.Unmarshal(payload, &config); err != nil {
		t.Fatalf("could not unmarshal discovery config: %v", err)
	}

	if got, want := config.Device.Identifiers[0], "pulse_node_1"; got != want {
		t.Errorf("unexpected device ID: got %q, want %q", got, want)
	}
	if got, want := config.AvailabilityTopic, "pulse2mqtt/pulse_node_1/status"; got != want {
		t.Errorf("unexpected availability topic: got %q, want %q", got, want)
	}
	if got, want := discoveryTopic(), "homeassistant/device/pulse_node_1/config"; got != want {
		t.Errorf("unexpected discovery topic: got %q, want %q", got, want)
	}

	energy := config.Components["energy_consumed"]
	if energy.DeviceClass != "energy" || energy.StateClass != "total_increasing" || energy.Unit != "kWh" {
		t.Errorf("unexpected energy metadata: %+v", energy)
	}
	power := config.Components["power"]
	if power.DeviceClass != "power" || power.StateClass != "measurement" || power.Unit != "W" {
		t.Errorf("unexpected power metadata: %+v", power)
	}
	battery := config.Components["battery_level"]
	if battery.DeviceClass != "battery" || battery.StateClass != "measurement" || battery.Unit != "%" {
		t.Errorf("unexpected battery metadata: %+v", battery)
	}
	if battery.ValueTemplate != "{{ value_json.NodeBatteryLevel }}" {
		t.Errorf("unexpected battery value template: %q", battery.ValueTemplate)
	}
	if config.Components["temperature"].EntityCategory != "diagnostic" {
		t.Error("temperature must be a diagnostic entity")
	}

	settings.Load.Service.Pulse.BatteryProfile = "regulated_1_5v"
	payload, err = buildDiscoveryConfig()
	if err != nil {
		t.Fatalf("could not build discovery config for regulated cells: %v", err)
	}
	config = discoveryConfig{}
	if err := json.Unmarshal(payload, &config); err != nil {
		t.Fatalf("could not unmarshal discovery config for regulated cells: %v", err)
	}
	if _, ok := config.Components["battery_level"]; !ok {
		t.Error("regulated cells must create a battery level entity")
	}
}

func TestCurrentMetricsMessageUsesBatteryProfile(t *testing.T) {
	originalSettings := settings.Load
	originalMetrics := metrics.MResult
	t.Cleanup(func() {
		settings.Load = originalSettings
		metrics.MResult = originalMetrics
	})

	metrics.MResult.NodeStatus.NodeBatteryVoltage = 2.95
	settings.Load.Service.Pulse.BatteryProfile = "lfb_aa"
	message := currentMetricsMessage()
	if message.NodeBatteryLevel == nil || *message.NodeBatteryLevel != 50 {
		t.Errorf("unexpected LFB battery level: %v", message.NodeBatteryLevel)
	}

	settings.Load.Service.Pulse.BatteryProfile = "regulated_1_5v"
	metrics.MResult.NodeStatus.NodeBatteryVoltage = 3.0
	message = currentMetricsMessage()
	if message.NodeBatteryLevel == nil || *message.NodeBatteryLevel != 80 {
		t.Errorf("regulated cells must report 80 while active, got %v", message.NodeBatteryLevel)
	}
	payload, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("could not marshal regulated metrics message: %v", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(payload, &fields); err != nil {
		t.Fatalf("could not unmarshal regulated metrics message: %v", err)
	}
	if _, ok := fields["NodeBatteryLevel"]; !ok {
		t.Error("regulated metrics payload must contain NodeBatteryLevel")
	}

	metrics.MResult.NodeStatus.NodeBatteryVoltage = 2.79
	message = currentMetricsMessage()
	if message.NodeBatteryLevel == nil || *message.NodeBatteryLevel != 0 {
		t.Errorf("regulated cells must report 0 below the threshold, got %v", message.NodeBatteryLevel)
	}
}

func TestSendDataPublishesExpectedPayload(t *testing.T) {
	originalSettings := settings.Load
	originalData := data.DResult
	t.Cleanup(func() {
		settings.Load = originalSettings
		data.DResult = originalData
	})

	settings.Load.Service.Mqtt.Topics.Data = "pulse/data"
	data.DResult.NodeValue.Total.Consume = "1.2345 kWh"
	data.DResult.NodeValue.Total.Feed = "0.0000 kWh"
	data.DResult.NodeValue.Current.Consume = "42 W"

	client := &mockPublisher{connected: true}
	sendData(client)

	if client.topic != "pulse/data" || client.qos != 0 || !client.retained {
		t.Fatalf("unexpected publish options: topic=%q qos=%d retained=%t", client.topic, client.qos, client.retained)
	}

	var message Message
	if err := json.Unmarshal([]byte(client.payload.(string)), &message); err != nil {
		t.Fatalf("could not decode data payload: %v", err)
	}
	if message.Stromverbrauch != "1.2345 kWh" || message.Stromeinspeißung != "0.0000 kWh" || message.Aktuellerverbrauch != "42 W" {
		t.Fatalf("unexpected data payload: %+v", message)
	}
	if message.Time == "" {
		t.Fatal("data payload has no timestamp")
	}
}

func TestSendMetricsPublishesExpectedPayload(t *testing.T) {
	originalSettings := settings.Load
	originalMetrics := metrics.MResult
	t.Cleanup(func() {
		settings.Load = originalSettings
		metrics.MResult = originalMetrics
	})

	settings.Load.Service.Mqtt.Topics.Metrics = "pulse/metrics"
	settings.Load.Service.Pulse.BatteryProfile = "alkaline"
	metrics.MResult.NodeStatus.NodeBatteryVoltage = 2.6
	metrics.MResult.NodeStatus.NodeTemperature = 21.5
	metrics.MResult.NodeStatus.NodeAvgRssi = -55
	metrics.MResult.NodeStatus.MeterMsgCountSent = 12
	metrics.MResult.NodeStatus.MeterPkgCountSent = 4
	metrics.MResult.HubAttachments.NodeVersion = "1.2.3"

	client := &mockPublisher{connected: true}
	sendMetrics(client)

	if client.topic != "pulse/metrics" || client.qos != 0 || !client.retained {
		t.Fatalf("unexpected publish options: topic=%q qos=%d retained=%t", client.topic, client.qos, client.retained)
	}

	var message metricsMessage
	if err := json.Unmarshal([]byte(client.payload.(string)), &message); err != nil {
		t.Fatalf("could not decode metrics payload: %v", err)
	}
	if message.NodeBatteryLevel == nil || *message.NodeBatteryLevel != 50 {
		t.Fatalf("unexpected battery level: %v", message.NodeBatteryLevel)
	}
	if message.NodeVersion != "1.2.3" || message.MeterMsgCountSent != "12" {
		t.Fatalf("unexpected metrics payload: %+v", message)
	}
}

func TestSendDoesNothingWhenDisconnected(t *testing.T) {
	dataClient := &mockPublisher{}
	metricsClient := &mockPublisher{}

	sendData(dataClient)
	sendMetrics(metricsClient)

	if dataClient.topic != "" || metricsClient.topic != "" {
		t.Fatal("disconnected clients must not receive messages")
	}
}

func TestStopPublishesOfflineBeforeDisconnect(t *testing.T) {
	originalSettings := settings.Load
	t.Cleanup(func() { settings.Load = originalSettings })

	settings.Load.Service.HomeAssistant.Discovery = true
	settings.Load.Service.Pulse.Node = 1
	client := &mockPublisher{connected: true}

	Stop(client)

	if client.topic != "pulse2mqtt/pulse2mqtt_1/status" {
		t.Fatalf("unexpected availability topic: %q", client.topic)
	}
	if client.payload != "offline" || !client.retained {
		t.Fatalf("unexpected offline message: payload=%v retained=%t", client.payload, client.retained)
	}
	if !client.disconnected {
		t.Fatal("client was not disconnected")
	}
}

func TestStopDisconnectsWithoutPublishingWhenDiscoveryDisabled(t *testing.T) {
	originalSettings := settings.Load
	t.Cleanup(func() { settings.Load = originalSettings })

	settings.Load.Service.HomeAssistant.Discovery = false
	client := &mockPublisher{connected: true}

	Stop(client)

	if client.topic != "" || !client.disconnected {
		t.Fatalf("unexpected stop behavior: topic=%q disconnected=%t", client.topic, client.disconnected)
	}
}

func TestClientOptions(t *testing.T) {
	originalSettings := settings.Load
	originalClientID := vars.Mqtt_cID
	t.Cleanup(func() {
		settings.Load = originalSettings
		vars.Mqtt_cID = originalClientID
	})

	settings.Load.Service.Mqtt.Host = "mqtt.example"
	settings.Load.Service.Mqtt.Port = 1884
	settings.Load.Service.Mqtt.User = "mqtt-user"
	settings.Load.Service.Mqtt.Pass = "mqtt-pass"
	settings.Load.Service.HomeAssistant.Discovery = true
	settings.Load.Service.Pulse.Node = 3
	vars.Mqtt_cID = "test-client"

	opts := clientOptions()
	if len(opts.Servers) != 1 || opts.Servers[0].String() != "mqtt://mqtt.example:1884" {
		t.Fatalf("unexpected broker: %+v", opts.Servers)
	}
	if opts.ClientID != "pulse2mqtt-pulse2mqtt_3" || opts.Username != "mqtt-user" || opts.Password != "mqtt-pass" {
		t.Fatalf("unexpected client credentials: id=%q user=%q", opts.ClientID, opts.Username)
	}
	if !opts.AutoReconnect || opts.KeepAlive != 30 || !opts.WillEnabled {
		t.Fatalf("unexpected connection options: reconnect=%t keepalive=%d will=%t", opts.AutoReconnect, opts.KeepAlive, opts.WillEnabled)
	}
	if opts.WillTopic != "pulse2mqtt/pulse2mqtt_3/status" || string(opts.WillPayload) != "offline" || !opts.WillRetained {
		t.Fatalf("unexpected last will: topic=%q payload=%q retained=%t", opts.WillTopic, opts.WillPayload, opts.WillRetained)
	}
}
