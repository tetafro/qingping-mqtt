package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Sensor metrics.
	temperatureGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qingping_temperature_celsius",
		Help: "Temperature in Celsius",
	}, []string{"mac"})
	humidityGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qingping_humidity_percent",
		Help: "Humidity in percent",
	}, []string{"mac"})
	co2Gauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qingping_co2_ppm",
		Help: "CO2 level in parts per million",
	}, []string{"mac"})
	pm1Gauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qingping_pm1_ugm3",
		Help: "PM1 particulate matter in mg/m3",
	}, []string{"mac"})
	pm25Gauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qingping_pm25_ugm3",
		Help: "PM2.5 particulate matter in mg/m3",
	}, []string{"mac"})
	pm10Gauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qingping_pm10_ugm3",
		Help: "PM10 particulate matter in mg/m3",
	}, []string{"mac"})
	tvocGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qingping_tvoc_ppb",
		Help: "Total Volatile Organic Compounds in parts per billion",
	}, []string{"mac"})
	radonGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qingping_radon_index",
		Help: "Radon index",
	}, []string{"mac"})
	batteryGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qingping_battery_percent",
		Help: "Battery level in percent",
	}, []string{"mac"})

	// Service metrics.
	messagesReceivedCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "qingping_mqtt_messages_received_total",
		Help: "Total number of MQTT messages received by message type",
	}, []string{"type"})
	acksSentCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "qingping_mqtt_acks_sent_total",
		Help: "Total number of acknowledgments sent to devices",
	})
	parseErrorsCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "qingping_mqtt_parse_errors_total",
		Help: "Total number of message parsing errors",
	})
	ackErrorsCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "qingping_mqtt_ack_errors_total",
		Help: "Total number of acknowledgment send errors",
	})
)
