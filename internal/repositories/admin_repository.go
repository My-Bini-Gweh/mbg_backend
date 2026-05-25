package repositories

import (
	"context"
	"database/sql"

	"mbg-backend/internal/models"
)

type AdminRepository struct {
	db *sql.DB
}

func NewAdminRepository(db *sql.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

func (r *AdminRepository) Summary(ctx context.Context) (*models.AdminSummary, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	var summary models.AdminSummary
	err := r.db.QueryRowContext(ctx, `
		SELECT
			(SELECT COUNT(*) FROM mahasiswa),
			(SELECT COUNT(*) FROM merchant WHERE status = 'ACTIVE'),
			(SELECT COUNT(*) FROM transaksi),
			(SELECT COALESCE(SUM(nominal), 0.00) FROM transaksi WHERE status = 'SUCCESS')
	`).Scan(
		&summary.TotalUsers,
		&summary.TotalMerchants,
		&summary.TotalTransactions,
		&summary.TotalSuccessfulAmount,
	)
	if err != nil {
		return nil, err
	}

	return &summary, nil
}

func (r *AdminRepository) ListTransactions(ctx context.Context) ([]models.Transaksi, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id_transaksi, t.kode_transaksi, t.jenis_transaksi, t.nominal, t.status, t.waktu,
			t.bank_id_bank, COALESCE(t.merchant_id, ''), t.wallet_id_wallet,
			COALESCE(t.keterangan, ''), COALESCE(b.nama_bank, ''), COALESCE(m.nama_merchant, '')
		FROM transaksi t
		LEFT JOIN bank b ON b.id_bank = t.bank_id_bank
		LEFT JOIN merchant m ON m.id_merchant = t.merchant_id
		ORDER BY t.waktu DESC
		LIMIT 200
	`)
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

func (r *AdminRepository) ListAuditLogs(ctx context.Context) ([]models.AuditLog, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, `
		SELECT id_audit, transaksi_id, action, description, created_at
		FROM audit_logs
		ORDER BY created_at DESC
		LIMIT 200
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]models.AuditLog, 0)
	for rows.Next() {
		var log models.AuditLog
		var transaksiID sql.NullInt64
		if err := rows.Scan(&log.IDAudit, &transaksiID, &log.Action, &log.Description, &log.CreatedAt); err != nil {
			return nil, err
		}

		log.TransaksiID = nullableUint64(transaksiID)
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

func (r *AdminRepository) ListDailyReports(ctx context.Context) ([]models.DailyReport, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, `
		SELECT DATE_FORMAT(tanggal, '%Y-%m-%d'), total_transaksi, total_nominal, total_topup, total_payment, total_transaksi_success
		FROM v_laporan_transaksi_harian
		ORDER BY tanggal DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reports := make([]models.DailyReport, 0)
	for rows.Next() {
		var report models.DailyReport
		if err := rows.Scan(
			&report.Tanggal,
			&report.TotalTransaksi,
			&report.TotalNominal,
			&report.TotalTopup,
			&report.TotalPayment,
			&report.TotalTransaksiOK,
		); err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}

	return reports, rows.Err()
}

func nullableUint64(value sql.NullInt64) *uint64 {
	if !value.Valid {
		return nil
	}
	converted := uint64(value.Int64)
	return &converted
}
