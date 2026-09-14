package settings

import (
	"fmt"
	"os"
	"strings"

	log "github.com/DjSni/go-log"
	"gopkg.in/yaml.v3"
)

const CONFIG_PATH = "./settings.yaml"
const CONFIG_PATH_DEB = "/usr/local/etc/pulse2mqtt/settings.yaml"

type Config struct {
	Service struct {
		Mqtt struct {
			Host   string `yaml:"host"`
			Port   int    `yaml:"port"`
			User   string `yaml:"user"`
			Pass   string `yaml:"pass"`
			Topics struct {
				Data    string `yaml:"data"`
				Metrics string `yaml:"metrics"`
			} `yaml:"topics"`
		} `yaml:"mqtt"`
		Pulse struct {
			User           string `yaml:"user"`
			Password       string `yaml:"password"`
			IP             string `yaml:"ip"`
			Node           int    `yaml:"node"`
			BatteryProfile string `yaml:"battery_profile"`
		} `yaml:"pulse"`
		HomeAssistant struct {
			Discovery       bool   `yaml:"discovery"`
			DiscoveryPrefix string `yaml:"discovery_prefix"`
			DeviceID        string `yaml:"device_id"`
			DeviceName      string `yaml:"device_name"`
		} `yaml:"home_assistant"`
	} `yaml:"service"`
}

func LogConfig(config Config) {
	log.Debug("Pulse Settings:")
	log.Debug(" User:		", config.Service.Pulse.User)
	log.Debug(" Password set:	", config.Service.Pulse.Password != "")
	log.Debug(" IP:		", config.Service.Pulse.IP)
	log.Debug(" Node:		", config.Service.Pulse.Node)
	log.Debug(" Battery profile: ", config.Service.Pulse.BatteryProfile)
	log.Debug("MQTT Settings:")
	log.Debug(" Host:		", config.Service.Mqtt.Host)
	log.Debug(" Port:		", config.Service.Mqtt.Port)
	log.Debug(" User:		", config.Service.Mqtt.User)
	log.Debug(" Password set:	", config.Service.Mqtt.Pass != "")
	log.Debug(" Data:		", config.Service.Mqtt.Topics.Data)
	log.Debug(" Metrics:	", config.Service.Mqtt.Topics.Metrics)
	log.Debug("Home Assistant Settings:")
	log.Debug(" Discovery:	", config.Service.HomeAssistant.Discovery)
	log.Debug(" Device ID:	", config.Service.HomeAssistant.DeviceID)
}

// Validate checks the settings required to start the application.
func (config Config) Validate() error {
	if strings.TrimSpace(config.Service.Mqtt.Host) == "" {
		return fmt.Errorf("MQTT host is required")
	}
	if config.Service.Mqtt.Port < 1 || config.Service.Mqtt.Port > 65535 {
		return fmt.Errorf("MQTT port must be between 1 and 65535")
	}
	if strings.TrimSpace(config.Service.Mqtt.Topics.Data) == "" {
		return fmt.Errorf("MQTT data topic is required")
	}
	if strings.TrimSpace(config.Service.Mqtt.Topics.Metrics) == "" {
		return fmt.Errorf("MQTT metrics topic is required")
	}
	if strings.TrimSpace(config.Service.Pulse.IP) == "" {
		return fmt.Errorf("Pulse IP or hostname is required")
	}
	if config.Service.Pulse.Node < 1 {
		return fmt.Errorf("Pulse node must be greater than 0")
	}
	return nil
}

func readConfig() (Config, error) {
	path := ""

	if _, err := os.Stat(CONFIG_PATH_DEB); err == nil {
		path = CONFIG_PATH_DEB
	} else if _, err := os.Stat(CONFIG_PATH); err == nil {
		path = CONFIG_PATH
	} else {
		return Config{}, fmt.Errorf("configuration file not found (checked %s and %s)", CONFIG_PATH_DEB, CONFIG_PATH)
	}
	return loadConfig(path)
}

func loadConfig(path string) (Config, error) {
	var config Config
	file, err := os.Open(path)
	if err != nil {
		return config, fmt.Errorf("open configuration file %s: %w", path, err)
	}
	defer file.Close()

	if err := yaml.NewDecoder(file).Decode(&config); err != nil {
		return config, fmt.Errorf("decode configuration file %s: %w", path, err)
	}
	return config, nil
}

var (
	Load      Config
	LoadError error
)

func init() {
	Load, LoadError = readConfig()
}
