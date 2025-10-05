package storer

import (
	"context"

	"github.com/aryadiwwt/synctodb-anggarandetail/domain"
	customErrors "github.com/aryadiwwt/synctodb-anggarandetail/errors"

	"github.com/jmoiron/sqlx"
)

// Storer mendefinisikan kontrak untuk menyimpan data post.
type Storer interface {
	StoreAnggaranDetails(ctx context.Context, details []domain.AnggaranDetail) error
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
