// Package store menyimpan dan mengambil data dari SQLite.
// Driver: modernc.org/sqlite (Go murni, tidak butuh CGO/compiler C).
package store

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"

	"silo/internal/model"
)

type Store struct{ db *sql.DB }

var schema = []string{
	`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE,
		nama TEXT NOT NULL,
		password_hash TEXT NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS readings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ts INTEGER NOT NULL,
		device_id TEXT NOT NULL,
		temperature REAL NOT NULL,
		humidity REAL NOT NULL,
		gas_level REAL NOT NULL,
		risk TEXT NOT NULL,
		fan INTEGER NOT NULL,
		buzzer INTEGER NOT NULL,
		mode TEXT NOT NULL,
		conveyor TEXT NOT NULL,
		distance_cm REAL NOT NULL DEFAULT 0,
		vent INTEGER NOT NULL DEFAULT 0
	)`,
	`CREATE INDEX IF NOT EXISTS idx_readings_ts ON readings(ts)`,
	`CREATE TABLE IF NOT EXISTS scans (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ts INTEGER NOT NULL,
		device_id TEXT NOT NULL,
		batch_id TEXT NOT NULL UNIQUE,
		material TEXT NOT NULL DEFAULT ''
	)`,
	`CREATE INDEX IF NOT EXISTS idx_scans_ts ON scans(ts)`,
	`CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ts INTEGER NOT NULL,
		device_id TEXT NOT NULL,
		jenis TEXT NOT NULL,
		level TEXT NOT NULL,
		pesan TEXT NOT NULL
	)`,
	`CREATE INDEX IF NOT EXISTS idx_events_ts ON events(ts)`,
}

func Open(path string) (*Store, error) {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	dsn := "file:" + path + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// SQLite hanya mengizinkan satu penulis; satu koneksi menghindari "database is locked".
	db.SetMaxOpenConns(1)
	for _, q := range schema {
		if _, err := db.Exec(q); err != nil {
			db.Close()
			return nil, err
		}
	}
	// Migrasi: database lama belum punya kolom payload/snapshot/vent.
	if err := ensureColumn(db, "scans", "payload", "TEXT NOT NULL DEFAULT ''"); err != nil {
		db.Close()
		return nil, err
	}
	if err := ensureColumn(db, "scans", "snapshot", "TEXT NOT NULL DEFAULT ''"); err != nil {
		db.Close()
		return nil, err
	}
	if err := ensureColumn(db, "readings", "vent", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

// ensureColumn menambahkan kolom jika belum ada. Dipanggil saat start, sebelum ada penulis lain.
func ensureColumn(db *sql.DB, table, col, def string) error {
	rows, err := db.Query(`SELECT name FROM pragma_table_info(?)`, table)
	if err != nil {
		return err
	}
	found := false
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			return err
		}
		if name == col {
			found = true
		}
	}
	rows.Close() // harus ditutup sebelum Exec: koneksi hanya satu
	if found {
		return nil
	}
	_, err = db.Exec("ALTER TABLE " + table + " ADD COLUMN " + col + " " + def)
	return err
}

func (s *Store) Close() error { return s.db.Close() }

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ---------- users ----------

func (s *Store) UpsertUser(email, nama, hash string) error {
	_, err := s.db.Exec(`INSERT INTO users(email, nama, password_hash) VALUES(?,?,?)
		ON CONFLICT(email) DO UPDATE SET nama=excluded.nama, password_hash=excluded.password_hash`, email, nama, hash)
	return err
}

func (s *Store) UserByEmail(email string) (*model.User, error) {
	var u model.User
	err := s.db.QueryRow(`SELECT id, email, nama, password_hash FROM users WHERE email = ?`, email).
		Scan(&u.ID, &u.Email, &u.Nama, &u.PasswordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// ---------- readings ----------

const readingCols = `id, ts, device_id, temperature, humidity, gas_level, risk, fan, buzzer, mode, conveyor, distance_cm, vent`

type scanner interface{ Scan(dest ...any) error }

func scanReading(sc scanner) (model.Reading, error) {
	var r model.Reading
	var fan, buz, vent int
	err := sc.Scan(&r.ID, &r.TS, &r.DeviceID, &r.Suhu, &r.Kelembaban, &r.Gas, &r.Risiko, &fan, &buz, &r.Mode, &r.Konveyor, &r.JarakCM, &vent)
	r.Kipas, r.Buzzer, r.Ventilasi = fan == 1, buz == 1, vent == 1
	return r, err
}

func (s *Store) InsertReading(r model.Reading) error {
	_, err := s.db.Exec(`INSERT INTO readings(ts, device_id, temperature, humidity, gas_level, risk, fan, buzzer, mode, conveyor, distance_cm, vent)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
		r.TS, r.DeviceID, r.Suhu, r.Kelembaban, r.Gas, r.Risiko, b2i(r.Kipas), b2i(r.Buzzer), r.Mode, r.Konveyor, r.JarakCM, b2i(r.Ventilasi))
	return err
}

// LatestReading mengembalikan pembacaan terbaru, atau nil jika tabel kosong.
func (s *Store) LatestReading() (*model.Reading, error) {
	row := s.db.QueryRow(`SELECT ` + readingCols + ` FROM readings ORDER BY ts DESC, id DESC LIMIT 1`)
	r, err := scanReading(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// ReadingsSince mengembalikan pembacaan sejak ts (unix detik), urut waktu naik.
func (s *Store) ReadingsSince(ts int64) ([]model.Reading, error) {
	rows, err := s.db.Query(`SELECT `+readingCols+` FROM readings WHERE ts >= ? ORDER BY ts ASC, id ASC`, ts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Reading
	for rows.Next() {
		r, err := scanReading(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) PurgeOlderThan(ts int64) error {
	for _, t := range []string{"readings", "events"} {
		if _, err := s.db.Exec(`DELETE FROM `+t+` WHERE ts < ?`, ts); err != nil {
			return err
		}
	}
	return nil
}

// ---------- scans (barang masuk) ----------

// InsertScan menyimpan barang masuk (atau memperbarui timestamp, JSON, & snapshot jika batch dipindai ulang).
func (s *Store) InsertScan(sc model.Scan) (bool, error) {
	res, err := s.db.Exec(`INSERT INTO scans(ts, device_id, batch_id, material, payload, snapshot) VALUES(?,?,?,?,?,?)
		ON CONFLICT(batch_id) DO UPDATE SET
			ts=excluded.ts,
			device_id=excluded.device_id,
			material=excluded.material,
			payload=CASE WHEN excluded.payload<>'' THEN excluded.payload ELSE scans.payload END,
			snapshot=CASE WHEN excluded.snapshot<>'' THEN excluded.snapshot ELSE scans.snapshot END`,
		sc.TS, sc.DeviceID, sc.BatchID, sc.Material, sc.Payload, sc.Snapshot)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func (s *Store) CountScans() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM scans`).Scan(&n)
	return n, err
}

func (s *Store) CountScansSince(ts int64) (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM scans WHERE ts >= ?`, ts).Scan(&n)
	return n, err
}

func (s *Store) ScansSince(ts int64) ([]model.Scan, error) {
	return s.queryScans(`SELECT id, ts, device_id, batch_id, material, snapshot FROM scans WHERE ts >= ? ORDER BY ts ASC`, ts)
}

func (s *Store) RecentScans(limit int) ([]model.Scan, error) {
	return s.queryScans(`SELECT id, ts, device_id, batch_id, material, snapshot FROM scans ORDER BY ts DESC, id DESC LIMIT ?`, limit)
}

func (s *Store) queryScans(q string, arg any) ([]model.Scan, error) {
	rows, err := s.db.Query(q, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Scan{}
	for rows.Next() {
		var x model.Scan
		if err := rows.Scan(&x.ID, &x.TS, &x.DeviceID, &x.BatchID, &x.Material, &x.Snapshot); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

// ---------- events ----------

func (s *Store) InsertEvent(e model.Event) error {
	_, err := s.db.Exec(`INSERT INTO events(ts, device_id, jenis, level, pesan) VALUES(?,?,?,?,?)`,
		e.TS, e.DeviceID, e.Jenis, e.Level, e.Pesan)
	return err
}

func (s *Store) RecentEvents(limit int) ([]model.Event, error) {
	rows, err := s.db.Query(`SELECT id, ts, device_id, jenis, level, pesan FROM events ORDER BY ts DESC, id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Event{}
	for rows.Next() {
		var e model.Event
		if err := rows.Scan(&e.ID, &e.TS, &e.DeviceID, &e.Jenis, &e.Level, &e.Pesan); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// RecentTracking: batch dari barcode (yang punya data lengkap), terbaru dulu.
func (s *Store) RecentTracking(limit int) ([]model.Scan, error) {
	return s.queryTracking(`SELECT id, ts, device_id, batch_id, material, payload, snapshot FROM scans WHERE payload <> '' ORDER BY ts DESC, id DESC LIMIT ?`, limit)
}

// TrackingByBatch mengembalikan satu batch berdasarkan kode box, atau nil jika tidak ada.
func (s *Store) TrackingByBatch(batch string) (*model.Scan, error) {
	list, err := s.queryTracking(`SELECT id, ts, device_id, batch_id, material, payload, snapshot FROM scans WHERE batch_id = ? AND payload <> ''`, batch)
	if err != nil || len(list) == 0 {
		return nil, err
	}
	return &list[0], nil
}

func (s *Store) queryTracking(q string, arg any) ([]model.Scan, error) {
	rows, err := s.db.Query(q, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Scan{}
	for rows.Next() {
		var x model.Scan
		if err := rows.Scan(&x.ID, &x.TS, &x.DeviceID, &x.BatchID, &x.Material, &x.Payload, &x.Snapshot); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
