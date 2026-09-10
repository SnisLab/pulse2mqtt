package settings

import (
	log "github.com/DjSni/go-log"
	"os"

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

func readConfig() Config {
	var config Config

	if _, err := os.Stat(CONFIG_PATH_DEB); err == nil {
		// Open YAML file
		file, err := os.Open(CONFIG_PATH_DEB)
		if err != nil {
			log.Error(err.Error())
		}
		defer file.Close()

		// Decode YAML file to struct
		if file != nil {
			decoder := yaml.NewDecoder(file)
			if err := decoder.Decode(&config); err != nil {
				log.Error(err.Error())
			}
		}
	} else if _, err := os.Stat(CONFIG_PATH); err == nil {
		// Open YAML file
		file, err := os.Open(CONFIG_PATH)
		if err != nil {
			log.Error(err.Error())
		}
		defer file.Close()

		// Decode YAML file to struct
		if file != nil {
			decoder := yaml.NewDecoder(file)
			if err := decoder.Decode(&config); err != nil {
				log.Error(err.Error())
			}
		}
	}

	return config
}

var Load Config = readConfig()
