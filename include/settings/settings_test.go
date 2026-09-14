package settings

import (
	"os"
	"path/filepath"
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

func TestConfigValidate(t *testing.T) {
	valid := Config{}
	valid.Service.Mqtt.Host = "mqtt.example"
	valid.Service.Mqtt.Port = 1883
	valid.Service.Mqtt.Topics.Data = "pulse/data"
	valid.Service.Mqtt.Topics.Metrics = "pulse/metrics"
	valid.Service.Pulse.IP = "pulse.example"
	valid.Service.Pulse.Node = 1

	if err := valid.Validate(); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}

	tests := []struct {
		name  string
		setup func(*Config)
	}{
		{name: "missing MQTT host", setup: func(config *Config) { config.Service.Mqtt.Host = "" }},
		{name: "invalid MQTT port", setup: func(config *Config) { config.Service.Mqtt.Port = 70000 }},
		{name: "missing data topic", setup: func(config *Config) { config.Service.Mqtt.Topics.Data = "" }},
		{name: "missing metrics topic", setup: func(config *Config) { config.Service.Mqtt.Topics.Metrics = "" }},
		{name: "missing Pulse address", setup: func(config *Config) { config.Service.Pulse.IP = "" }},
		{name: "invalid Pulse node", setup: func(config *Config) { config.Service.Pulse.Node = 0 }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := valid
			tt.setup(&config)
			if err := config.Validate(); err == nil {
				t.Fatal("invalid config was accepted")
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "settings.yaml")
	if err := os.WriteFile(path, []byte("service:\n  mqtt:\n    host: mqtt.example\n    port: 1883\n"), 0600); err != nil {
		t.Fatalf("could not create config fixture: %v", err)
	}

	config, err := loadConfig(path)
	if err != nil {
		t.Fatalf("could not load valid config: %v", err)
	}
	if config.Service.Mqtt.Host != "mqtt.example" || config.Service.Mqtt.Port != 1883 {
		t.Fatalf("unexpected loaded config: %+v", config.Service.Mqtt)
	}
}

func TestLoadConfigRejectsInvalidYAML(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "settings.yaml")
	if err := os.WriteFile(path, []byte("service: [invalid"), 0600); err != nil {
		t.Fatalf("could not create invalid config fixture: %v", err)
	}

	if _, err := loadConfig(path); err == nil {
		t.Fatal("invalid YAML was accepted")
	}
}

func TestLoadConfigReportsMissingFile(t *testing.T) {
	if _, err := loadConfig(filepath.Join(t.TempDir(), "missing.yaml")); err == nil {
		t.Fatal("missing config file was accepted")
	}
}

func TestReadConfigUsesEnvironmentPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "custom.yaml")
	if err := os.WriteFile(path, []byte("service:\n  mqtt:\n    host: custom.example\n    port: 1883\n"), 0600); err != nil {
		t.Fatalf("could not create config fixture: %v", err)
	}
	t.Setenv(CONFIG_PATH_ENV, path)

	config, err := readConfig()
	if err != nil {
		t.Fatalf("could not load config from environment path: %v", err)
	}
	if config.Service.Mqtt.Host != "custom.example" {
		t.Fatalf("unexpected config loaded from environment path: %q", config.Service.Mqtt.Host)
	}
}

func TestReadConfigReportsMissingEnvironmentPath(t *testing.T) {
	t.Setenv(CONFIG_PATH_ENV, filepath.Join(t.TempDir(), "missing.yaml"))

	if _, err := readConfig(); err == nil {
		t.Fatal("missing environment config file was accepted")
	}
}
