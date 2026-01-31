package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
)

func TestApp(t *testing.T) {
	ctx := context.Background()
	log := logrus.New()
	log.Out = io.Discard

	httpAddr := "127.0.0.1:18080"
	mqttAddr := "127.0.0.1:11883"

	// Set a short interval for testing
	HeartbeatInterval = 50 * time.Millisecond

	app, err := NewApp(httpAddr, mqttAddr, log)
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	go func() {
		if err := app.Start(ctx); err != nil {
			t.Errorf("Failed to start app: %v", err)
		}
	}()
	defer func() {
		if err := app.Stop(); err != nil {
			t.Errorf("Failed to stop app: %v", err)
		}
	}()

	// Wait for services to start
	time.Sleep(10 * time.Millisecond)

	t.Run("health endpoint", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("http://%s/health", httpAddr))
		if err != nil {
			t.Fatalf("Failed to call health endpoint: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("metrics endpoint", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("http://%s/metrics", httpAddr))
		if err != nil {
			t.Fatalf("Failed to call metrics endpoint: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("send mqtt message and verify metrics", func(t *testing.T) {
		message := `{
			"type": "17",
			"id": 12345,
			"need_ack": 1,
			"mac": "112233445566",
			"timestamp": 1594815555,
			"sensorData": [{
				"timestamp": {"value": 1592192453},
				"temperature": {"value": 23.5},
				"humidity": {"value": 45.2},
				"co2": {"value": 850},
				"pm25": {"value": 12.3},
				"pm10": {"value": 15.8},
				"battery": {"value": 85}
			}]
		}`

		err := sendMQTTMessage(t, mqttAddr, "qingping/test-device/up", message)
		if err != nil {
			t.Fatalf("Failed to send MQTT message: %v", err)
		}

		// Wait for message processing
		time.Sleep(10 * time.Millisecond)

		body, err := httpGet(fmt.Sprintf("http://%s/metrics", httpAddr))
		if err != nil {
			t.Fatalf("Failed to read metrics: %v", err)
		}

		if !strings.Contains(body, `qingping_mqtt_messages_received_total{mac="112233445566",topic="qingping/test-device/up",type="17"} 1`) {
			t.Fatal(`Expected qingping_mqtt_messages_received_total{mac="112233445566",topic="qingping/test-device/up",type="17"} to be 1`)
		}
		if !strings.Contains(body, `qingping_mqtt_acks_sent_total{topic="qingping/test-device/up"} 1`) {
			t.Fatal("Expected qingping_mqtt_acks_sent_total{topic=\"qingping/test-device/up\"} to be 1")
		}
		if !strings.Contains(body, `qingping_temperature_celsius{mac="112233445566"} 23.5`) {
			t.Fatal("Expected temperature metric to be 23.5")
		}
		if !strings.Contains(body, `qingping_humidity_percent{mac="112233445566"} 45.2`) {
			t.Fatal("Expected humidity metric to be 45.2")
		}
		if !strings.Contains(body, `qingping_co2_ppm{mac="112233445566"} 850`) {
			t.Error("Expected CO2 metric to be 850")
		}
	})

	t.Run("invalid json message increments parse error counter", func(t *testing.T) {
		message := `{"invalid json syntax`

		err := sendMQTTMessage(t, mqttAddr, "qingping/test-device/up", message)
		if err != nil {
			t.Fatalf("Failed to send MQTT message: %v", err)
		}

		// Wait for message processing
		time.Sleep(10 * time.Millisecond)

		body, err := httpGet(fmt.Sprintf("http://%s/metrics", httpAddr))
		if err != nil {
			t.Fatalf("Failed to read metrics: %v", err)
		}

		if !strings.Contains(body, `qingping_mqtt_parse_errors_total{topic="qingping/test-device/up"} 1`) {
			t.Error("Expected parse error counter to be incremented")
		}
	})

	t.Run("message without ack required", func(t *testing.T) {
		message := `{
			"type": "17",
			"id": 54321,
			"need_ack": 0,
			"mac": "112233445566",
			"timestamp": 1594815555,
			"sensorData": [{
				"timestamp": {"value": 1592192453},
				"temperature": {"value": 25.0},
				"humidity": {"value": 50.0}
			}]
		}`

		body, err := httpGet(fmt.Sprintf("http://%s/metrics", httpAddr))
		if err != nil {
			t.Fatalf("Failed to read metrics: %v", err)
		}
		acksBefore := strings.Count(body, `qingping_mqtt_acks_sent_total{topic="qingping/test-device/up"}`)

		err = sendMQTTMessage(t, mqttAddr, "qingping/test-device/up", message)
		if err != nil {
			t.Fatalf("Failed to send MQTT message: %v", err)
		}

		// Wait for message processing
		time.Sleep(10 * time.Millisecond)

		body, err = httpGet(fmt.Sprintf("http://%s/metrics", httpAddr))
		if err != nil {
			t.Fatalf("Failed to read metrics: %v", err)
		}
		acksAfter := strings.Count(body, `qingping_mqtt_acks_sent_total{topic="qingping/test-device/up"}`)

		if acksAfter != acksBefore {
			t.Error("Expected no acknowledgment to be sent")
		}
	})

	t.Run("heartbeat message type 13", func(t *testing.T) {
		message := `{
			"type": "13",
			"id": 11111,
			"wifi_mac": "112233445566"
		}`

		err := sendMQTTMessage(t, mqttAddr, "qingping/test-device/up", message)
		if err != nil {
			t.Fatalf("Failed to send MQTT message: %v", err)
		}

		// Wait for message processing
		time.Sleep(10 * time.Millisecond)

		body, err := httpGet(fmt.Sprintf("http://%s/metrics", httpAddr))
		if err != nil {
			t.Fatalf("Failed to read metrics: %v", err)
		}

		if !strings.Contains(body, `qingping_mqtt_messages_received_total{mac="112233445566",topic="qingping/test-device/up",type="13"} 1`) {
			t.Error("Expected heartbeat message to be counted")
		}
	})

	t.Run("heartbeat timeout resets metrics", func(t *testing.T) {
		// Send initial data for device 1
		msg := `{
			"type": "17",
			"wifi_mac": "MAC1",
			"id": 54321,
			"need_ack": 0,
			"timestamp": 1594815555,
			"sensorData": [{
				"temperature": {"value": 20.0},
				"humidity": {"value": 60.0}
			}]
		}`
		err := sendMQTTMessage(t, mqttAddr, "qingping/device1/up", msg)
		if err != nil {
			t.Fatalf("Failed to send MQTT message for device 1: %v", err)
		}

		// Send initial data for device 2
		msg = `{
			"type": "17",
			"wifi_mac": "MAC2",
			"id": 54322,
			"need_ack": 0,
			"timestamp": 1594815555,
			"sensorData": [{
				"temperature": {"value": 20.0},
				"humidity": {"value": 60.0}
			}]
		}`
		err = sendMQTTMessage(t, mqttAddr, "qingping/device2/up", msg)
		if err != nil {
			t.Fatalf("Failed to send MQTT message for device 2: %v", err)
		}

		// Wait for message processing
		time.Sleep(10 * time.Millisecond)

		// Keep sending heartbeats for device 1 only
		msg = `{"type": "13", "wifi_mac": "MAC1"}`
		for range 4 {
			err = sendMQTTMessage(t, mqttAddr, "qingping/device1/up", msg)
			if err != nil {
				t.Fatalf("Failed to send heartbeat for device 1: %v", err)
			}
			time.Sleep(HeartbeatInterval)
		}

		// Check metrics
		body, err := httpGet(fmt.Sprintf("http://%s/metrics", httpAddr))
		if err != nil {
			t.Fatalf("Failed to read metrics: %v", err)
		}

		// Device 1 should have its original metrics
		if !strings.Contains(body, `qingping_temperature_celsius{mac="MAC1"} 20`) {
			t.Errorf("Expected device 1 temperature to remain 20")
		}
		if !strings.Contains(body, `qingping_humidity_percent{mac="MAC1"} 60`) {
			t.Errorf("Expected device 1 humidity to remain 60")
		}

		// Device 2 should have empty metrics
		if !strings.Contains(body, `qingping_temperature_celsius{mac="MAC2"} 0`) {
			t.Errorf("Expected device 2 temperature to be reset to 0")
		}
		if !strings.Contains(body, `qingping_humidity_percent{mac="MAC2"} 0`) {
			t.Errorf("Expected device 2 humidity to be reset to 0")
		}
	})
}

func httpGet(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	return string(body), nil
}

func sendMQTTMessage(t *testing.T, addr, topic, payload string) error {
	t.Helper()

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", addr))
	opts.SetClientID("test-client")
	opts.SetConnectTimeout(2 * time.Second)
	opts.SetAutoReconnect(false)

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if !token.WaitTimeout(2 * time.Second) {
		return fmt.Errorf("connection timeout")
	}
	if err := token.Error(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Disconnect(250)

	// Publish message
	pubToken := client.Publish(topic, 0, false, payload)
	if !pubToken.WaitTimeout(2 * time.Second) {
		return fmt.Errorf("publish timeout")
	}
	if err := pubToken.Error(); err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	return nil
}
