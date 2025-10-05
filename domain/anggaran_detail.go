package domain

// AnggaranDetail merepresentasikan struktur data pendapatan.
type AnggaranDetail struct {
	Tahun         string `json:"tahun" db:"tahun"`
	KodeProvinsi  string `json:"kd_prov" db:"kd_prov"`
	NamaProvinsi  string `json:"nama_provinsi" db:"nama_provinsi"`
	KodeKabupaten string `json:"kd_kab" db:"kd_kab"`
	NamaKabupaten string `json:"nama_kabupaten" db:"nama_kabupaten"`
	KodeKecamatan string `json:"kd_kec" db:"kd_kec"`
	NamaKecamatan string `json:"nama_kecamatan" db:"nama_kecamatan"`
	KodeDesa      string `json:"kd_desa" db:"kd_desa"`
	NamaDesa      string `json:"nama_desa" db:"nama_desa"`

	// Gunakan pointer (*string) untuk field yang bisa bernilai null
	KodeBidang    *string `json:"kd_bid" db:"kd_bid"`
	NamaBidang    *string `json:"nama_bidang" db:"nama_bidang"`
	KodeSubBidang *string `json:"kd_sub" db:"kd_sub"`
	NamaSubBidang *string `json:"nama_subbidang" db:"nama_subbidang"`
	IDKegiatan    *string `json:"id_keg" db:"id_keg"`
	NamaKegiatan  *string `json:"nama_kegiatan" db:"nama_kegiatan"`

	KodeSubRinci string `json:"kd_subrinci" db:"kd_subrinci"`
	KodeSumber   string `json:"kode_sumber" db:"kode_sumber"`
	Akun         string `json:"akun" db:"akun"`
	NamaAkun     string `json:"nama_akun" db:"nama_akun"`
	Kelompok     string `json:"kelompok" db:"kelompok"`
	NamaKelompok string `json:"nama_kelompok" db:"nama_kelompok"`
	Jenis        string `json:"jenis" db:"jenis"`
	NamaJenis    string `json:"nama_jenis" db:"nama_jenis"`
	Obyek        string `json:"obyek" db:"obyek"`
	NamaObyek    string `json:"nama_obyek" db:"nama_obyek"`

	// Gunakan float64 dengan tag ",string" untuk angka dalam format string
	Anggaran1  float64 `json:"anggaran1,string" db:"anggaran1"`
	Anggaran2  float64 `json:"anggaran2,string" db:"anggaran2"`
	Realisasi1 float64 `json:"realisasi1,string" db:"realisasi1"`
	Realisasi2 float64 `json:"realisasi2,string" db:"realisasi2"`
}
