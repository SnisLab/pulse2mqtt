package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/DjSni/go-log"
	ex_mqtt "github.com/eclipse/paho.mqtt.golang"

	"pulse2mqtt/include/data"
	"pulse2mqtt/include/metrics"
	"pulse2mqtt/include/mqtt"
	"pulse2mqtt/include/settings"
	"pulse2mqtt/include/vars"
	"pulse2mqtt/include/version"
)

// doKeepAlive sends a keep-alive message every 60 seconds.
func doKeepAlive() {
	time.Sleep(60 * time.Second)
	for {
		log.Info("Keep alive")
		time.Sleep(60 * time.Second)
	}
}

// doMetrics retrieves and sends metrics data every 10 seconds.
func doMetrics(MqttClient ex_mqtt.Client) {
	log.Info("run Metrics")
	for {
		metrics.GetMetrics()
		mqtt.SendMetrics(MqttClient)
		time.Sleep(10 * time.Second)
	}
}

// doData retrieves and sends data every second.
func doData(MqttClient ex_mqtt.Client) {
	log.Info("run Data")
	for {
		data.GetData()
		mqtt.SendData(MqttClient)
		time.Sleep(time.Second)
	}
}

func main() {
	log.SetDebugLevel(3)
	log.Info("Welcome to Pulse2MQTT v" + version.Version)
	vars.Vars()

	if vars.Help {
		flag.PrintDefaults()
		return
	}
	if vars.Verbose {
		log.SetDebugLevel(1)
	}

	settings.LogConfig(settings.Load)

	log.Info("MQTT:")
	MqttClient := mqtt.Start()
	// wait 1 sec
	time.Sleep(time.Second)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go doKeepAlive()
	go doMetrics(MqttClient)
	go doData(MqttClient)

	sig := <-c
	log.Info("Received signal:", sig)
	mqtt.Stop(MqttClient)
}
