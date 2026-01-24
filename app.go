package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

// App represents the application with all its components.
type App struct {
	http *http.Server
	mqtt *MQTTBroker
}

// NewApp creates and initializes a new application instance.
func NewApp(httpAddr, mqttAddr, mqttTopic string, log *logrus.Logger) (*App, error) {
	var app App

	// Create HTTP server
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`)) //nolint:errcheck,gosec
	})
	//nolint:gosec
	app.http = &http.Server{
		Addr:    httpAddr,
		Handler: mux,
	}

	// Create MQTT broker
	broker, err := NewMQTTBroker(mqttAddr, mqttTopic, log)
	if err != nil {
		return nil, fmt.Errorf("create MQTT broker: %w", err)
	}
	app.mqtt = broker

	return &app, nil
}

// Start starts all application services (MQTT broker and HTTP server).
func (a *App) Start() error {
	var g errgroup.Group

	// Start MQTT broker
	g.Go(func() error {
		err := a.mqtt.Start()
		if err != nil {
			return fmt.Errorf("MQTT broker error: %w", err)
		}
		return nil
	})

	// Start HTTP server
	g.Go(func() error {
		err := a.http.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server error: %w", err)
		}
		return nil
	})

	return g.Wait() //nolint:wrapcheck
}

// Stop gracefully stops all application services.
func (a *App) Stop() error {
	var errs []error

	// Shutdown MQTT broker
	if err := a.mqtt.Stop(); err != nil {
		errs = append(errs, fmt.Errorf("MQTT broker shutdown error: %w", err))
	}

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := a.http.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("HTTP server shutdown error: %w", err))
	}

	return errors.Join(errs...)
}
