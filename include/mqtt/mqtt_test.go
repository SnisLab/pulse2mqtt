package mqtt

import (
	"encoding/json"
	"testing"

	"pulse2mqtt/include/metrics"
	"pulse2mqtt/include/settings"
)

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
