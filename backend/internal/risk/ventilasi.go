package risk

// VentThresholds mengatur kapan ventilasi (flap) dibuka otomatis.
//
// Buka: salah satu dari suhu, kelembapan, atau gas MELEWATI batas bukanya.
// Tutup: semua nilai sudah turun di bawah (batas - histeresis), sehingga flap
// tidak bolak-balik saat nilai berada di sekitar batas.
type VentThresholds struct {
	Temp float64 `json:"temp"`
	Hum  float64 `json:"hum"`
	Gas  float64 `json:"gas"`

	TempHyst float64 `json:"temp_hyst"`
	HumHyst  float64 `json:"hum_hyst"`
	GasHyst  float64 `json:"gas_hyst"`
}

// DefaultVent memakai suhu 33°C sebagai batas buka, dengan kelembapan dan gas
// mengikuti nilai pulih pada Default(). Histeresis dipilih kecil agar flap
// tidak terlalu sering bergerak; semuanya bisa diubah lewat .env.
func DefaultVent() VentThresholds {
	return VentThresholds{
		Temp: 33, Hum: 62, Gas: 2050,
		TempHyst: 2, HumHyst: 3, GasHyst: 100,
	}
}

// Evaluate mengembalikan apakah ventilasi harus terbuka, berdasarkan keadaan sebelumnya.
func (v VentThresholds) Evaluate(prevOpen bool, temp, hum, gas float64) bool {
	if temp > v.Temp || hum > v.Hum || gas > v.Gas {
		return true
	}
	if !prevOpen {
		return false
	}
	// Masih terbuka selama ada nilai yang belum turun cukup jauh dari batas.
	return temp >= v.Temp-v.TempHyst || hum >= v.Hum-v.HumHyst || gas >= v.Gas-v.GasHyst
}
