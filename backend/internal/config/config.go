// Package config membaca pengaturan dari environment variable (dan berkas .env).
package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"silo/internal/risk"
)

type Config struct {
	Port        string
	DBPath      string
	JWTSecret   string
	JWTHours    int
	AdminEmail  string
	AdminName   string
	AdminPass   string
	CORSOrigins []string
	Timezone    string

	MQTTBroker     string
	MQTTUsername   string
	MQTTPassword   string
	TopicTelemetry string
	TopicScan      string
	TopicControl   string

	Simulator        bool
	SimBackfillHours int
	RetentionDays    int
	DeviceTimeoutSec int

	Thresholds risk.Thresholds
	Vent       risk.VentThresholds
}

// LoadDotEnv membaca berkas KEY=VALUE sederhana. Variabel yang sudah ada
// di environment TIDAK ditimpa.
func LoadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.Trim(strings.TrimSpace(v), `"'`)
		if _, exists := os.LookupEnv(k); !exists {
			os.Setenv(k, v)
		}
	}
}

func str(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func integer(key string, def int) int {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func float(key string, def float64) float64 {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			return n
		}
	}
	return def
}

func boolean(key string, def bool) bool {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func Load() Config {
	LoadDotEnv(".env")

	th := risk.Default()
	th.WarnTemp = float("TH_WARN_TEMP", th.WarnTemp)
	th.WarnHum = float("TH_WARN_HUM", th.WarnHum)
	th.WarnGas = float("TH_WARN_GAS", th.WarnGas)
	th.CritTemp = float("TH_CRIT_TEMP", th.CritTemp)
	th.CritHum = float("TH_CRIT_HUM", th.CritHum)
	th.CritGas = float("TH_CRIT_GAS", th.CritGas)
	th.WarnTempOff = float("TH_WARN_TEMP_OFF", th.WarnTempOff)
	th.WarnHumOff = float("TH_WARN_HUM_OFF", th.WarnHumOff)
	th.WarnGasOff = float("TH_WARN_GAS_OFF", th.WarnGasOff)
	th.CritTempOff = float("TH_CRIT_TEMP_OFF", th.CritTempOff)
	th.CritHumOff = float("TH_CRIT_HUM_OFF", th.CritHumOff)
	th.CritGasOff = float("TH_CRIT_GAS_OFF", th.CritGasOff)

	vt := risk.DefaultVent()
	vt.Temp = float("VENT_TEMP", vt.Temp)
	vt.Hum = float("VENT_HUM", vt.Hum)
	vt.Gas = float("VENT_GAS", vt.Gas)
	vt.TempHyst = float("VENT_TEMP_HYST", vt.TempHyst)
	vt.HumHyst = float("VENT_HUM_HYST", vt.HumHyst)
	vt.GasHyst = float("VENT_GAS_HYST", vt.GasHyst)

	origins := []string{}
	for _, o := range strings.Split(str("CORS_ORIGINS", "http://localhost:8000,http://127.0.0.1:8000"), ",") {
		if o = strings.TrimSpace(o); o != "" {
			origins = append(origins, o)
		}
	}

	return Config{
		Port:        str("PORT", "3000"),
		DBPath:      str("DB_PATH", "data/silo.db"),
		JWTSecret:   str("JWT_SECRET", "ganti-rahasia-ini-sebelum-dipakai"),
		JWTHours:    integer("JWT_HOURS", 8),
		AdminEmail:  str("ADMIN_EMAIL", "admin@silo.local"),
		AdminName:   str("ADMIN_NAME", "Admin Silo"),
		AdminPass:   str("ADMIN_PASSWORD", "admin123"),
		CORSOrigins: origins,
		Timezone:    str("TIMEZONE", "Asia/Jakarta"),

		MQTTBroker:     str("MQTT_BROKER", "tcp://broker.hivemq.com:1883"),
		MQTTUsername:   str("MQTT_USERNAME", ""),
		MQTTPassword:   str("MQTT_PASSWORD", ""),
		TopicTelemetry: str("TOPIC_TELEMETRY", "silo/scm/telemetry/data"),
		TopicScan:      str("TOPIC_SCAN", "silo/scm/scan"),
		TopicControl:   str("TOPIC_CONTROL", "silo/scm/telemetry/control"),

		Simulator:        boolean("SIMULATOR", false),
		SimBackfillHours: integer("SIM_BACKFILL_HOURS", 6),
		RetentionDays:    integer("RETENTION_DAYS", 30),
		DeviceTimeoutSec: integer("DEVICE_TIMEOUT_SEC", 30),

		Thresholds: th,
		Vent:       vt,
	}
}
