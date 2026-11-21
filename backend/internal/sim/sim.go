// Package sim membangkitkan data palsu yang realistis (suhu, kelembaban, gas,
// konveyor, barang masuk) supaya dashboard bisa diuji TANPA perangkat keras
// dan TANPA broker MQTT. Simulator juga menerima perintah kontrol dari dashboard.
package sim

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"silo/internal/model"
	"silo/internal/risk"
	"silo/internal/service"
)

type Sim struct {
	svc *service.Service

	mu         sync.Mutex
	rng        *rand.Rand
	temp, hum  float64
	gas        float64
	fan        bool
	vent       bool
	mode       string
	conveyorOn bool
	phase      int
	lvl        risk.Level
}

func New(svc *service.Service) *Sim {
	return &Sim{
		svc: svc, rng: rand.New(rand.NewSource(time.Now().UnixNano())),
		temp: 30, hum: 58, gas: 1200, mode: "AUTO", conveyorOn: true, lvl: risk.Normal,
	}
}

func ptr(v float64) *float64 { return &v }

// profile adalah "kondisi lingkungan" tanpa pengaruh kipas, berupa gabungan gelombang sinus
// sehingga suhu/kelembaban/gas naik-turun dan sesekali melewati batas WARNING/CRITICAL.
func profile(t time.Time) (temp, hum, gas float64) {
	s := float64(t.Unix())
	temp = 32 + 6*math.Sin(2*math.Pi*s/900) + 3*math.Sin(2*math.Pi*s/5400+1)
	hum = 62 + 10*math.Sin(2*math.Pi*s/1100+2) + 4*math.Sin(2*math.Pi*s/3600)
	gas = 1500 + 700*math.Sin(2*math.Pi*s/1300+1)
	if math.Sin(2*math.Pi*s/2700) > 0.9 {
		gas += 500 // lonjakan gas sesekali
	}
	return
}

func fanEffect(on bool) (dt, dh, dg float64) {
	if on {
		return -3, -6, -250
	}
	return 0, 0, 0
}

// ventEffect: flap terbuka mengalirkan udara luar, lebih lemah dari kipas.
func ventEffect(open bool) (dt, dh, dg float64) {
	if open {
		return -1.5, -3, -120
	}
	return 0, 0, 0
}

func batchID(t time.Time) string { return "BATCH-" + t.Format("060102-150405") }

func material(n int64) string {
	list := []string{"Biji Kopi Arabika", "Biji Kopi Robusta", "Gandum"}
	return list[int(n)%len(list)]
}

// Backfill mengisi riwayat `hours` jam ke belakang (hanya jika database masih kosong).
func (s *Sim) Backfill(hours int) error {
	if hours <= 0 {
		return nil
	}
	if last, err := s.svc.Store.LatestReading(); err != nil || last != nil {
		return err
	}
	th := s.svc.Cfg.Thresholds
	now := time.Now()
	step := 30 * time.Second
	n := int(time.Duration(hours) * time.Hour / step)
	lvl, fan, vent := risk.Normal, false, false

	for i := n; i >= 1; i-- {
		at := now.Add(-time.Duration(i) * step)
		t, h, g := profile(at)
		dt, dh, dg := fanEffect(fan)
		vt, vh, vg := ventEffect(vent)
		t += dt + vt + s.rng.NormFloat64()*0.15
		h += dh + vh + s.rng.NormFloat64()*0.4
		g += dg + vg + s.rng.NormFloat64()*20
		h = math.Max(20, math.Min(99, h))
		g = math.Max(0, g)

		lvl = th.Evaluate(lvl, t, h, g)
		fan = lvl != risk.Normal
		vent = s.svc.Cfg.Vent.Evaluate(vent, t, h, g)

		conv := "RUNNING"
		if i%10 == 0 {
			conv = "SCANNING"
		}
		err := s.svc.IngestTelemetry(model.Telemetry{
			DeviceID: "silo-01", Temperature: ptr(t), Humidity: ptr(h), GasLevel: ptr(g),
			RiskLevel: string(lvl), ExhaustFan: model.OnOff(fan), BuzzerAlarm: model.OnOff(lvl == risk.Critical),
			ControlMode: "AUTO", ConveyorStatus: conv, DistanceCM: 20, VentOpen: model.OnOff(vent),
		}, at)
		if err != nil {
			return err
		}
		if i%10 == 0 {
			if _, err := s.ingestBarcode(at); err != nil {
				return err
			}
		}
	}
	return nil
}

// Run menghasilkan satu pembacaan setiap 2 detik sampai ctx dibatalkan.
func (s *Sim) Run(ctx context.Context) {
	tk := time.NewTicker(2 * time.Second)
	defer tk.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-tk.C:
			s.tick(now)
		}
	}
}

func (s *Sim) tick(now time.Time) {
	s.mu.Lock()
	th := s.svc.Cfg.Thresholds

	t, h, g := profile(now)
	dt, dh, dg := fanEffect(s.fan)
	vt, vh, vg := ventEffect(s.vent)
	// bergerak perlahan menuju target (seperti sensor sungguhan) + sedikit derau
	s.temp += 0.3*((t+dt+vt)-s.temp) + s.rng.NormFloat64()*0.08
	s.hum += 0.3*((h+dh+vh)-s.hum) + s.rng.NormFloat64()*0.2
	s.gas += 0.3*((g+dg+vg)-s.gas) + s.rng.NormFloat64()*10
	s.hum = math.Max(20, math.Min(99, s.hum))
	s.gas = math.Max(0, s.gas)

	s.lvl = th.Evaluate(s.lvl, s.temp, s.hum, s.gas)
	if s.mode == "AUTO" {
		s.fan = s.lvl != risk.Normal
		s.vent = s.svc.Cfg.Vent.Evaluate(s.vent, s.temp, s.hum, s.gas)
	}

	// siklus konveyor: 8 tick bergerak, 2 tick scanning, lalu barang masuk
	conv, dist, scan := "STOPPED", 40.0, false
	if s.conveyorOn {
		s.phase = (s.phase + 1) % 10
		if s.phase < 8 {
			conv, dist = "RUNNING", 40-float64(s.phase)*5
		} else {
			conv, dist = "SCANNING", 5
			scan = s.phase == 9
		}
	}

	tel := model.Telemetry{
		DeviceID: "silo-01", Temperature: ptr(s.temp), Humidity: ptr(s.hum), GasLevel: ptr(s.gas),
		RiskLevel: string(s.lvl), ExhaustFan: model.OnOff(s.fan), BuzzerAlarm: model.OnOff(s.lvl == risk.Critical),
		ControlMode: s.mode, ConveyorStatus: conv, DistanceCM: dist, VentOpen: model.OnOff(s.vent),
	}
	s.mu.Unlock()

	if err := s.svc.IngestTelemetry(tel, now); err != nil {
		fmt.Println("simulator: gagal simpan telemetri:", err)
	}
	if scan {
		_, err := s.ingestBarcode(now)
		if err != nil {
			fmt.Println("simulator: gagal simpan scan:", err)
		}
	}
}

// Publish dipanggil handler saat dashboard mengirim perintah; di mode simulator
// perintah langsung diterapkan ke keadaan simulasi.
func (s *Sim) Publish(payload []byte) error {
	var cmd map[string]string
	if err := json.Unmarshal(payload, &cmd); err != nil {
		return errors.New("format perintah tidak valid")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if m, ok := cmd["mode"]; ok {
		s.mode = strings.ToUpper(m)
	}
	if f, ok := cmd["fan"]; ok {
		if _, hasMode := cmd["mode"]; !hasMode {
			s.mode = "MANUAL"
		}
		s.fan = strings.EqualFold(f, "ON")
	}
	if v, ok := cmd["vent"]; ok {
		if _, hasMode := cmd["mode"]; !hasMode {
			s.mode = "MANUAL"
		}
		s.vent = strings.EqualFold(v, "OPEN")
	}
	if c, ok := cmd["conveyor"]; ok {
		s.conveyorOn = strings.EqualFold(c, "START")
	}
	return nil
}

var (
	simVar = []string{"Arabika-S795", "Robusta-R12", "Arabika-Gayo1"}
	simPrc = []string{"RedHoney", "Washed", "NaturalSun"}
)

// ingestBarcode meniru barcode batch dari perangkat (format JSON lengkap) untuk simulator.
func (s *Sim) ingestBarcode(at time.Time) (bool, error) {
	k := at.Unix()
	raw, err := json.Marshal(map[string]any{
		"id":   fmt.Sprintf("BOX-%s-%03d", at.Format("0601"), k%1000),
		"lot":  fmt.Sprintf("LOT-TRJ-B%02d", (k/300)%20),
		"frm":  fmt.Sprintf("FRM-%03d", (k/60)%200),
		"geo":  fmt.Sprintf("-2.%04d,119.%04d|%dm", k%10000, (k/7)%10000, 1500+k%300),
		"var":  simVar[(k/300)%int64(len(simVar))],
		"prc":  simPrc[(k/600)%int64(len(simPrc))],
		"dt":   at.Format("2006-01-02"),
		"mc":   11.0 + float64(k%15)/10,
		"grd":  "G1-Specialty",
		"wt":   40.0 + float64(k%20),
		"silo": "SILO-A2",
	})
	if err != nil {
		return false, err
	}
	var p model.ScanPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return false, err
	}
	p.Raw = string(raw)
	return s.svc.IngestScan(p, at)
}
