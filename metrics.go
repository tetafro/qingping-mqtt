package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Sensor metrics.
	TemperatureGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qingping_temperature_celsius",
		Help: "Temperature in Celsius",
	}, []string{"mac"})
	HumidityGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qingping_humidity_percent",
		Help: "Humidity in percent",
	}, []string{"mac"})
	CO2Gauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qingping_co2_ppm",
		Help: "CO2 level in parts per million",
	}, []string{"mac"})
	PM1Gauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qingping_pm1_ugm3",
		Help: "PM1 particulate matter in mg/m3",
	}, []string{"mac"})
	PM25Gauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qingping_pm25_ugm3",
		Help: "PM2.5 particulate matter in mg/m3",
	}, []string{"mac"})
	PM10Gauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qingping_pm10_ugm3",
		Help: "PM10 particulate matter in mg/m3",
	}, []string{"mac"})
	TVOCGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qingping_tvoc_ppb",
		Help: "Total Volatile Organic Compounds in parts per billion",
	}, []string{"mac"})
	RadonGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qingping_radon_index",
		Help: "Radon index",
	}, []string{"mac"})
	BatteryGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "qingping_battery_percent",
		Help: "Battery level in percent",
	}, []string{"mac"})

	// Service metrics.
	MessagesReceivedCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "qingping_mqtt_messages_received_total",
		Help: "Total number of MQTT messages received by message type",
	}, []string{"type", "topic", "mac"})
	AcksSentCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "qingping_mqtt_acks_sent_total",
		Help: "Total number of acknowledgments sent to devices",
	}, []string{"topic"})
	ParseErrorsCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "qingping_mqtt_parse_errors_total",
		Help: "Total number of message parsing errors",
	}, []string{"topic"})
	AckErrorsCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "qingping_mqtt_ack_errors_total",
		Help: "Total number of acknowledgment send errors",
	}, []string{"topic"})
)

// SetMetrics sets metrics from provided sensor data.
func SetMetrics(mac string, data SensorData) {
	TemperatureGauge.WithLabelValues(mac).Set(data.Temperature.Value)
	HumidityGauge.WithLabelValues(mac).Set(data.Humidity.Value)
	CO2Gauge.WithLabelValues(mac).Set(data.CO2.Value)
	PM1Gauge.WithLabelValues(mac).Set(data.PM1.Value)
	PM25Gauge.WithLabelValues(mac).Set(data.PM25.Value)
	PM10Gauge.WithLabelValues(mac).Set(data.PM10.Value)
	TVOCGauge.WithLabelValues(mac).Set(data.TVOC.Value)
	RadonGauge.WithLabelValues(mac).Set(data.Radon.Value)
	BatteryGauge.WithLabelValues(mac).Set(data.Battery.Value)
}
