package fetcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aryadiwwt/synctodb-anggarandetail/domain" // Ganti dengan domain Anda, misal: domain.AnggaranDetail
)

// Definisikan struct untuk menampung response dari API login
type loginResponse struct {
	Token string `json:"token"`
	// Tambahkan field lain jika ada, misal: "user", "expires_in", dll.
}

// Definisikan struct untuk request body login
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Definisikan struct untuk request body data yang diinginkan ex: output detail
type dataRequestBody struct {
	Tahun  int    `json:"tahun"`
	KdProv string `json:"kd_prov"`
	KdKab  string `json:"kd_kab"`
}

type Fetcher interface {
	FetchAnggaranDetails(ctx context.Context) ([]domain.AnggaranDetail, error)
}

// httpFetcher sekarang memiliki state untuk token dan info login
type httpFetcher struct {
	client    *http.Client
	dataURL   string
	loginURL  string
	username  string
	password  string
	authToken string // Tempat menyimpan token setelah login berhasil
	tahun     int
	kdProv    string
	kdKab     string
}

// NewHTTPFetcher sekarang menerima konfigurasi login
func NewHTTPFetcher(client *http.Client, dataURL, loginURL, username, password string, tahun int, kdProv, kdKab string) Fetcher {
	return &httpFetcher{
		client:   client,
		dataURL:  dataURL,
		loginURL: loginURL,
		username: username,
		password: password,
		tahun:    tahun,
		kdProv:   kdProv,
		kdKab:    kdKab,
	}
}

// FetchOutputDetail adalah nama baru yang lebih deskriptif
func (f *httpFetcher) FetchAnggaranDetails(ctx context.Context) ([]domain.AnggaranDetail, error) {
	// 1. Cek token. Jika kosong, lakukan autentikasi.
	if f.authToken == "" {
		if err := f.authenticate(ctx); err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
	}

	// 2. Buat request body (seperti pada metode POST)
	dataPayload := dataRequestBody{
		Tahun:  f.tahun,
		KdProv: f.kdProv,
		KdKab:  f.kdKab,
	}
	body, err := json.Marshal(dataPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data request body: %w", err)
	}

	// 3. Buat request GET, namun kali ini dengan body (tidak standar)
	// Perhatikan: Method adalah GET, URL dasar, dan body disertakan.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.dataURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create non-standard GET request with body: %w", err)
	}
	// 4. Tambahkan token ke Authorization header
	req.Header.Set("Authorization", "Bearer "+f.authToken)
	// Header Content-Type mungkin diperlukan oleh server untuk memproses body
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute data request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Di sini kita bisa menambahkan logika jika token expired (misal status 401)
		// untuk melakukan re-autentikasi.
		return nil, fmt.Errorf("unexpected status code on data fetch: %d", resp.StatusCode)
	}

	// 5. DECODE RESPONSE (BAGIAN YANG DIUBAH)
	// =================================================================

	// BARU: Definisikan struct yang cocok dengan response API yang kompleks
	// Struct untuk objek 'data' yang berisi pagination dan array data utama
	type PaginatedData struct {
		CurrentPage int                     `json:"current_page"`
		Data        []domain.AnggaranDetail `json:"data"` // Ini adalah array yang kita inginkan!
		Total       int                     `json:"total"`
		// Field lain seperti last_page, links, dll, akan diabaikan jika tidak didefinisikan
	}

	// Struct untuk response API level teratas
	type ApiResponse struct {
		Result  string        `json:"r"`
		Error   string        `json:"e"`
		Data    PaginatedData `json:"data"` // Data sekarang adalah sebuah objek, bukan array
		Message string        `json:"message"`
	}

	// DIUBAH: Decode response ke dalam struct ApiResponse yang baru
	var fullResponse ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&fullResponse); err != nil {
		return nil, fmt.Errorf("failed to decode full api response: %w", err)
	}

	// DIUBAH: Kembalikan hanya array yang kita butuhkan
	return fullResponse.Data.Data, nil
	// =================================================================
}

// authenticate adalah fungsi internal untuk login dan menyimpan token.
func (f *httpFetcher) authenticate(ctx context.Context) error {
	loginPayload := loginRequest{
		Username: f.username,
		Password: f.password,
	}

	body, err := json.Marshal(loginPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal login request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.loginURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status code: %d", resp.StatusCode)
	}

	var lr loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return fmt.Errorf("failed to decode login response: %w", err)
	}

	if lr.Token == "" {
		return fmt.Errorf("login successful but token is empty")
	}

	// Simpan token untuk request selanjutnya
	f.authToken = lr.Token
	fmt.Println("Successfully authenticated and obtained token.") // Ganti dengan logger
	return nil
}
