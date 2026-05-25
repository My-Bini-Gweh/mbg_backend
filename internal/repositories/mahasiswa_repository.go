package repositories

import (
	"context"
	"database/sql"

	"mbg-backend/internal/models"
)

type MahasiswaRepository struct {
	db *sql.DB
}

func NewMahasiswaRepository(db *sql.DB) *MahasiswaRepository {
	return &MahasiswaRepository{db: db}
}

func (r *MahasiswaRepository) GetByID(ctx context.Context, id uint64) (*models.Mahasiswa, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	var mahasiswa models.Mahasiswa
	err := r.db.QueryRowContext(ctx, `
		SELECT id_mahasiswa, nrp, nama_mahasiswa, email, role, status,
			DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s')
		FROM mahasiswa
		WHERE id_mahasiswa = ?
	`, id).Scan(
		&mahasiswa.ID,
		&mahasiswa.NRP,
		&mahasiswa.NamaMahasiswa,
		&mahasiswa.Email,
		&mahasiswa.Role,
		&mahasiswa.Status,
		&mahasiswa.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &mahasiswa, nil
}

func (r *MahasiswaRepository) GetWalletByMahasiswaID(ctx context.Context, mahasiswaID uint64) (*models.Wallet, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	var wallet models.Wallet
	err := r.db.QueryRowContext(ctx, `
		SELECT id_wallet, mahasiswa_id, jenis_wallet, saldo
		FROM wallet
		WHERE mahasiswa_id = ?
	`, mahasiswaID).Scan(
		&wallet.IDWallet,
		&wallet.MahasiswaID,
		&wallet.JenisWallet,
		&wallet.Saldo,
	)
	if err != nil {
		return nil, err
	}

	return &wallet, nil
}

func (r *MahasiswaRepository) WalletBelongsToMahasiswa(ctx context.Context, walletID string, mahasiswaID uint64) (bool, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM wallet
			WHERE id_wallet = ? AND mahasiswa_id = ?
		)
	`, walletID, mahasiswaID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *MahasiswaRepository) ListTransactionsByMahasiswa(ctx context.Context, mahasiswaID uint64) ([]models.Transaksi, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, `
		SELECT id_transaksi, kode_transaksi, jenis_transaksi, nominal, status, waktu,
			bank_id_bank, merchant_id, wallet_id_wallet, keterangan, bank_name, merchant_name
		FROM v_riwayat_transaksi_mahasiswa
		WHERE id_mahasiswa = ?
		ORDER BY waktu DESC
	`, mahasiswaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := make([]models.Transaksi, 0)
	for rows.Next() {
		var transaksi models.Transaksi
		var bankID sql.NullInt64

		err := rows.Scan(
			&transaksi.IDTransaksi,
			&transaksi.KodeTransaksi,
			&transaksi.JenisTransaksi,
			&transaksi.Nominal,
			&transaksi.Status,
			&transaksi.Waktu,
			&bankID,
			&transaksi.MerchantID,
			&transaksi.WalletIDWallet,
			&transaksi.Keterangan,
			&transaksi.BankName,
			&transaksi.MerchantName,
		)
		if err != nil {
			return nil, err
		}

		transaksi.BankIDBank = nullableUint64(bankID)
		transactions = append(transactions, transaksi)
	}

	return transactions, rows.Err()
}
