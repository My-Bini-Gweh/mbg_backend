package models

import "time"

type Pagination struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type PaginatedResult struct {
	Items      any        `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type AdminMahasiswa struct {
	ID            uint64     `json:"id_mahasiswa"`
	NRP           string     `json:"nrp"`
	NamaMahasiswa string     `json:"nama_mahasiswa"`
	Email         string     `json:"email"`
	Role          string     `json:"role"`
	Status        string     `json:"status"`
	WalletID      string     `json:"wallet_id,omitempty"`
	WalletBalance string     `json:"wallet_balance,omitempty"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type AdminAuthRecord struct {
	IDAuth        uint64     `json:"id_auth"`
	MahasiswaID   uint64     `json:"mahasiswa_id"`
	MahasiswaName string     `json:"mahasiswa_name"`
	NRP           string     `json:"nrp"`
	Email         string     `json:"email"`
	PasswordHash  string     `json:"password_hash"`
	PINHash       *string    `json:"pin_hash"`
	LastLoginAt   *time.Time `json:"last_login_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

type AdminWallet struct {
	IDWallet      string    `json:"id_wallet"`
	MahasiswaID   uint64    `json:"mahasiswa_id"`
	MahasiswaName string    `json:"mahasiswa_name"`
	NRP           string    `json:"nrp"`
	JenisWallet   string    `json:"jenis_wallet"`
	Saldo         string    `json:"saldo"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type AdminBank struct {
	IDBank     uint64    `json:"id_bank"`
	NamaBank   string    `json:"nama_bank"`
	KodeBank   string    `json:"kode_bank"`
	BiayaAdmin string    `json:"biaya_admin"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}

type AdminRekening struct {
	IDRekening    uint64    `json:"id_rekening"`
	MahasiswaID   uint64    `json:"mahasiswa_id"`
	MahasiswaName string    `json:"mahasiswa_name"`
	NRP           string    `json:"nrp"`
	BankID        uint64    `json:"bank_id_bank"`
	BankName      string    `json:"bank_name"`
	NoRekening    string    `json:"no_rekening"`
	NamaPemilik   string    `json:"nama_pemilik"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
}

type AdminMerchant struct {
	IDMerchant    string    `json:"id_merchant"`
	NamaMerchant  string    `json:"nama_merchant"`
	Kategori      string    `json:"kategori"`
	SaldoMerchant string    `json:"saldo_merchant"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type AdminTransaction struct {
	IDTransaksi    uint64    `json:"id_transaksi"`
	KodeTransaksi  string    `json:"kode_transaksi"`
	JenisTransaksi string    `json:"jenis_transaksi"`
	Nominal        string    `json:"nominal"`
	Status         string    `json:"status"`
	Waktu          time.Time `json:"waktu"`
	BankID         *uint64   `json:"bank_id_bank,omitempty"`
	BankName       string    `json:"bank_name,omitempty"`
	MerchantID     string    `json:"merchant_id,omitempty"`
	MerchantName   string    `json:"merchant_name,omitempty"`
	WalletID       string    `json:"wallet_id_wallet"`
	MahasiswaID    uint64    `json:"mahasiswa_id"`
	MahasiswaName  string    `json:"mahasiswa_name"`
	NRP            string    `json:"nrp"`
	Keterangan     string    `json:"keterangan,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type AdminAuditLog struct {
	IDAudit         uint64    `json:"id_audit"`
	TransaksiID     *uint64   `json:"transaksi_id,omitempty"`
	TransactionCode string    `json:"transaction_code,omitempty"`
	Action          string    `json:"action"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
}
