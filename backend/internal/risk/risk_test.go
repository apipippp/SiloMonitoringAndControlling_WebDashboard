package risk

import "testing"

func TestEvaluate(t *testing.T) {
	th := Default()
	cases := []struct {
		name           string
		prev           Level
		temp, hum, gas float64
		want           Level
	}{
		{"normal", Normal, 30, 55, 1500, Normal},
		{"naik ke warning karena suhu", Normal, 36, 55, 1500, Warning},
		{"naik ke warning karena gas", Normal, 30, 55, 2200, Warning},
		{"naik ke critical karena kelembaban", Normal, 30, 80, 1500, Critical},
		{"tepat di batas tidak naik", Normal, 35, 65, 2175, Normal},
		{"histeresis warning: masih di atas batas pulih", Warning, 34, 55, 1500, Warning},
		{"warning pulih ke normal", Warning, 32, 55, 1500, Normal},
		{"critical turun ke warning", Critical, 36, 55, 1500, Warning},
		{"critical bertahan di atas batas pulih", Critical, 39, 55, 1500, Critical},
		{"critical langsung normal jika semua rendah", Critical, 30, 55, 1500, Normal},
	}
	for _, c := range cases {
		if got := th.Evaluate(c.prev, c.temp, c.hum, c.gas); got != c.want {
			t.Errorf("%s: got %s, want %s", c.name, got, c.want)
		}
	}
}

func TestParse(t *testing.T) {
	if l, ok := Parse(" warning "); !ok || l != Warning {
		t.Fatalf("Parse gagal: %v %v", l, ok)
	}
	if _, ok := Parse("apa"); ok {
		t.Fatal("seharusnya tidak valid")
	}
}
