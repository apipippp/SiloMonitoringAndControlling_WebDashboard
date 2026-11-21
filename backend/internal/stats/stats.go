// Package stats berisi seluruh perhitungan data: ringkasan statistik,
// tren (regresi linier), estimasi waktu mencapai batas, distribusi status,
// agregasi per interval (bucket), dan hitungan barang per jam.
package stats

import (
	"math"
	"strings"

	"silo/internal/risk"
)

// Point adalah satu pembacaan sensor yang sudah disederhanakan untuk dihitung.
type Point struct {
	TS       int64 // unix detik
	Temp     float64
	Hum      float64
	Gas      float64
	Risk     risk.Level
	Fan      bool
	Buzzer   bool
	Conveyor string
}

// Metric adalah ringkasan statistik satu besaran (suhu / kelembaban / gas).
type Metric struct {
	Min           float64 `json:"min"`
	Max           float64 `json:"maks"`
	RataRata      float64 `json:"rata_rata"`
	SimpanganBaku float64 `json:"simpangan_baku"`
	Terakhir      float64 `json:"terakhir"`
	// LajuPerJam adalah kemiringan regresi linier seluruh jendela waktu (satuan per jam).
	LajuPerJam float64 `json:"laju_per_jam"`
	Tren       string  `json:"tren"` // naik | turun | stabil
	// EstimasiMenitKeWarning dihitung dari tren 30 menit terakhir.
	// nil jika tidak sedang naik atau sudah melewati batas WARNING.
	EstimasiMenitKeWarning *float64 `json:"estimasi_menit_ke_warning"`
}

type RiskStat struct {
	NormalPct   float64 `json:"normal_pct"`
	WarningPct  float64 `json:"warning_pct"`
	CriticalPct float64 `json:"critical_pct"`
	Transisi    int     `json:"transisi"` // berapa kali level berubah
}

type FanStat struct {
	OnPct float64 `json:"on_pct"`
}

type AlarmStat struct {
	OnPct    float64 `json:"on_pct"`
	Kejadian int     `json:"kejadian"` // berapa kali alarm menyala (rising edge)
}

type Summary struct {
	JumlahSampel int                `json:"jumlah_sampel"`
	Suhu         Metric             `json:"suhu"`
	Kelembaban   Metric             `json:"kelembaban"`
	Gas          Metric             `json:"gas"`
	Risiko       RiskStat           `json:"risiko"`
	Kipas        FanStat            `json:"kipas"`
	Alarm        AlarmStat          `json:"alarm"`
	Konveyor     map[string]float64 `json:"konveyor_pct"` // persen sampel per status
}

func r2(v float64) float64 { return math.Round(v*100) / 100 }

func pct(n, total int) float64 {
	if total == 0 {
		return 0
	}
	return r2(float64(n) * 100 / float64(total))
}

// slope menghitung kemiringan regresi linier (satuan nilai per jam)
// dengan metode kuadrat terkecil. Butuh minimal 3 titik.
func slope(vals []float64, ts []int64) float64 {
	n := len(vals)
	if n < 3 {
		return 0
	}
	t0 := ts[0]
	var sx, sy, sxx, sxy float64
	for i := range vals {
		x := float64(ts[i]-t0) / 3600.0
		sx += x
		sy += vals[i]
		sxx += x * x
		sxy += x * vals[i]
	}
	fn := float64(n)
	den := fn*sxx - sx*sx
	if math.Abs(den) < 1e-9 {
		return 0
	}
	return (fn*sxy - sx*sy) / den
}

// metric menghitung ringkasan satu besaran.
// eps = batas laju per jam agar dianggap "naik/turun", warn = batas WARNING,
// recentFrom = indeks awal data "30 menit terakhir" untuk estimasi.
func metric(vals []float64, ts []int64, eps, warn float64, recentFrom int) Metric {
	n := len(vals)
	m := Metric{Min: vals[0], Max: vals[0], Terakhir: vals[n-1]}
	var sum float64
	for _, v := range vals {
		if v < m.Min {
			m.Min = v
		}
		if v > m.Max {
			m.Max = v
		}
		sum += v
	}
	mean := sum / float64(n)
	m.RataRata = mean
	if n > 1 {
		var ss float64
		for _, v := range vals {
			ss += (v - mean) * (v - mean)
		}
		m.SimpanganBaku = math.Sqrt(ss / float64(n-1)) // simpangan baku sampel
	}
	m.LajuPerJam = slope(vals, ts)
	switch {
	case m.LajuPerJam > eps:
		m.Tren = "naik"
	case m.LajuPerJam < -eps:
		m.Tren = "turun"
	default:
		m.Tren = "stabil"
	}

	recent := slope(vals[recentFrom:], ts[recentFrom:])
	if recent > eps/2 && m.Terakhir < warn {
		eta := (warn - m.Terakhir) / recent * 60
		eta = r2(eta)
		m.EstimasiMenitKeWarning = &eta
	}

	m.Min, m.Max, m.RataRata = r2(m.Min), r2(m.Max), r2(m.RataRata)
	m.SimpanganBaku, m.Terakhir, m.LajuPerJam = r2(m.SimpanganBaku), r2(m.Terakhir), r2(m.LajuPerJam)
	return m
}

// Compute menghitung semua ringkasan dari data yang SUDAH terurut naik menurut waktu.
func Compute(points []Point, th risk.Thresholds) Summary {
	s := Summary{JumlahSampel: len(points), Konveyor: map[string]float64{}}
	n := len(points)
	if n == 0 {
		return s
	}

	temps := make([]float64, n)
	hums := make([]float64, n)
	gases := make([]float64, n)
	ts := make([]int64, n)
	for i, p := range points {
		temps[i], hums[i], gases[i], ts[i] = p.Temp, p.Hum, p.Gas, p.TS
	}

	// indeks awal 30 menit terakhir
	recentFrom := 0
	cutoff := ts[n-1] - 1800
	for i := range ts {
		if ts[i] >= cutoff {
			recentFrom = i
			break
		}
	}
	if n-recentFrom < 3 { // terlalu sedikit data untuk tren pendek
		recentFrom = max(0, n-3)
	}

	s.Suhu = metric(temps, ts, 0.5, th.WarnTemp, recentFrom)
	s.Kelembaban = metric(hums, ts, 1.0, th.WarnHum, recentFrom)
	s.Gas = metric(gases, ts, 25, th.WarnGas, recentFrom)

	var nN, nW, nC, fanOn, buzOn, alarmEvents, transitions int
	conv := map[string]int{}
	var prevRisk risk.Level
	var prevBuz bool
	for i, p := range points {
		switch p.Risk {
		case risk.Critical:
			nC++
		case risk.Warning:
			nW++
		default:
			nN++
		}
		if p.Fan {
			fanOn++
		}
		if p.Buzzer {
			buzOn++
			if i > 0 && !prevBuz {
				alarmEvents++
			}
			if i == 0 {
				alarmEvents++
			}
		}
		if i > 0 && p.Risk != prevRisk {
			transitions++
		}
		prevRisk, prevBuz = p.Risk, p.Buzzer
		conv[strings.ToUpper(p.Conveyor)]++
	}
	s.Risiko = RiskStat{NormalPct: pct(nN, n), WarningPct: pct(nW, n), CriticalPct: pct(nC, n), Transisi: transitions}
	s.Kipas = FanStat{OnPct: pct(fanOn, n)}
	s.Alarm = AlarmStat{OnPct: pct(buzOn, n), Kejadian: alarmEvents}
	for k, v := range conv {
		s.Konveyor[k] = pct(v, n)
	}
	return s
}

// Bucket adalah rata-rata data dalam satu interval waktu (untuk grafik riwayat).
type Bucket struct {
	TS         int64      `json:"ts"`
	Suhu       float64    `json:"suhu"`
	Kelembaban float64    `json:"kelembaban"`
	Gas        float64    `json:"gas"`
	Risiko     risk.Level `json:"risiko"` // level TERTINGGI dalam interval
	N          int        `json:"n"`
}

// Bucketize merata-ratakan titik data per interval bucketSec detik.
// Input harus terurut naik menurut waktu.
func Bucketize(points []Point, bucketSec int64) []Bucket {
	if bucketSec < 1 {
		bucketSec = 1
	}
	var out []Bucket
	var sumT, sumH, sumG float64
	var n int
	var worst risk.Level
	var key int64 = -1

	flush := func() {
		if n == 0 {
			return
		}
		fn := float64(n)
		out = append(out, Bucket{TS: key, Suhu: r2(sumT / fn), Kelembaban: r2(sumH / fn), Gas: r2(sumG / fn), Risiko: worst, N: n})
	}
	for _, p := range points {
		k := p.TS / bucketSec * bucketSec
		if k != key {
			flush()
			key, sumT, sumH, sumG, n, worst = k, 0, 0, 0, 0, risk.Normal
		}
		sumT += p.Temp
		sumH += p.Hum
		sumG += p.Gas
		n++
		if p.Risk.Rank() > worst.Rank() {
			worst = p.Risk
		}
	}
	flush()
	return out
}

// ScanTime adalah waktu satu barang masuk (hasil scan sukses).
type ScanTime struct{ TS int64 }

// HourCount adalah jumlah barang masuk dalam satu jam.
type HourCount struct {
	TS     int64 `json:"ts"` // awal jam (unix detik)
	Jumlah int   `json:"jumlah"`
}

// HourlyCounts menghitung barang masuk per jam untuk `hours` jam terakhir
// (jam berjalan termasuk). Hasil selalu berisi tepat `hours` elemen.
func HourlyCounts(scans []ScanTime, hours int, nowUnix int64) []HourCount {
	if hours < 1 {
		hours = 1
	}
	curHour := nowUnix / 3600 * 3600
	start := curHour - int64(hours-1)*3600
	out := make([]HourCount, hours)
	for i := range out {
		out[i].TS = start + int64(i)*3600
	}
	for _, s := range scans {
		if s.TS < start {
			continue
		}
		idx := int((s.TS - start) / 3600)
		if idx >= 0 && idx < hours {
			out[idx].Jumlah++
		}
	}
	return out
}
