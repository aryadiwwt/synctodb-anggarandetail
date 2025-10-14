package storer

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aryadiwwt/synctodb-anggarandetail/domain"
	customErrors "github.com/aryadiwwt/synctodb-anggarandetail/errors"

	"github.com/jmoiron/sqlx"
)

type Wilayah struct {
	KodeProvinsi  string `db:"provinsi_id"`
	KodeKabupaten string `db:"kota_id"`
}

// Storer mendefinisikan kontrak untuk menyimpan data post.
type Storer interface {
	StoreAnggaranDetails(ctx context.Context, details []domain.AnggaranDetail) error
	GetWilayahByProvinsi(ctx context.Context, kodeProvinsi []string) ([]Wilayah, error)
}

// Implementasi fungsi untuk memfilter berdasarkan kd_prov
func (s *dbStorer) GetWilayahByProvinsi(ctx context.Context, kodeProvinsi []string) ([]Wilayah, error) {
	var wilayah []Wilayah

	// Query dasar
	baseQuery := `SELECT provinsi_id, kota_id FROM master_kota`

	var args []interface{}

	// Jika daftar provinsi diberikan, tambahkan klausa WHERE IN
	if len(kodeProvinsi) > 0 {
		baseQuery += ` WHERE provinsi_id IN (?)`
		args = append(args, kodeProvinsi)
	}

	baseQuery += ` ORDER BY provinsi_id, kota_id`

	// sqlx.In secara aman akan mengubah query (?) menjadi ($1, $2, ...)
	// dan menyesuaikan argumennya. Ini cara aman untuk klausa IN.
	query, args, err := sqlx.In(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat query IN: %w", err)
	}

	// Rebind query agar sesuai dengan placeholder PostgreSQL ($1, $2)
	query = s.db.Rebind(query)

	err = s.db.SelectContext(ctx, &wilayah, query, args...)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data wilayah yang difilter: %w", err)
	}

	// Lakukan loop untuk memformat kode kabupaten setelah data didapat
	for i := range wilayah {
		// Ambil kode kabupaten mentah
		kodeKabStr := wilayah[i].KodeKabupaten

		// Ubah string menjadi integer
		num, err := strconv.Atoi(kodeKabStr)
		if err != nil {
			// Jika gagal (misal format tidak standar), biarkan apa adanya dan beri peringatan
			fmt.Printf("Peringatan: Format kd_kab '%s' tidak valid, tidak diformat.\n", kodeKabStr)
			continue
		}

		// Format integer menjadi string 2 digit dengan awalan nol
		wilayah[i].KodeKabupaten = fmt.Sprintf("%02d", num)
	}

	return wilayah, nil
}

type dbStorer struct {
	db *sqlx.DB
}

func NewDBStorer(db *sqlx.DB) Storer {
	return &dbStorer{db: db}
}

const (
	// Query disimpan sebagai konstanta untuk menghindari 'magic strings'
	// dan memudahkan pengelolaan.
	upsertAnggaranDetailQuery = `INSERT INTO siskeudes_detail_anggaran (
            tahun, kd_prov, nama_provinsi, kd_kab, nama_kabupaten,
            kd_kec, nama_kecamatan, kd_desa, nama_desa, kd_bid,
            nama_bidang, kd_sub, nama_subbidang, id_keg, nama_kegiatan,
            kd_subrinci, kode_sumber, akun, nama_akun, kelompok, nama_kelompok,
            jenis, nama_jenis, obyek, nama_obyek, anggaran1, anggaran2,
            realisasi1, realisasi2
        ) VALUES (
            :tahun, :kd_prov, :nama_provinsi, :kd_kab, :nama_kabupaten,
            :kd_kec, :nama_kecamatan, :kd_desa, :nama_desa, :kd_bid,
            :nama_bidang, :kd_sub, :nama_subbidang, :id_keg, :nama_kegiatan,
            :kd_subrinci, :kode_sumber, :akun, :nama_akun, :kelompok, :nama_kelompok,
            :jenis, :nama_jenis, :obyek, :nama_obyek, :anggaran1, :anggaran2,
            :realisasi1, :realisasi2
        )
        ON CONFLICT (kd_prov, kd_kec, kd_desa, id_keg, akun, obyek, tahun) DO UPDATE SET
            anggaran1 = EXCLUDED.anggaran1,
            anggaran2 = EXCLUDED.anggaran2,
            realisasi1 = EXCLUDED.realisasi1,
            realisasi2 = EXCLUDED.realisasi2;`
)

func (s *dbStorer) StoreAnggaranDetails(ctx context.Context, details []domain.AnggaranDetail) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return &customErrors.ErrDBOperationFailed{Operation: "begin_transaction", Err: err}
	}
	defer tx.Rollback() // Aman untuk dipanggil meskipun sudah di-commit.

	for _, detail := range details {
		if _, err := tx.NamedExecContext(ctx, upsertAnggaranDetailQuery, detail); err != nil {
			return &customErrors.ErrDBOperationFailed{Operation: "upsert_post", Err: err}
		}
	}

	if err := tx.Commit(); err != nil {
		return &customErrors.ErrDBOperationFailed{Operation: "commit_transaction", Err: err}
	}

	return nil
}
