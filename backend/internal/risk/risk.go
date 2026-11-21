// Package risk menentukan level kondisi silo (NORMAL / WARNING / CRITICAL)
// dari suhu, kelembaban, dan gas, lengkap dengan histeresis.
package risk

import "strings"

type Level string

const (
	Normal   Level = "NORMAL"
	Warning  Level = "WARNING"
	Critical Level = "CRITICAL"
)

// Parse mengubah teks menjadi Level. ok=false jika teks bukan level yang dikenal.
func Parse(s string) (Level, bool) {
	switch Level(strings.ToUpper(strings.TrimSpace(s))) {
	case Normal:
		return Normal, true
	case Warning:
		return Warning, true
	case Critical:
		return Critical, true
	}
	return Normal, false
}

// Rank memberi urutan keparahan: NORMAL=0, WARNING=1, CRITICAL=2.
func (l Level) Rank() int {
	switch l {
	case Warning:
		return 1
	case Critical:
		return 2
	}
	return 0
}

// Thresholds menyimpan batas naik (…) dan batas pulih/histeresis (…Off).
// Nilai awal diambil dari dokumen konsultasi Layer 1 dan WAJIB divalidasi
// ulang untuk sensor DHT11/MQ-2 dan jenis raw material yang dipakai.
type Thresholds struct {
	WarnTemp float64 `json:"warn_temp"`
	WarnHum  float64 `json:"warn_hum"`
	WarnGas  float64 `json:"warn_gas"`
	CritTemp float64 `json:"crit_temp"`
	CritHum  float64 `json:"crit_hum"`
	CritGas  float64 `json:"crit_gas"`

	WarnTempOff float64 `json:"warn_temp_off"`
	WarnHumOff  float64 `json:"warn_hum_off"`
	WarnGasOff  float64 `json:"warn_gas_off"`
	CritTempOff float64 `json:"crit_temp_off"`
	CritHumOff  float64 `json:"crit_hum_off"`
	CritGasOff  float64 `json:"crit_gas_off"`
}

// Default mengembalikan threshold simulasi awal kelompok.
func Default() Thresholds {
	return Thresholds{
		WarnTemp: 35, WarnHum: 65, WarnGas: 2175,
		CritTemp: 40, CritHum: 75, CritGas: 2500,
		WarnTempOff: 33, WarnHumOff: 62, WarnGasOff: 2050,
		CritTempOff: 38, CritHumOff: 72, CritGasOff: 2375,
	}
}

// Evaluate menghitung level baru berdasarkan level sebelumnya (prev).
// Naik level: nilai harus MELEWATI batas naik.
// Turun level: semua nilai harus berada di bawah batas pulih (histeresis),
// sehingga status tidak berkedip saat nilai berada di sekitar batas.
func (t Thresholds) Evaluate(prev Level, temp, hum, gas float64) Level {
	level := Normal
	switch {
	case temp > t.CritTemp || hum > t.CritHum || gas > t.CritGas:
		level = Critical
	case temp > t.WarnTemp || hum > t.WarnHum || gas > t.WarnGas:
		level = Warning
	}

	if prev == Critical && level.Rank() < Critical.Rank() {
		if !(temp < t.CritTempOff && hum < t.CritHumOff && gas < t.CritGasOff) {
			level = Critical
		}
	}
	if prev.Rank() >= Warning.Rank() && level.Rank() < Warning.Rank() {
		if !(temp < t.WarnTempOff && hum < t.WarnHumOff && gas < t.WarnGasOff) {
			level = Warning
		}
	}
	return level
}
