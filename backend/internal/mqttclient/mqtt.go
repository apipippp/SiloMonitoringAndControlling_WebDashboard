// Package mqttclient menghubungkan backend ke broker MQTT (mis. HiveMQ):
// subscribe telemetri & scan dari perangkat, publish perintah kontrol dari dashboard.
package mqttclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"silo/internal/config"
	"silo/internal/model"
	"silo/internal/service"
)

type Client struct {
	cfg config.Config
	svc *service.Service
	c   mqtt.Client
}

func New(cfg config.Config, svc *service.Service) *Client {
	return &Client{cfg: cfg, svc: svc}
}

// Connect memulai koneksi. Koneksi dicoba ulang otomatis di latar belakang,
// jadi backend tetap jalan walau internet/broker sedang bermasalah.
func (m *Client) Connect() {
	clientID := fmt.Sprintf("silo-backend-%d", time.Now().UnixNano()%1_000_000)
	opts := mqtt.NewClientOptions().
		AddBroker(m.cfg.MQTTBroker).
		SetClientID(clientID).
		SetCleanSession(true).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(5 * time.Second).
		SetKeepAlive(30 * time.Second)
	if m.cfg.MQTTUsername != "" {
		opts.SetUsername(m.cfg.MQTTUsername).SetPassword(m.cfg.MQTTPassword)
	}

	opts.SetOnConnectHandler(func(cl mqtt.Client) {
		log.Printf("MQTT terhubung ke %s", m.cfg.MQTTBroker)
		// Subscribe dilakukan di sini agar otomatis diulang setiap reconnect.
		if t := cl.Subscribe(m.cfg.TopicTelemetry, 1, m.onTelemetry); t.Wait() && t.Error() != nil {
			log.Printf("MQTT gagal subscribe %s: %v", m.cfg.TopicTelemetry, t.Error())
		}
		if t := cl.Subscribe(m.cfg.TopicScan, 1, m.onScan); t.Wait() && t.Error() != nil {
			log.Printf("MQTT gagal subscribe %s: %v", m.cfg.TopicScan, t.Error())
		}
	})
	opts.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
		log.Printf("MQTT terputus: %v (mencoba tersambung lagi)", err)
	})

	m.c = mqtt.NewClient(opts)
	m.c.Connect() // tidak ditunggu: retry berjalan di latar belakang
}

func (m *Client) onTelemetry(_ mqtt.Client, msg mqtt.Message) {
	var t model.Telemetry
	if err := json.Unmarshal(msg.Payload(), &t); err != nil {
		log.Printf("telemetri ditolak (JSON tidak valid): %v", err)
		return
	}
	if err := m.svc.IngestTelemetry(t, time.Now()); err != nil {
		log.Printf("telemetri ditolak: %v", err)
	}
}

func (m *Client) onScan(_ mqtt.Client, msg mqtt.Message) {
	var p model.ScanPayload
	if err := json.Unmarshal(msg.Payload(), &p); err != nil {
		log.Printf("scan ditolak (JSON tidak valid): %v", err)
		return
	}
	p.Raw = string(msg.Payload())
	inserted, err := m.svc.IngestScan(p, time.Now())
	if err != nil {
		log.Printf("scan ditolak: %v", err)
		return
	}
	if !inserted {
		log.Printf("scan diabaikan (gagal atau batch %q sudah tercatat)", p.BatchID)
	}
}

// Publish mengirim perintah kontrol (JSON) ke topik kontrol.
func (m *Client) Publish(payload []byte) error {
	if m.c == nil || !m.c.IsConnectionOpen() {
		return errors.New("broker MQTT belum terhubung")
	}
	t := m.c.Publish(m.cfg.TopicControl, 1, false, payload)
	if !t.WaitTimeout(5 * time.Second) {
		return errors.New("waktu habis saat mengirim ke broker")
	}
	return t.Error()
}
