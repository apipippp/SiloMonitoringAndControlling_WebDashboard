package risk

import "testing"

func TestVentBukaSaatSalahSatuLewatBatas(t *testing.T) {
	v := DefaultVent()
	cases := []struct {
		name           string
		temp, hum, gas float64
		want           bool
	}{
		{"semua normal", 28, 55, 1200, false},
		{"suhu tepat 33 belum buka", 33, 55, 1200, false},
		{"suhu lewat 33", 33.5, 55, 1200, true},
		{"kelembapan lewat 62", 28, 63, 1200, true},
		{"gas lewat 2050", 28, 55, 2100, true},
	}
	for _, c := range cases {
		if got := v.Evaluate(false, c.temp, c.hum, c.gas); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestVentTutupSetelahTurunDenganHisteresis(t *testing.T) {
	v := DefaultVent()
	// Suhu turun ke 32 (masih dalam pita histeresis 31–33): tetap terbuka.
	if !v.Evaluate(true, 32, 55, 1200) {
		t.Error("harus tetap terbuka saat suhu 32 (belum di bawah 31)")
	}
	// Suhu turun ke 30 (di bawah 31) dan nilai lain normal: tutup.
	if v.Evaluate(true, 30, 55, 1200) {
		t.Error("harus tertutup saat semua nilai sudah di bawah batas - histeresis")
	}
	// Kelembapan masih tinggi walau suhu normal: tetap terbuka.
	if !v.Evaluate(true, 28, 60, 1200) {
		t.Error("harus tetap terbuka saat kelembapan 60 (masih dalam pita 59–62)")
	}
}
