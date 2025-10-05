package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aryadiwwt/synctodb-anggarandetail/config"
	"github.com/aryadiwwt/synctodb-anggarandetail/fetcher"
	"github.com/aryadiwwt/synctodb-anggarandetail/storer"
	"github.com/aryadiwwt/synctodb-anggarandetail/synchronizer"
	"github.com/joho/godotenv"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: Could not load .env file")
	}
	logger := log.New(os.Stdout, "DATA-SYNC-SERVICE: ", log.LstdFlags|log.Lshortfile)

	// 1. Load Configuration
	cfg := config.New()
	// Pastikan username dan password tidak kosong
	if cfg.APIUsername == "" || cfg.APIPassword == "" {
		logger.Fatal("FATAL: API_USERNAME and API_PASSWORD environment variables must be set.")
	}
	// 2. Setup Dependencies
	// Koneksi DB
	db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.Fatalf("FATAL: Could not connect to database: %v", err)
	}
	defer db.Close()
	// ---- KONFIGURASI POOL----

	// SetConnMaxLifetime: Durasi maksimum koneksi boleh dibuka.
	// Mengaturnya lebih rendah dari timeout firewall (misal 5 menit) akan
	// secara otomatis mendaur ulang koneksi sebelum diputus oleh firewall.
	db.SetConnMaxLifetime(10 * time.Minute)

	// SetMaxIdleConns: Jumlah maksimum koneksi yang boleh idle di pool.
	db.SetMaxIdleConns(10)

	// SetMaxOpenConns: Jumlah maksimum koneksi yang boleh dibuka ke database.
	db.SetMaxOpenConns(100)

	// SetConnMaxIdleTime: Durasi maksimum koneksi boleh idle sebelum ditutup.
	// Ini membantu membuang koneksi yang tidak terpakai.
	db.SetConnMaxIdleTime(10 * time.Minute)

	// ---------------------------------------------
	// HTTP Client - dikonfigurasi sekali dan di-inject
	httpClient := &http.Client{
		Timeout: 90 * time.Second,
	}

	// 3. Create Concrete Implementations
	// Berikan semua konfigurasi yang dibutuhkan oleh Fetcher
	dataFetcher := fetcher.NewHTTPFetcher(
		httpClient,
		cfg.APIURL,
		cfg.APILoginURL,
		cfg.APIUsername,
		cfg.APIPassword,
		cfg.APIDataTahun,
		cfg.APIDataKdProv,
		cfg.APIDataKdKab,
	)
	dataStorer := storer.NewDBStorer(db)

	// 4. Compose The Application
	// Inject semua dependensi ke dalam synchronizer
	postSync := synchronizer.NewAnggaranDetailSynchronizer(dataFetcher, dataStorer, logger)

	// 5. Run The Application
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if err := postSync.Synchronize(ctx); err != nil {
		logger.Fatalf("FATAL: Post synchronization process failed: %v", err)
	}

	logger.Println("Application finished successfully.")
}
