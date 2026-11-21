// Package service berisi alur pengolahan data masuk (dari MQTT atau simulator):
// validasi -> penentuan level risiko -> simpan -> catat kejadian.
package service

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
	"unicode"

	"silo/internal/config"
	"silo/internal/model"
	"silo/internal/risk"
	"silo/internal/store"
)

type deviceState struct {
	risk     risk.Level
	fan      bool
	vent     bool
	conveyor string
}

type Service struct {
	Store *store.Store
	Cfg   config.Config

	mu    sync.Mutex
	state map[string]*deviceState
}

func New(st *store.Store, cfg config.Config) *Service {
	return &Service{Store: st, Cfg: cfg, state: map[string]*deviceState{}}
}

// clean membersihkan teks dari perangkat (tidak dipercaya): buang karakter kontrol, batasi panjang.
func clean(s string, max int, def string) string {
	s = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, strings.TrimSpace(s))
	if r := []rune(s); len(r) > max {
		s = string(r[:max])
	}
	if s == "" {
		return def
	}
	return s
}

func validNumber(v float64, lo, hi float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0) && v >= lo && v <= hi
}

// IngestTelemetry memvalidasi dan menyimpan satu pembacaan. `at` adalah waktu pembacaan.
func (s *Service) IngestTelemetry(t model.Telemetry, at time.Time) error {
	if t.Temperature == nil || t.Humidity == nil || t.GasLevel == nil {
		return errors.New("temperature, humidity, dan gas_level wajib ada")
	}
	temp, hum, gas := *t.Temperature, *t.Humidity, *t.GasLevel
	if !validNumber(temp, -40, 125) || !validNumber(hum, 0, 100) || !validNumber(gas, 0, 100000) {
		return fmt.Errorf("nilai sensor di luar rentang: suhu=%v kelembaban=%v gas=%v", temp, hum, gas)
	}
	dist := t.DistanceCM
	if !validNumber(dist, 0, 1000) {
		dist = 0
	}

	dev := clean(t.DeviceID, 32, "silo-01")
	mode := strings.ToUpper(clean(t.ControlMode, 10, "AUTO"))
	conv := strings.ToUpper(clean(t.ConveyorStatus, 12, "UNKNOWN"))

	s.mu.Lock()
	defer s.mu.Unlock()

	st, known := s.state[dev]
	if !known {
		st = &deviceState{risk: risk.Normal, conveyor: conv, fan: bool(t.ExhaustFan), vent: bool(t.VentOpen)}
		s.state[dev] = st
	}

	// Level dari perangkat dipakai jika valid (perangkat yang menggerakkan kipas);
	// jika tidak ada, server menghitung sendiri dengan histeresis.
	level, ok := risk.Parse(t.RiskLevel)
	if !ok {
		level = s.Cfg.Thresholds.Evaluate(st.risk, temp, hum, gas)
	}

	rd := model.Reading{
		TS: at.Unix(), DeviceID: dev, Suhu: temp, Kelembaban: hum, Gas: gas,
		Risiko: string(level), Kipas: bool(t.ExhaustFan), Buzzer: bool(t.BuzzerAlarm),
		Mode: mode, Konveyor: conv, JarakCM: dist, Ventilasi: bool(t.VentOpen),
	}
	if err := s.Store.InsertReading(rd); err != nil {
		return err
	}

	if known {
		if level != st.risk {
			msg := fmt.Sprintf("Kondisi berubah %s → %s (suhu %.1f°C, kelembaban %.0f%%, gas %.0f)", st.risk, level, temp, hum, gas)
			_ = s.Store.InsertEvent(model.Event{TS: rd.TS, DeviceID: dev, Jenis: "risiko", Level: string(level), Pesan: msg})
		}
		if rd.Kipas != st.fan {
			state := "OFF"
			if rd.Kipas {
				state = "ON"
			}
			_ = s.Store.InsertEvent(model.Event{TS: rd.TS, DeviceID: dev, Jenis: "kipas", Level: "INFO", Pesan: "Kipas ventilasi " + state + " (mode " + mode + ")"})
		}
		if conv != st.conveyor {
			_ = s.Store.InsertEvent(model.Event{TS: rd.TS, DeviceID: dev, Jenis: "konveyor", Level: "INFO", Pesan: "Konveyor " + st.conveyor + " → " + conv})
		}
		if rd.Ventilasi != st.vent {
			state := "TUTUP"
			if rd.Ventilasi {
				state = "BUKA"
			}
			_ = s.Store.InsertEvent(model.Event{TS: rd.TS, DeviceID: dev, Jenis: "ventilasi", Level: "INFO", Pesan: "Ventilasi " + state + " (mode " + mode + ")"})
		}
	}
	st.risk, st.fan, st.vent, st.conveyor = level, rd.Kipas, rd.Ventilasi, conv
	return nil
}

// IngestScan menyimpan satu barang masuk. Mendukung format lama (batch_id) dan format barcode (id + data lengkap).
// inserted=false untuk scan gagal atau batch yang sudah tercatat.
func (s *Service) IngestScan(p model.ScanPayload, at time.Time) (bool, error) {
	if p.OK != nil && !*p.OK {
		return false, nil
	}
	batch := clean(p.BatchID, 64, "")
	if batch == "" {
		batch = clean(p.ID, 64, "")
	}
	if batch == "" {
		return false, errors.New("batch_id atau id wajib ada")
	}
	dev := clean(p.DeviceID, 32, "silo-01")
	material := clean(p.Material, 64, "")
	if material == "" {
		material = clean(p.Var, 64, "-")
	}

	inserted, err := s.Store.InsertScan(model.Scan{TS: at.Unix(), DeviceID: dev, BatchID: batch, Material: material, Payload: p.Raw})
	if err != nil {
		return false, err
	}
	if inserted {
		pesan := fmt.Sprintf("Barang masuk: %s (%s)", batch, material)
		if p.Silo != "" {
			pesan += " → " + clean(p.Silo, 32, "")
		}
		_ = s.Store.InsertEvent(model.Event{TS: at.Unix(), DeviceID: dev, Jenis: "barang", Level: "INFO", Pesan: pesan})
	}
	return inserted, nil
}
