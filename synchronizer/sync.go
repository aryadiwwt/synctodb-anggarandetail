package synchronizer

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aryadiwwt/synctodb-anggarandetail/domain"
	"github.com/aryadiwwt/synctodb-anggarandetail/fetcher"
	"github.com/aryadiwwt/synctodb-anggarandetail/storer"
)

// PostSynchronizer mengorkestrasi proses sinkronisasi data post.
type AnggaranDetailSynchronizer struct {
	fetcher fetcher.Fetcher
	storer  storer.Storer
	log     *log.Logger
}

func NewAnggaranDetailSynchronizer(f fetcher.Fetcher, s storer.Storer, l *log.Logger) *AnggaranDetailSynchronizer {
	return &AnggaranDetailSynchronizer{
		fetcher: f,
		storer:  s,
		log:     l,
	}
}

func (s *AnggaranDetailSynchronizer) Synchronize(ctx context.Context, kodeProvinsi []string) error {
	s.log.Println("Starting output detail synchronization...")

	daftarWilayah, err := s.storer.GetWilayahByProvinsi(ctx, kodeProvinsi)
	if err != nil {
		s.log.Fatalf("Gagal mendapatkan daftar wilayah: %v", err)
	}

	if len(daftarWilayah) == 0 {
		s.log.Println("Tidak ada data wilayah yang ditemukan untuk diproses. Selesai.")
		return nil
	}

	s.log.Printf("Akan memproses data untuk %d kabupaten/kota...", len(daftarWilayah))

	// Lakukan loop untuk setiap wilayah
	for _, wilayah := range daftarWilayah {
		s.log.Printf("=== Memulai proses untuk Provinsi: %s, Kabupaten: %s ===", wilayah.KodeProvinsi, wilayah.KodeKabupaten)

		// Fetch data untuk wilayah saat ini
		// Perhatikan bagaimana kita sekarang memberikan kode wilayah sebagai argumen
		details, err := s.fetcher.FetchAnggaranDetails(ctx, wilayah.KodeProvinsi, wilayah.KodeKabupaten)
		if err != nil {
			s.log.Printf("ERROR saat mengambil data untuk Prov %s Kab %s: %v. Melanjutkan ke wilayah berikutnya.", wilayah.KodeProvinsi, wilayah.KodeKabupaten, err)
			continue // Lanjut ke iterasi berikutnya jika ada error
		}

		if len(details) == 0 {
			s.log.Println("Tidak ada data untuk wilayah ini.")
			continue
		}

		// Transformasi data
		s.log.Println("Transforming data...")
		transformedDetails := transformDetails(details)
		s.log.Println("Data transformation complete.")

		// Simpan data ke database (menggunakan batch processing)
		if err := s.storer.StoreAnggaranDetails(ctx, transformedDetails); err != nil {
			s.log.Printf("ERROR saat menyimpan data untuk Prov %s Kab %s: %v", wilayah.KodeProvinsi, wilayah.KodeKabupaten, err)
			continue
		}

		s.log.Printf("=== Selesai memproses untuk Provinsi: %s, Kabupaten: %s. Total %d data disimpan. ===", wilayah.KodeProvinsi, wilayah.KodeKabupaten, len(details))

		// Opsional: Beri jeda singkat antar request untuk tidak membebani API
		s.log.Println("Memberi jeda 30 Detik...")
		time.Sleep(30 * time.Second)
	}

	s.log.Println("Semua proses sinkronisasi untuk seluruh wilayah telah selesai.")
	return nil
}

// transformDetails berisi logika untuk mengubah data
func transformDetails(details []domain.AnggaranDetail) []domain.AnggaranDetail {
	// Loop melalui setiap record dan modifikasi nilainya
	for i := range details {
		// Simpan nilai asli untuk digunakan dalam penggabungan
		prov := details[i].KodeProvinsi
		kab := details[i].KodeKabupaten
		kec := details[i].KodeKecamatan
		desa := details[i].KodeDesa

		// Aturan 1: kd_kab = kd_prov.kd_kab
		details[i].KodeKabupaten = fmt.Sprintf("%s.%s", prov, kab)

		// Aturan 2: kd_kec = kd_prov.kd_kab.kd_kec
		details[i].KodeKecamatan = fmt.Sprintf("%s.%s.%s", prov, kab, kec)

		// Aturan 3: kd_desa = kd_prov.kd_kab.kd_desa
		// Membersihkan titik di akhir kd_desa dari API jika ada
		cleanedDesa := strings.TrimSuffix(desa, ".")
		details[i].KodeDesa = fmt.Sprintf("%s.%s.%s", prov, kab, cleanedDesa)
	}

	return details
}
