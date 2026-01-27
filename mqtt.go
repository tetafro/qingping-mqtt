package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
	"github.com/sirupsen/logrus"
)

// HeartBeatInterval is the expected interval of heartbeats from the device.
const HeartBeatInterval = 1 * time.Minute

// List of known message types.
const (
	RealTimeDataType = "12"
	HeadrbeatType    = "13"
	HistoryDataType  = "17"
)

// AllowedMessageTypes is the list of message types that the app can process.
var AllowedMessageTypes = []string{
	RealTimeDataType,
	HeadrbeatType,
	HistoryDataType,
}

// MQTTBroker wraps the MQTT server and provides message handling.
type MQTTBroker struct {
	server *mqtt.Server
	topic  string
}

// QingpingMessage represents the message envelope from Qingping devices.
type QingpingMessage struct {
	ID         int          `json:"id"`
	Type       string       `json:"type"`
	NeedAck    int          `json:"need_ack"`
	MAC        string       `json:"mac"`
	Timestamp  int64        `json:"timestamp"`
	SensorData []SensorData `json:"sensorData"`
}

// SensorData represents sensor readings in type "12" and "17" messages.
type SensorData struct {
	Timestamp   ValueWrapper `json:"timestamp"`
	Temperature ValueWrapper `json:"temperature"`
	Humidity    ValueWrapper `json:"humidity"`
	CO2         ValueWrapper `json:"co2"`
	PM1         ValueWrapper `json:"pm1"`
	PM25        ValueWrapper `json:"pm25"`
	PM10        ValueWrapper `json:"pm10"`
	TVOC        ValueWrapper `json:"tvoc"`
	Radon       ValueWrapper `json:"radon"`
	Battery     ValueWrapper `json:"battery"`
}

// ValueWrapper wraps sensor values with optional additional fields.
type ValueWrapper struct {
	Value           float64 `json:"value"`
	Status          *int    `json:"status,omitempty"`
	Level           *int    `json:"level,omitempty"`
	Unit            string  `json:"unit,omitempty"`
	StatusDuration  *int    `json:"status_duration,omitempty"`
	StatusStartTime *int64  `json:"status_start_time,omitempty"`
}

// AckResponse represents the acknowledgment message sent back to the device.
type AckResponse struct {
	Type      string `json:"type"`
	AckID     int    `json:"ack_id"`
	Code      int    `json:"code"`
	Timestamp int64  `json:"timestamp"`
	Desc      string `json:"desc,omitempty"`
}

// NewMQTTBroker creates and configures a new MQTT broker.
func NewMQTTBroker(addr, topic string, log *logrus.Logger) (*MQTTBroker, error) {
	opts := &mqtt.Options{
		InlineClient: true, // allow publishing from within hooks
		Logger:       slog.New(slog.DiscardHandler),
	}
	broker := &MQTTBroker{
		server: mqtt.New(opts),
		topic:  topic,
	}

	// Allow all connections (no authentication for simplicity)
	err := broker.server.AddHook(new(auth.AllowHook), nil)
	if err != nil {
		return nil, fmt.Errorf("add allow-all hook: %w", err)
	}

	// Add message handler hook
	hook := &MessageHook{
		publish: broker.server.Publish,
		log:     log,
	}
	err = broker.server.AddHook(hook, nil)
	if err != nil {
		return nil, fmt.Errorf("add processing hook: %w", err)
	}

	// Create TCP listener
	tcp := listeners.NewTCP(listeners.Config{
		ID:      "tcp",
		Address: addr,
	})
	err = broker.server.AddListener(tcp)
	if err != nil {
		return nil, fmt.Errorf("add listener: %w", err)
	}

	return broker, nil
}

// Start starts the MQTT broker.
func (b *MQTTBroker) Start() error {
	return b.server.Serve() //nolint:wrapcheck
}

// Stop stops the MQTT broker.
func (b *MQTTBroker) Stop() error {
	return b.server.Close() //nolint:wrapcheck
}

// MessageHook handles MQTT message events.
type MessageHook struct {
	mqtt.HookBase
	publish  PublishFunc
	lastSeen time.Time
	mx       sync.Mutex
	log      *logrus.Logger
}

// PublishFunc describes sending message to topics.
type PublishFunc func(topic string, payload []byte, retain bool, qos byte) error

// ID returns the hook ID.
func (h *MessageHook) ID() string {
	return "message-handler"
}

// Provides indicates which hook methods this hook provides.
func (h *MessageHook) Provides(flag byte) bool {
	return flag == mqtt.OnPublish
}

// OnPublish is called when a message is published to the broker.
func (h *MessageHook) OnPublish(_ *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	h.log.WithFields(logrus.Fields{
		"topic":   pk.TopicName,
		"payload": string(pk.Payload),
	}).Debug("Received MQTT message")

	// Parse the message envelope
	var msg QingpingMessage
	if err := json.Unmarshal(pk.Payload, &msg); err != nil {
		h.log.WithError(err).Error("Failed to parse message")
		parseErrorsCounter.Inc()
		return pk, nil
	}
	messagesReceivedCounter.WithLabelValues(msg.Type).Inc()

	if !slices.Contains(AllowedMessageTypes, msg.Type) {
		h.log.WithField("type", msg.Type).Debug("Ignoring message type")
		return pk, nil
	}

	if msg.Type == HeadrbeatType {
		h.heartbeat(msg.MAC, pk.TopicName)
		return pk, nil
	}

	// Take the latest data
	var data SensorData
	var latest float64
	for _, d := range msg.SensorData {
		if d.Timestamp.Value > latest {
			data = d
		}
	}
	h.setMetrics(msg.MAC, pk.TopicName, data)

	if msg.NeedAck == 1 {
		h.sendAcknowledgment(pk.TopicName, msg.ID)
	}

	return pk, nil
}

// sendAcknowledgment sends an acknowledgment message back to the device.
func (h *MessageHook) sendAcknowledgment(upTopic string, msgID int) {
	downTopic := strings.Replace(upTopic, "/up", "/down", 1)
	log := h.log.WithFields(logrus.Fields{
		"msg_id": msgID,
		"topic":  downTopic,
	})

	ack := AckResponse{
		Type:      "18",
		AckID:     msgID,
		Timestamp: time.Now().Unix(),
	}

	payload, err := json.Marshal(ack)
	if err != nil {
		log.WithError(err).Error("Failed to marshal acknowledgment")
		ackErrorsCounter.Inc()
		return
	}

	if err := h.publish(downTopic, payload, false, 0); err != nil {
		log.WithError(err).Error("Failed to publish acknowledgment")
		ackErrorsCounter.Inc()
		return
	}

	acksSentCounter.Inc()
	log.Debug("Sent acknowledgment")
}

func (h *MessageHook) heartbeat(mac, topic string) {
	h.mx.Lock()
	defer h.mx.Unlock()

	h.lastSeen = time.Now()
	lastSeen := h.lastSeen

	// If `lastSeen` has not increased after some time,
	// set all metrics to 0
	go func() {
		time.Sleep(3 * HeartBeatInterval)
		if !h.lastSeen.After(lastSeen) {
			h.setMetrics(mac, topic, SensorData{})
		}
	}()
}

func (h *MessageHook) setMetrics(mac, topic string, data SensorData) {
	h.mx.Lock()
	defer h.mx.Unlock()

	temperatureGauge.WithLabelValues(mac, topic).Set(data.Temperature.Value)
	humidityGauge.WithLabelValues(mac, topic).Set(data.Humidity.Value)
	co2Gauge.WithLabelValues(mac, topic).Set(data.CO2.Value)
	pm1Gauge.WithLabelValues(mac, topic).Set(data.PM1.Value)
	pm25Gauge.WithLabelValues(mac, topic).Set(data.PM25.Value)
	pm10Gauge.WithLabelValues(mac, topic).Set(data.PM10.Value)
	tvocGauge.WithLabelValues(mac, topic).Set(data.TVOC.Value)
	radonGauge.WithLabelValues(mac, topic).Set(data.Radon.Value)
	batteryGauge.WithLabelValues(mac, topic).Set(data.Battery.Value)
}
