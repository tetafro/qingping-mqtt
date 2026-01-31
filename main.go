// qingping-mqtt is an MQTT broker that receives metrics from Qingping Lite
// air monitor and exposes them as Prometheus metrics.
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

func main() {
	debug := flag.Bool("debug", false, "enable debug logs")
	httpAddr := flag.String("http-addr", "0.0.0.0:8080", "HTTP server listen address")
	mqttAddr := flag.String("mqtt-addr", "0.0.0.0:1883", "MQTT broker listen address")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	level := logrus.InfoLevel
	if *debug {
		level = logrus.DebugLevel
	}
	log.SetLevel(level)

	app, err := NewApp(*httpAddr, *mqttAddr, log)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	go func() {
		<-ctx.Done()
		log.Info("Stopping application")
		if err := app.Stop(); err != nil {
			log.Errorf("Error during shutdown: %v", err)
		}
	}()

	log.WithField("http_addr", *httpAddr).
		WithField("mqtt_addr", *mqttAddr).
		Info("Starting...")
	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}
	log.Info("Shutdown")
}
