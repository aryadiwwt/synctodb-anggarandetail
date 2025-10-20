package fetcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	FetchAnggaranDetails(ctx context.Context, kdProv string, kdKab string) ([]domain.AnggaranDetail, error)
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
}

// NewHTTPFetcher sekarang menerima konfigurasi login
func NewHTTPFetcher(client *http.Client, dataURL, loginURL, username, password string, tahun int) Fetcher {
	return &httpFetcher{
		client:   client,
		dataURL:  dataURL,
		loginURL: loginURL,
		username: username,
		password: password,
		tahun:    tahun,
	}
}

func (f *httpFetcher) FetchAnggaranDetails(ctx context.Context, kdProv string, kdKab string) ([]domain.AnggaranDetail, error) {
	// 1. Proses autentikasi hanya dilakukan sekali di awal
	if f.authToken == "" {
		if err := f.authenticate(ctx); err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
	}

	// Slice untuk menampung hasil dari SEMUA halaman
	var allData []domain.AnggaranDetail

	// 2. Siapkan request body awal. Ini tidak akan berubah antar halaman.
	dataPayload := dataRequestBody{
		Tahun:  f.tahun,
		KdProv: kdProv,
		KdKab:  kdKab,
	}
	body, err := json.Marshal(dataPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data request body: %w", err)
	}

	// 3. Mulai loop dari URL data utama
	nextPageURL := f.dataURL

	for nextPageURL != "" { // Lakukan loop selama masih ada halaman berikutnya
		// Gunakan bytes.NewReader agar body bisa dibaca berulang kali di setiap loop
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, nextPageURL, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("failed to create request for page %s: %w", nextPageURL, err)
		}

		req.Header.Set("Authorization", "Bearer "+f.authToken)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		log.Printf("Fetching data from: %s", nextPageURL)

		resp, err := f.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to execute request for page %s: %w", nextPageURL, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code %d on page %s", resp.StatusCode, nextPageURL)
		}

		// Definisikan struct yang cocok dengan respons API yang kompleks
		type PaginatedData struct {
			Data        []domain.AnggaranDetail `json:"data"`          // Array data yang kita inginkan
			NextPageURL *string                 `json:"next_page_url"` // Pointer agar bisa null
		}
		type ApiResponse struct {
			Data PaginatedData `json:"data"`
		}

		var fullResponse ApiResponse
		if err := json.NewDecoder(resp.Body).Decode(&fullResponse); err != nil {
			return nil, fmt.Errorf("failed to decode api response for page %s: %w", nextPageURL, err)
		}

		// Tambahkan hasil dari halaman ini ke slice utama
		allData = append(allData, fullResponse.Data.Data...)

		// Perbarui URL untuk iterasi selanjutnya, atau hentikan loop
		if fullResponse.Data.NextPageURL != nil {
			nextPageURL = *fullResponse.Data.NextPageURL
		} else {
			nextPageURL = "" // Hentikan loop jika next_page_url adalah null
		}
	}

	log.Printf("Total %d records fetched from all pages.", len(allData))
	return allData, nil
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
