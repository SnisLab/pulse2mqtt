package main

import (
	"context"
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
	"pulse2mqtt/include/pulse"
	"pulse2mqtt/include/settings"
	"pulse2mqtt/include/vars"
	"pulse2mqtt/include/version"
)

// doKeepAlive sends a keep-alive message every 60 seconds.
func doKeepAlive(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Info("Keep alive")
		}
	}
}

// doMetrics retrieves and sends metrics data every 10 seconds.
func doMetrics(ctx context.Context, MqttClient ex_mqtt.Client) {
	log.Info("run Metrics")
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		mqtt.SendMetrics(MqttClient, metrics.GetMetrics())
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// doData retrieves and sends data every second.
func doData(ctx context.Context, MqttClient ex_mqtt.Client) {
	log.Info("run Data")
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		mqtt.SendData(MqttClient, data.GetData())
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func main() {
	log.SetDebugLevel(3)
	log.Info("Welcome to Pulse2MQTT v" + version.Version)
	options := vars.Vars()

	if options.Help {
		flag.PrintDefaults()
		return
	}
	if options.Verbose {
		log.SetDebugLevel(1)
	}

	if settings.LoadError != nil {
		log.Fatal("Can not load configuration:", settings.LoadError)
	}
	if err := settings.Load.Validate(); err != nil {
		log.Fatal("Invalid configuration:", err)
	}

	settings.LogConfig(settings.Load)
	if err := pulse.Initialize(); err != nil {
		log.Fatal("Can not detect Pulse API:", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Info("MQTT:")
	MqttClient := mqtt.Start()
	select {
	case <-ctx.Done():
		mqtt.Stop(MqttClient)
		return
	case <-time.After(time.Second):
	}

	go doKeepAlive(ctx)
	go doMetrics(ctx, MqttClient)
	go doData(ctx, MqttClient)

	<-ctx.Done()
	log.Info("Shutdown requested")
	mqtt.Stop(MqttClient)
}
