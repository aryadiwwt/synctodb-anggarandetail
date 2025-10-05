package config

import (
	"os"
	"strconv"
)

// Config menyimpan semua konfigurasi aplikasi.
type Config struct {
	// Konfigurasi yang sudah ada
	APIURL      string
	DatabaseURL string
	// Konfigurasi untuk request body data
	APILoginURL string
	APIUsername string
	APIPassword string
	// Konfigurasi untuk request body data
	APIDataTahun  int
	APIDataKdProv string
	APIDataKdKab  string
}

// New memuat konfigurasi dari environment variables.
func New() *Config {
	// Mengambil tahun sebagai string lalu mengonversinya ke integer
	tahunStr := getEnv("API_DATA_TAHUN", "2025")
	tahun, err := strconv.Atoi(tahunStr)
	if err != nil {
		// Jika ada error (misal format salah), gunakan nilai default
		tahun = 2025
	}

	return &Config{
		APIURL:        getEnv("API_URL", "https://konsolidasi-apbdesa.kemendagri.go.id/api/rekap/anggaranrealisasikegobyek/detail"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/dbname?sslmode=disable"),
		APILoginURL:   getEnv("API_LOGIN_URL", "https://konsolidasi-apbdesa.kemendagri.go.id/api/login"),
		APIUsername:   getEnv("API_USERNAME", ""),
		APIPassword:   getEnv("API_PASSWORD", ""),
		APIDataTahun:  tahun,
		APIDataKdProv: getEnv("API_DATA_KD_PROV", "51"),
		APIDataKdKab:  getEnv("API_DATA_KD_KAB", "03"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
