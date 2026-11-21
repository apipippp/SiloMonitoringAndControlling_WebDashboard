package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "time/tzdata" // zona waktu tertanam, agar "Asia/Jakarta" tersedia juga di Windows

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"golang.org/x/crypto/bcrypt"

	"silo/internal/config"
	"silo/internal/handler"
	"silo/internal/mqttclient"
	"silo/internal/service"
	"silo/internal/sim"
	"silo/internal/store"
)

func main() {
	cfg := config.Load()

	if cfg.JWTSecret == "ganti-rahasia-ini-sebelum-dipakai" {
		log.Println("PERINGATAN: JWT_SECRET masih bawaan. Ganti di berkas .env sebelum dipakai sungguhan.")
	}

	st, err := store.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("gagal membuka database: %v", err)
	}
	defer st.Close()

	// Buat/perbarui akun admin dari .env
	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPass), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	if err := st.UpsertUser(strings.ToLower(cfg.AdminEmail), cfg.AdminName, string(hash)); err != nil {
		log.Fatalf("gagal membuat akun admin: %v", err)
	}

	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		log.Printf("zona waktu %q tidak dikenal, memakai UTC+7", cfg.Timezone)
		loc = time.FixedZone("WIB", 7*3600)
	}

	svc := service.New(st, cfg)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var ctl handler.Controller
	if cfg.Simulator {
		sm := sim.New(svc)
		if err := sm.Backfill(cfg.SimBackfillHours); err != nil {
			log.Fatalf("gagal mengisi data awal simulator: %v", err)
		}
		go sm.Run(ctx)
		ctl = sm
		log.Println("MODE SIMULATOR aktif: data dibangkitkan otomatis, MQTT tidak dipakai")
	} else {
		mc := mqttclient.New(cfg, svc)
		mc.Connect()
		ctl = mc
	}

	// Hapus data lama secara berkala
	go func() {
		purge := func() {
			cutoff := time.Now().AddDate(0, 0, -cfg.RetentionDays).Unix()
			if err := st.PurgeOlderThan(cutoff); err != nil {
				log.Printf("gagal membersihkan data lama: %v", err)
			}
		}
		purge()
		tk := time.NewTicker(time.Hour)
		defer tk.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tk.C:
				purge()
			}
		}
	}()

	app := fiber.New(fiber.Config{ErrorHandler: handler.ErrorHandler})

	// CORS harus dipasang paling awal (lihat Modul 9, Langkah 2).
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSOrigins,
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	h := &handler.Handler{Store: st, Cfg: cfg, Ctl: ctl, Loc: loc}
	h.Register(app)

	go func() {
		<-ctx.Done()
		_ = app.Shutdown()
	}()

	log.Printf("API berjalan di http://localhost:%s  (CORS: %v)", cfg.Port, cfg.CORSOrigins)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
