package models

import "time"

type Mahasiswa struct {
	ID            uint64 `json:"id_mahasiswa"`
	NRP           string `json:"nrp"`
	NamaMahasiswa string `json:"nama_mahasiswa"`
	Email         string `json:"email"`
	Role          string `json:"role,omitempty"`
	Status        string `json:"status,omitempty"`
	CreatedAt     string `json:"created_at,omitempty"`
}

type MahasiswaAuth struct {
	ID           uint64
	MahasiswaID  uint64
	PasswordHash string
	Role         string
}

type Wallet struct {
	IDWallet    string `json:"id_wallet"`
	MahasiswaID uint64 `json:"mahasiswa_id"`
	JenisWallet string `json:"jenis_wallet"`
	Saldo       string `json:"saldo"`
}

type Bank struct {
	IDBank     uint64 `json:"id_bank"`
	NamaBank   string `json:"nama_bank"`
	KodeBank   string `json:"kode_bank"`
	BiayaAdmin string `json:"biaya_admin"`
	IsActive   bool   `json:"is_active"`
}

type Merchant struct {
	IDMerchant    string `json:"id_merchant"`
	NamaMerchant  string `json:"nama_merchant"`
	Kategori      string `json:"kategori"`
	SaldoMerchant string `json:"saldo_merchant,omitempty"`
	Status        string `json:"status"`
}

type Transaksi struct {
	IDTransaksi    uint64    `json:"id_transaksi"`
	KodeTransaksi  string    `json:"kode_transaksi"`
	JenisTransaksi string    `json:"jenis_transaksi"`
	Nominal        string    `json:"nominal"`
	Status         string    `json:"status"`
	Waktu          time.Time `json:"waktu"`
	BankIDBank     *uint64   `json:"bank_id_bank,omitempty"`
	MerchantID     string    `json:"merchant_id,omitempty"`
	WalletIDWallet string    `json:"wallet_id_wallet"`
	Keterangan     string    `json:"keterangan,omitempty"`
	BankName       string    `json:"bank_name,omitempty"`
	MerchantName   string    `json:"merchant_name,omitempty"`
}

type AuditLog struct {
	IDAudit     uint64    `json:"id_audit"`
	TransaksiID *uint64   `json:"transaksi_id,omitempty"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type DailyReport struct {
	Tanggal          string `json:"tanggal"`
	TotalTransaksi   uint64 `json:"total_transaksi"`
	TotalNominal     string `json:"total_nominal"`
	TotalTopup       string `json:"total_topup"`
	TotalPayment     string `json:"total_payment"`
	TotalTransaksiOK uint64 `json:"total_transaksi_success"`
}
