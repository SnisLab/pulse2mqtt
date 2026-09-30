package mqtt

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/DjSni/go-log"
	paho "github.com/eclipse/paho.mqtt.golang"

	"pulse2mqtt/include/data"
	"pulse2mqtt/include/metrics"
	"pulse2mqtt/include/settings"
	"pulse2mqtt/include/vars"
	"pulse2mqtt/include/version"
)

const supportURL = "https://github.com/SnisLab/pulse2mqtt"

type discoveryDevice struct {
	Identifiers  []string `json:"identifiers"`
	Name         string   `json:"name"`
	Manufacturer string   `json:"manufacturer"`
	Model        string   `json:"model"`
	SWVersion    string   `json:"sw_version"`
}

type discoveryOrigin struct {
	Name       string `json:"name"`
	SWVersion  string `json:"sw_version"`
	SupportURL string `json:"support_url"`
}

type discoveryComponent struct {
	Platform         string `json:"platform"`
	UniqueID         string `json:"unique_id"`
	Name             string `json:"name"`
	StateTopic       string `json:"state_topic"`
	ValueTemplate    string `json:"value_template"`
	DeviceClass      string `json:"device_class,omitempty"`
	StateClass       string `json:"state_class,omitempty"`
	Unit             string `json:"unit_of_measurement,omitempty"`
	EntityCategory   string `json:"entity_category,omitempty"`
	EnabledByDefault *bool  `json:"enabled_by_default,omitempty"`
}

type discoveryConfig struct {
	Device              discoveryDevice               `json:"device"`
	Origin              discoveryOrigin               `json:"origin"`
	AvailabilityTopic   string                        `json:"availability_topic"`
	PayloadAvailable    string                        `json:"payload_available"`
	PayloadNotAvailable string                        `json:"payload_not_available"`
	Components          map[string]discoveryComponent `json:"components"`
}

func discoverySettings() (prefix string, deviceID string, deviceName string) {
	prefix = settings.Load.Service.HomeAssistant.DiscoveryPrefix
	if prefix == "" {
		prefix = "homeassistant"
	}
	prefix = strings.Trim(prefix, "/")
	if prefix == "" {
		prefix = "homeassistant"
	}

	deviceID = settings.Load.Service.HomeAssistant.DeviceID
	if deviceID == "" {
		deviceID = "pulse2mqtt_" + strconv.Itoa(settings.Load.Service.Pulse.Node)
	}
	deviceID = sanitizeID(deviceID)
	if deviceID == "" {
		deviceID = "pulse2mqtt_" + strconv.Itoa(settings.Load.Service.Pulse.Node)
	}

	deviceName = settings.Load.Service.HomeAssistant.DeviceName
	if deviceName == "" {
		deviceName = "Tibber Pulse"
	}

	return
}

func sanitizeID(value string) string {
	var result strings.Builder
	for _, char := range value {
		if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || char == '_' || char == '-' {
			result.WriteRune(char)
		} else {
			result.WriteByte('_')
		}
	}
	return result.String()
}

func availabilityTopic() string {
	_, deviceID, _ := discoverySettings()
	return "pulse2mqtt/" + deviceID + "/status"
}

func discoveryTopic() string {
	prefix, deviceID, _ := discoverySettings()
	return prefix + "/device/" + deviceID + "/config"
}

func buildDiscoveryConfig() ([]byte, error) {
	_, deviceID, deviceName := discoverySettings()
	disabled := false
	uniqueID := func(suffix string) string { return deviceID + "_" + suffix }

	config := discoveryConfig{
		Device: discoveryDevice{
			Identifiers:  []string{deviceID},
			Name:         deviceName,
			Manufacturer: "Tibber",
			Model:        "Pulse",
			SWVersion:    version.Version,
		},
		Origin: discoveryOrigin{
			Name:       "pulse2mqtt",
			SWVersion:  version.Version,
			SupportURL: supportURL,
		},
		AvailabilityTopic:   availabilityTopic(),
		PayloadAvailable:    "online",
		PayloadNotAvailable: "offline",
		Components: map[string]discoveryComponent{
			"energy_consumed": {
				Platform:      "sensor",
				UniqueID:      uniqueID("energy_consumed"),
				Name:          "Energy consumption",
				StateTopic:    settings.Load.Service.Mqtt.Topics.Data,
				ValueTemplate: "{{ value_json.Stromverbrauch.split(' ')[0] }}",
				DeviceClass:   "energy",
				StateClass:    "total_increasing",
				Unit:          "kWh",
			},
			"energy_feed": {
				Platform:      "sensor",
				UniqueID:      uniqueID("energy_feed"),
				Name:          "Energy feed-in",
				StateTopic:    settings.Load.Service.Mqtt.Topics.Data,
				ValueTemplate: "{{ value_json['Stromeinspeißung'].split(' ')[0] }}",
				DeviceClass:   "energy",
				StateClass:    "total_increasing",
				Unit:          "kWh",
			},
			"power": {
				Platform:      "sensor",
				UniqueID:      uniqueID("power"),
				Name:          "Power",
				StateTopic:    settings.Load.Service.Mqtt.Topics.Data,
				ValueTemplate: "{{ value_json.Aktuellerverbrauch.split(' ')[0] }}",
				DeviceClass:   "power",
				StateClass:    "measurement",
				Unit:          "W",
			},
			"battery_voltage": {
				Platform:         "sensor",
				UniqueID:         uniqueID("battery_voltage"),
				Name:             "Battery voltage",
				StateTopic:       settings.Load.Service.Mqtt.Topics.Metrics,
				ValueTemplate:    "{{ value_json.NodeBatteryVoltage }}",
				DeviceClass:      "voltage",
				StateClass:       "measurement",
				Unit:             "V",
				EntityCategory:   "diagnostic",
				EnabledByDefault: &disabled,
			},
			"temperature": {
				Platform:       "sensor",
				UniqueID:       uniqueID("temperature"),
				Name:           "Temperature",
				StateTopic:     settings.Load.Service.Mqtt.Topics.Metrics,
				ValueTemplate:  "{{ value_json.NodeTemperature }}",
				DeviceClass:    "temperature",
				StateClass:     "measurement",
				Unit:           "°C",
				EntityCategory: "diagnostic",
			},
			"signal_strength": {
				Platform:         "sensor",
				UniqueID:         uniqueID("signal_strength"),
				Name:             "Signal strength",
				StateTopic:       settings.Load.Service.Mqtt.Topics.Metrics,
				ValueTemplate:    "{{ value_json.NodeAvgRssi }}",
				DeviceClass:      "signal_strength",
				StateClass:       "measurement",
				Unit:             "dBm",
				EntityCategory:   "diagnostic",
				EnabledByDefault: &disabled,
			},
			"messages_sent": {
				Platform:         "sensor",
				UniqueID:         uniqueID("messages_sent"),
				Name:             "Messages sent",
				StateTopic:       settings.Load.Service.Mqtt.Topics.Metrics,
				ValueTemplate:    "{{ value_json.MeterMsgCountSent }}",
				EntityCategory:   "diagnostic",
				EnabledByDefault: &disabled,
			},
			"packages_sent": {
				Platform:         "sensor",
				UniqueID:         uniqueID("packages_sent"),
				Name:             "Packages sent",
				StateTopic:       settings.Load.Service.Mqtt.Topics.Metrics,
				ValueTemplate:    "{{ value_json.MeterPkgCountSent }}",
				EntityCategory:   "diagnostic",
				EnabledByDefault: &disabled,
			},
			"node_version": {
				Platform:         "sensor",
				UniqueID:         uniqueID("node_version"),
				Name:             "Node version",
				StateTopic:       settings.Load.Service.Mqtt.Topics.Metrics,
				ValueTemplate:    "{{ value_json.NodeVersion }}",
				EntityCategory:   "diagnostic",
				EnabledByDefault: &disabled,
			},
			"invalid_meter_readings": {
				Platform:       "sensor",
				UniqueID:       uniqueID("invalid_meter_readings"),
				Name:           "Invalid meter readings",
				StateTopic:     settings.Load.Service.Mqtt.Topics.Metrics,
				ValueTemplate:  "{{ value_json.InvalidMeterReadingsCount }}",
				EntityCategory: "diagnostic",
			},
			"valid_meter_readings": {
				Platform:       "sensor",
				UniqueID:       uniqueID("valid_meter_readings"),
				Name:           "Valid meter readings",
				StateTopic:     settings.Load.Service.Mqtt.Topics.Metrics,
				ValueTemplate:  "{{ value_json.ValidMeterReadingsCount }}",
				EntityCategory: "diagnostic",
			},
			"meter_uart_errors_9600": {
				Platform:       "sensor",
				UniqueID:       uniqueID("meter_uart_errors_9600"),
				Name:           "Meter UART errors (9600 baud)",
				StateTopic:     settings.Load.Service.Mqtt.Topics.Metrics,
				ValueTemplate:  "{{ value_json.MeterUARTErrorCount9600 }}",
				EntityCategory: "diagnostic",
			},
			"meter_messages_delta": {
				Platform:       "sensor",
				UniqueID:       uniqueID("meter_messages_delta"),
				Name:           "Meter messages (interval)",
				StateTopic:     settings.Load.Service.Mqtt.Topics.Metrics,
				ValueTemplate:  "{{ value_json.MeterMsgCountSentDelta }}",
				EntityCategory: "diagnostic",
			},
			"meter_packages_delta": {
				Platform:       "sensor",
				UniqueID:       uniqueID("meter_packages_delta"),
				Name:           "Meter packages (interval)",
				StateTopic:     settings.Load.Service.Mqtt.Topics.Metrics,
				ValueTemplate:  "{{ value_json.MeterPkgCountSentDelta }}",
				EntityCategory: "diagnostic",
			},
			"hub_meter_messages_received": {
				Platform:       "sensor",
				UniqueID:       uniqueID("hub_meter_messages_received"),
				Name:           "Hub meter messages received",
				StateTopic:     settings.Load.Service.Mqtt.Topics.Metrics,
				ValueTemplate:  "{{ value_json.HubMeterMsgCountReceived }}",
				EntityCategory: "diagnostic",
			},
			"hub_meter_packages_received": {
				Platform:       "sensor",
				UniqueID:       uniqueID("hub_meter_packages_received"),
				Name:           "Hub meter packages received",
				StateTopic:     settings.Load.Service.Mqtt.Topics.Metrics,
				ValueTemplate:  "{{ value_json.HubMeterPkgCountReceived }}",
				EntityCategory: "diagnostic",
			},
			"hub_meter_messages_received_delta": {
				Platform:       "sensor",
				UniqueID:       uniqueID("hub_meter_messages_received_delta"),
				Name:           "Hub meter messages (interval)",
				StateTopic:     settings.Load.Service.Mqtt.Topics.Metrics,
				ValueTemplate:  "{{ value_json.HubMeterMsgCountReceivedDelta }}",
				EntityCategory: "diagnostic",
			},
			"hub_meter_packages_received_delta": {
				Platform:       "sensor",
				UniqueID:       uniqueID("hub_meter_packages_received_delta"),
				Name:           "Hub meter packages (interval)",
				StateTopic:     settings.Load.Service.Mqtt.Topics.Metrics,
				ValueTemplate:  "{{ value_json.HubMeterPkgCountReceivedDelta }}",
				EntityCategory: "diagnostic",
			},
			"hub_corrupt_readings_received": {
				Platform:       "sensor",
				UniqueID:       uniqueID("hub_corrupt_readings_received"),
				Name:           "Hub corrupt readings (total)",
				StateTopic:     settings.Load.Service.Mqtt.Topics.Metrics,
				ValueTemplate:  "{{ value_json.HubCorruptReadingCountRecv }}",
				EntityCategory: "diagnostic",
			},
			"hub_corrupt_readings_delta": {
				Platform:       "sensor",
				UniqueID:       uniqueID("hub_corrupt_readings_delta"),
				Name:           "Hub corrupt readings (interval)",
				StateTopic:     settings.Load.Service.Mqtt.Topics.Metrics,
				ValueTemplate:  "{{ value_json.HubCorruptReadingCountReceivedDelta }}",
				EntityCategory: "diagnostic",
			},
			"packet_delivery_rate": {
				Platform:       "sensor",
				UniqueID:       uniqueID("packet_delivery_rate"),
				Name:           "Packet delivery rate",
				StateTopic:     settings.Load.Service.Mqtt.Topics.Metrics,
				ValueTemplate:  "{{ value_json.PacketDeliveryRate }}",
				Unit:           "%",
				EntityCategory: "diagnostic",
			},
			"node_available": {
				Platform:       "binary_sensor",
				UniqueID:       uniqueID("node_available"),
				Name:           "Pulse node available",
				StateTopic:     settings.Load.Service.Mqtt.Topics.Metrics,
				ValueTemplate:  "{{ 'ON' if value_json.NodeAvailable else 'OFF' }}",
				DeviceClass:    "connectivity",
				EntityCategory: "diagnostic",
			},
			"last_data_age": {
				Platform:       "sensor",
				UniqueID:       uniqueID("last_data_age"),
				Name:           "Age of last meter data",
				StateTopic:     settings.Load.Service.Mqtt.Topics.Metrics,
				ValueTemplate:  "{{ value_json.LastDataAgeSeconds }}",
				DeviceClass:    "duration",
				StateClass:     "measurement",
				Unit:           "s",
				EntityCategory: "diagnostic",
			},
			"wifi_rssi": {
				Platform:       "sensor",
				UniqueID:       uniqueID("wifi_rssi"),
				Name:           "Bridge Wi-Fi RSSI",
				StateTopic:     settings.Load.Service.Mqtt.Topics.Metrics,
				ValueTemplate:  "{{ value_json.WiFiRSSI }}",
				DeviceClass:    "signal_strength",
				StateClass:     "measurement",
				Unit:           "dBm",
				EntityCategory: "diagnostic",
			},
		},
	}
	if _, ok := metrics.EstimateBatteryLevel(0, settings.Load.Service.Pulse.BatteryProfile); ok {
		config.Components["battery_level"] = discoveryComponent{
			Platform:       "sensor",
			UniqueID:       uniqueID("battery_level"),
			Name:           "Battery level (estimated)",
			StateTopic:     settings.Load.Service.Mqtt.Topics.Metrics,
			ValueTemplate:  "{{ value_json.NodeBatteryLevel }}",
			DeviceClass:    "battery",
			StateClass:     "measurement",
			Unit:           "%",
			EntityCategory: "diagnostic",
		}
	}

	return json.Marshal(config)
}

func sendDiscovery(client mqttPublisher) {
	payload, err := buildDiscoveryConfig()
	if err != nil {
		log.Error("Can not create Home Assistant discovery message:", err)
		return
	}

	if token := client.Publish(discoveryTopic(), 1, true, payload); token.Wait() && token.Error() != nil {
		log.Error("Can not publish Home Assistant discovery message:", token.Error())
	}
}

func publishAvailability(client mqttPublisher, status string) {
	if token := client.Publish(availabilityTopic(), 1, true, status); token.Wait() && token.Error() != nil {
		log.Error("Can not publish Home Assistant availability:", token.Error())
	}
}

var connectHandler paho.OnConnectHandler = func(client paho.Client) {
	log.Info("MQTT connected")
	if !settings.Load.Service.HomeAssistant.Discovery {
		return
	}

	publishAvailability(client, "online")
	sendDiscovery(client)

	prefix, _, _ := discoverySettings()
	token := client.Subscribe(prefix+"/status", 0, func(client paho.Client, message paho.Message) {
		if string(message.Payload()) == "online" {
			sendDiscovery(client)
		}
	})
	if token.Wait() && token.Error() != nil {
		log.Error("Can not subscribe to Home Assistant status:", token.Error())
	}
}

var connectLostHandler paho.ConnectionLostHandler = func(client paho.Client, err error) {
	log.Info("Connect lost:", err)
}

// connect establishes a connection to the MQTT broker.
func clientOptions() *paho.ClientOptions {
	opts := paho.NewClientOptions().AddBroker("mqtt://" + settings.Load.Service.Mqtt.Host + ":" + strconv.Itoa(settings.Load.Service.Mqtt.Port))
	clientID := vars.MQTTClientID()
	if settings.Load.Service.HomeAssistant.Discovery {
		_, deviceID, _ := discoverySettings()
		clientID = "pulse2mqtt-" + deviceID
	}
	opts.SetClientID(clientID)
	if settings.Load.Service.Mqtt.User != "" {
		opts.SetUsername(settings.Load.Service.Mqtt.User)
	}
	if settings.Load.Service.Mqtt.Pass != "" {
		opts.SetPassword(settings.Load.Service.Mqtt.Pass)
	}
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	opts.AutoReconnect = true
	opts.KeepAlive = 30
	if settings.Load.Service.HomeAssistant.Discovery {
		opts.SetWill(availabilityTopic(), "offline", 1, true)
	}
	return opts
}

func connect() (client paho.Client, err error) {
	opts := clientOptions()
	client = paho.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		err = token.Error()
	}
	return
}

// publish sends a retained message to a specified MQTT topic.
type mqttPublisher interface {
	IsConnected() bool
	Publish(topic string, qos byte, retained bool, payload interface{}) paho.Token
}

func publish(client mqttPublisher, topic string, text string) {
	token := client.Publish(topic, 0, true, text)
	if token.Wait() && token.Error() != nil {
		log.Error("Can not publish MQTT message:", token.Error())
	}
}

type Message struct {
	Stromverbrauch     string `json:"Stromverbrauch"`
	Stromeinspeißung   string `json:"Stromeinspeißung"`
	Aktuellerverbrauch string `json:"Aktuellerverbrauch"`
	Time               string `json:"Time"`
}

type metricsMessage struct {
	NodeBatteryVoltage                  string   `json:"NodeBatteryVoltage"`
	NodeBatteryLevel                    *int     `json:"NodeBatteryLevel,omitempty"`
	NodeTemperature                     string   `json:"NodeTemperature"`
	NodeAvgRssi                         string   `json:"NodeAvgRssi"`
	MeterMsgCountSent                   string   `json:"MeterMsgCountSent"`
	MeterPkgCountSent                   string   `json:"MeterPkgCountSent"`
	HubMeterMsgCountReceived            string   `json:"HubMeterMsgCountReceived"`
	HubMeterPkgCountReceived            string   `json:"HubMeterPkgCountReceived"`
	NodeVersion                         string   `json:"NodeVersion"`
	InvalidMeterReadingsCount           *int     `json:"InvalidMeterReadingsCount,omitempty"`
	ValidMeterReadingsCount             *int     `json:"ValidMeterReadingsCount,omitempty"`
	MeterUARTErrorCount9600             *int     `json:"MeterUARTErrorCount9600,omitempty"`
	MeterMsgCountSentDelta              *int     `json:"MeterMsgCountSentDelta,omitempty"`
	MeterPkgCountSentDelta              *int     `json:"MeterPkgCountSentDelta,omitempty"`
	HubMeterMsgCountReceivedDelta       *int     `json:"HubMeterMsgCountReceivedDelta,omitempty"`
	HubMeterPkgCountReceivedDelta       *int     `json:"HubMeterPkgCountReceivedDelta,omitempty"`
	HubCorruptReadingCountRecv          *int     `json:"HubCorruptReadingCountRecv,omitempty"`
	HubCorruptReadingCountReceivedDelta *int     `json:"HubCorruptReadingCountReceivedDelta,omitempty"`
	PacketDeliveryRate                  *float64 `json:"PacketDeliveryRate,omitempty"`
	NodeAvailable                       *bool    `json:"NodeAvailable,omitempty"`
	LastDataAgeSeconds                  *float64 `json:"LastDataAgeSeconds,omitempty"`
	WiFiRSSI                            *int     `json:"WiFiRSSI,omitempty"`
	Time                                string   `json:"Time"`
}

func currentMetricsMessage(result metrics.Metrics) metricsMessage {
	batteryLevel, hasBatteryLevel := metrics.EstimateBatteryLevel(
		result.NodeStatus.NodeBatteryVoltage,
		settings.Load.Service.Pulse.BatteryProfile,
	)
	var batteryLevelValue *int
	if hasBatteryLevel {
		batteryLevelValue = &batteryLevel
	}

	var lastDataAgeSeconds *float64
	if result.Diagnostics.LastDataAgeMs != nil {
		age := float64(*result.Diagnostics.LastDataAgeMs) / 1000
		lastDataAgeSeconds = &age
	}
	return metricsMessage{
		NodeBatteryVoltage:                  fmt.Sprintf("%f", result.NodeStatus.NodeBatteryVoltage),
		NodeBatteryLevel:                    batteryLevelValue,
		NodeTemperature:                     fmt.Sprintf("%f", result.NodeStatus.NodeTemperature),
		NodeAvgRssi:                         fmt.Sprintf("%f", result.NodeStatus.NodeAvgRssi),
		MeterMsgCountSent:                   strconv.Itoa(result.NodeStatus.MeterMsgCountSent),
		MeterPkgCountSent:                   strconv.Itoa(result.NodeStatus.MeterPkgCountSent),
		HubMeterMsgCountReceived:            strconv.Itoa(result.HubAttachments.MeterReadingCountRecv),
		HubMeterPkgCountReceived:            strconv.Itoa(result.HubAttachments.MeterPkgCountRecv),
		NodeVersion:                         result.HubAttachments.NodeVersion,
		InvalidMeterReadingsCount:           result.Diagnostics.InvalidMeterReadingsCount,
		ValidMeterReadingsCount:             result.Diagnostics.ValidMeterReadingsCount,
		MeterUARTErrorCount9600:             result.Diagnostics.MeterUARTErrorCount9600,
		MeterMsgCountSentDelta:              result.Diagnostics.MeterMsgCountSentDelta,
		MeterPkgCountSentDelta:              result.Diagnostics.MeterPkgCountSentDelta,
		HubMeterMsgCountReceivedDelta:       result.Diagnostics.HubMeterMsgCountReceivedDelta,
		HubMeterPkgCountReceivedDelta:       result.Diagnostics.HubMeterPkgCountReceivedDelta,
		HubCorruptReadingCountRecv:          result.HubAttachments.MeterCorruptReadingCountRecv,
		HubCorruptReadingCountReceivedDelta: result.Diagnostics.HubCorruptReadingCountReceivedDelta,
		PacketDeliveryRate:                  result.Diagnostics.PacketDeliveryRate,
		NodeAvailable:                       result.Diagnostics.NodeAvailable,
		LastDataAgeSeconds:                  lastDataAgeSeconds,
		WiFiRSSI:                            result.Diagnostics.WiFiRSSI,
		Time:                                time.Now().Format(time.RFC3339),
	}
}

// SendData sends the current data to the MQTT broker.
func SendData(client paho.Client, result data.Data) {
	sendData(client, result)
}

func sendData(client mqttPublisher, result data.Data) {
	if !client.IsConnected() {
		log.Debug("MQTT reconnecting")
		return
	}

	payload, err := json.Marshal(Message{
		Stromverbrauch:     result.NodeValue.Total.Consume,
		Stromeinspeißung:   result.NodeValue.Total.Feed,
		Aktuellerverbrauch: result.NodeValue.Current.Consume,
		Time:               time.Now().Format(time.RFC3339),
	})
	if err != nil {
		log.Error("Can not create MQTT data message:", err)
		return
	}

	log.Debug("already connected to MQTT")
	publish(client, settings.Load.Service.Mqtt.Topics.Data, string(payload))
}

// SendMetrics sends the current metrics to the MQTT broker.
func SendMetrics(client paho.Client, result metrics.Metrics) {
	sendMetrics(client, result)
}

func sendMetrics(client mqttPublisher, result metrics.Metrics) {
	if !client.IsConnected() {
		log.Debug("MQTT reconnecting")
		return
	}

	payload, err := json.Marshal(currentMetricsMessage(result))
	if err != nil {
		log.Error("Can not create MQTT metrics message:", err)
		return
	}

	log.Debug("already connected to MQTT")
	publish(client, settings.Load.Service.Mqtt.Topics.Metrics, string(payload))
}

// Start initializes the MQTT client and connects to the broker.
func Start() paho.Client {
	client, err := connect()
	if err != nil {
		log.Error("Error connecting to MQTT:", err)
	}
	return client
}

// Stop disconnects the MQTT client from the broker.
type mqttConnection interface {
	mqttPublisher
	Disconnect(quiesce uint)
}

func Stop(client mqttConnection) {
	if client == nil {
		return
	}
	log.Warn("Disconnect MQTT")
	if client.IsConnected() && settings.Load.Service.HomeAssistant.Discovery {
		publishAvailability(client, "offline")
	}
	client.Disconnect(10)
}
