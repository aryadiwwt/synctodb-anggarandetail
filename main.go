package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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

	// Load Configuration
	cfg := config.New()
	// Pastikan username dan password tidak kosong
	if cfg.APIUsername == "" || cfg.APIPassword == "" {
		logger.Fatal("FATAL: API_USERNAME and API_PASSWORD environment variables must be set.")
	}
	// Definisikan flag untuk command line
	// Akan membaca flag seperti: -prov="11,12,51"
	provinsiPtr := flag.String("prov", "", "Daftar kode provinsi yang dipisahkan koma (contoh: 11,12,51)")
	kabupatenPtr := flag.String("kab", "", "Kode kabupaten untuk memulai proses (opsional)")
	flag.Parse() // Baca semua flag yang didefinisikan
	// Setup Dependencies
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
		Timeout: 120 * time.Minute,
	}
	// Proses input dari flag
	var daftarProvinsi []string
	if *provinsiPtr != "" {
		// Pisahkan string menjadi slice berdasarkan koma
		daftarProvinsi = strings.Split(*provinsiPtr, ",")
		logger.Printf("Akan memproses data untuk provinsi: %v", daftarProvinsi)
	} else {
		logger.Println("Tidak ada kode provinsi yang ditentukan. Untuk memproses semua, biarkan flag -prov kosong.")
	}
	// Ambil nilai dari flag kabupaten
	startKabupaten := *kabupatenPtr
	if startKabupaten != "" {
		// Ambil nilai mentah dari flag
		kodeKabStr := strings.TrimSpace(*kabupatenPtr)
		// Lakukan formatting yang sama seperti yang kita lakukan pada data lain
		num, err := strconv.Atoi(kodeKabStr)
		if err != nil {
			logger.Fatalf("Error: Kode kabupaten '%s' bukan angka yang valid.", kodeKabStr)
		}
		// Format menjadi string 2 digit
		startKabupaten = fmt.Sprintf("%02d", num)
		logger.Printf("Proses akan dimulai dari kabupaten dengan kode yang diformat: %s", startKabupaten)
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
	)
	dataStorer := storer.NewDBStorer(db)

	// 4. Compose The Application
	// Inject semua dependensi ke dalam synchronizer
	postSync := synchronizer.NewAnggaranDetailSynchronizer(dataFetcher, dataStorer, logger)

	// 5. Run The Application
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if err := postSync.Synchronize(ctx, daftarProvinsi, startKabupaten); err != nil {
		logger.Fatalf("FATAL: Post synchronization process failed: %v", err)
	}

	logger.Println("Application finished successfully.")
}
