package settings

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestConfigUnmarshalsAllUserSettings(t *testing.T) {
	var config Config
	err := yaml.Unmarshal([]byte(`
service:
  mqtt:
    host: mqtt.example
    port: 1883
    user: mqtt-user
    pass: mqtt-pass
    topics:
      data: pulse/data
      metrics: pulse/metrics
  pulse:
    user: pulse-user
    password: pulse-pass
    ip: 192.0.2.10
    node: 7
    battery_profile: alkaline
  home_assistant:
    discovery: true
    discovery_prefix: custom
    device_id: pulse-7
    device_name: Test Pulse
`), &config)
	if err != nil {
		t.Fatalf("could not unmarshal config: %v", err)
	}

	if config.Service.Mqtt.Host != "mqtt.example" || config.Service.Mqtt.Port != 1883 {
		t.Fatalf("unexpected MQTT settings: %+v", config.Service.Mqtt)
	}
	if config.Service.Pulse.Node != 7 || config.Service.Pulse.BatteryProfile != "alkaline" {
		t.Fatalf("unexpected Pulse settings: %+v", config.Service.Pulse)
	}
	if !config.Service.HomeAssistant.Discovery || config.Service.HomeAssistant.DeviceID != "pulse-7" {
		t.Fatalf("unexpected Home Assistant settings: %+v", config.Service.HomeAssistant)
	}
}
