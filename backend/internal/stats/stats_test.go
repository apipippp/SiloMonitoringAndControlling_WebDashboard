package stats

import (
	"math"
	"testing"

	"silo/internal/risk"
)

func near(a, b, tol float64) bool { return math.Abs(a-b) <= tol }

func TestComputeBasics(t *testing.T) {
	var pts []Point
	// 2 jam data tiap 60 detik: suhu naik 2 °C/jam dari 30
	for i := 0; i <= 120; i++ {
		ts := int64(1_000_000 + i*60)
		temp := 30 + 2*float64(i)/60.0
		lvl := risk.Normal
		if i >= 100 {
			lvl = risk.Warning
		}
		pts = append(pts, Point{TS: ts, Temp: temp, Hum: 60, Gas: 1000, Risk: lvl, Fan: i%2 == 0, Conveyor: "running"})
	}
	s := Compute(pts, risk.Default())

	if s.JumlahSampel != 121 {
		t.Fatalf("jumlah sampel = %d", s.JumlahSampel)
	}
	if !near(s.Suhu.LajuPerJam, 2, 0.01) {
		t.Errorf("laju suhu = %v, want 2", s.Suhu.LajuPerJam)
	}
	if s.Suhu.Tren != "naik" {
		t.Errorf("tren = %s", s.Suhu.Tren)
	}
	if s.Suhu.Min != 30 || !near(s.Suhu.Max, 34, 0.01) {
		t.Errorf("min/max = %v/%v", s.Suhu.Min, s.Suhu.Max)
	}
	if !near(s.Suhu.RataRata, 32, 0.01) {
		t.Errorf("rata-rata = %v", s.Suhu.RataRata)
	}
	// terakhir 34 °C, batas WARNING 35, laju 2/jam -> 0.5 jam = 30 menit
	if s.Suhu.EstimasiMenitKeWarning == nil || !near(*s.Suhu.EstimasiMenitKeWarning, 30, 0.5) {
		t.Errorf("estimasi = %v", s.Suhu.EstimasiMenitKeWarning)
	}
	if s.Kelembaban.Tren != "stabil" || s.Kelembaban.EstimasiMenitKeWarning != nil {
		t.Errorf("kelembaban seharusnya stabil tanpa estimasi: %+v", s.Kelembaban)
	}
	if !near(s.Risiko.WarningPct, pct(21, 121), 0.01) || s.Risiko.Transisi != 1 {
		t.Errorf("risiko = %+v", s.Risiko)
	}
	if !near(s.Kipas.OnPct, pct(61, 121), 0.01) {
		t.Errorf("kipas = %+v", s.Kipas)
	}
	if s.Konveyor["RUNNING"] != 100 {
		t.Errorf("konveyor = %v", s.Konveyor)
	}
}

func TestComputeEmpty(t *testing.T) {
	s := Compute(nil, risk.Default())
	if s.JumlahSampel != 0 {
		t.Fatal("harus kosong")
	}
}

func TestBucketize(t *testing.T) {
	pts := []Point{
		{TS: 0, Temp: 30, Risk: risk.Normal},
		{TS: 30, Temp: 32, Risk: risk.Warning},
		{TS: 60, Temp: 40, Risk: risk.Normal},
		{TS: 90, Temp: 42, Risk: risk.Critical},
	}
	b := Bucketize(pts, 60)
	if len(b) != 2 {
		t.Fatalf("len = %d", len(b))
	}
	if b[0].Suhu != 31 || b[0].Risiko != risk.Warning || b[0].N != 2 {
		t.Errorf("bucket 0 = %+v", b[0])
	}
	if b[1].Suhu != 41 || b[1].Risiko != risk.Critical {
		t.Errorf("bucket 1 = %+v", b[1])
	}
}

func TestHourlyCounts(t *testing.T) {
	now := int64(10*3600 + 1234)
	scans := []ScanTime{{TS: 10*3600 + 5}, {TS: 10*3600 + 900}, {TS: 9*3600 + 10}, {TS: 3600}}
	h := HourlyCounts(scans, 3, now)
	if len(h) != 3 {
		t.Fatalf("len = %d", len(h))
	}
	if h[2].Jumlah != 2 || h[1].Jumlah != 1 || h[0].Jumlah != 0 {
		t.Errorf("hasil = %+v", h)
	}
}
