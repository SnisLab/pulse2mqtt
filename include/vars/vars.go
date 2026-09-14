package vars

import (
	"flag"
	"pulse2mqtt/include/version"
)

type Options struct {
	MQTTClientID string
	Verbose      bool
	Help         bool
}

func MQTTClientID() string {
	return "go_Pulse2MQTT_v" + version.Version
}

func Vars() Options {
	options := Options{MQTTClientID: MQTTClientID()}
	flag.BoolVar(&options.Verbose, "v", false, "Run logging in Verbose Mode")
	//flag.BoolVar(&daemon, "d", false, "Daemon - Set it to daemon mode")
	//flag.BoolVar(&dryRun, "t", false, "DryRun - change nothing")
	flag.BoolVar(&options.Help, "h", false, "Help - shows this help")
	flag.Parse()
	return options
}
