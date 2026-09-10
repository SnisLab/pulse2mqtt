package vars

import (
	"flag"
	"pulse2mqtt/include/version"
)

var (
	// set vars
	Mqtt_cID string = "go_Pulse2MQTT_v" + version.Version
	Verbose bool
	Help bool
)

func Vars() {
	flag.BoolVar(&Verbose, "v", false, "Run logging in Verbose Mode")
	//flag.BoolVar(&daemon, "d", false, "Daemon - Set it to daemon mode")
	//flag.BoolVar(&dryRun, "t", false, "DryRun - change nothing")
	flag.BoolVar(&Help, "h", false, "Help - shows this help")
	flag.Parse()
}