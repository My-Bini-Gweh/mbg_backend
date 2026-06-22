package repositories

import (
	"context"
	"database/sql"
	"strings"

	"mbg-backend/internal/models"
)

type AdminRepository struct {
	db *sql.DB
}

type AdminListParams struct {
	Page    int
	PerPage int
	Search  string
	Sort    string
	Order   string
	Filters map[string]string
}

type AdminMahasiswaInput struct {
	NRP           string
	NamaMahasiswa string
	Email         string
	Role          string
	Status        string
	PasswordHash  string
}

type AdminBankInput struct {
	NamaBank   string
	KodeBank   string
	BiayaAdmin string
	IsActive   bool
}

type AdminRekeningInput struct {
	MahasiswaID uint64
	BankID      uint64
	NoRekening  string
	NamaPemilik string
	IsActive    bool
}

type AdminMerchantInput struct {
	IDMerchant   string
	NamaMerchant string
	Kategori     string
	Status       string
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

func (r *AdminRepository) ListMahasiswa(ctx context.Context, params AdminListParams) (*models.PaginatedResult, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	where := " WHERE 1=1"
	args := make([]any, 0)
	if params.Search != "" {
		where += " AND (m.nrp LIKE ? OR m.nama_mahasiswa LIKE ? OR m.email LIKE ?)"
		term := "%" + params.Search + "%"
		args = append(args, term, term, term)
	}
	if role := params.Filters["role"]; role != "" {
		where += " AND m.role = ?"
		args = append(args, role)
	}
	if status := params.Filters["status"]; status != "" {
		where += " AND m.status = ?"
		args = append(args, status)
	}

	total, err := r.count(ctx, "SELECT COUNT(*) FROM mahasiswa m"+where, args...)
	if err != nil {
		return nil, err
	}

	sort := safeSort(params.Sort, map[string]string{
		"id": "m.id_mahasiswa", "nrp": "m.nrp", "name": "m.nama_mahasiswa",
		"email": "m.email", "role": "m.role", "status": "m.status", "created_at": "m.created_at",
	}, "m.created_at")
	query := `
		SELECT m.id_mahasiswa, m.nrp, m.nama_mahasiswa, m.email, m.role, m.status,
			COALESCE(w.id_wallet, ''), COALESCE(w.saldo, 0.00), a.last_login_at,
			m.created_at, m.updated_at
		FROM mahasiswa m
		LEFT JOIN wallet w ON w.mahasiswa_id = m.id_mahasiswa
		LEFT JOIN mahasiswa_auth a ON a.mahasiswa_id = m.id_mahasiswa` + where +
		" ORDER BY " + sort + " " + safeOrder(params.Order) + " LIMIT ? OFFSET ?"
	queryArgs := append(append([]any{}, args...), params.PerPage, offset(params))
	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.AdminMahasiswa, 0)
	for rows.Next() {
		item, err := scanAdminMahasiswa(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return paginated(items, params, total), nil
}

func (r *AdminRepository) GetMahasiswa(ctx context.Context, id uint64) (*models.AdminMahasiswa, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	return scanAdminMahasiswa(r.db.QueryRowContext(ctx, `
		SELECT m.id_mahasiswa, m.nrp, m.nama_mahasiswa, m.email, m.role, m.status,
			COALESCE(w.id_wallet, ''), COALESCE(w.saldo, 0.00), a.last_login_at,
			m.created_at, m.updated_at
		FROM mahasiswa m
		LEFT JOIN wallet w ON w.mahasiswa_id = m.id_mahasiswa
		LEFT JOIN mahasiswa_auth a ON a.mahasiswa_id = m.id_mahasiswa
		WHERE m.id_mahasiswa = ?
	`, id))
}

func (r *AdminRepository) CreateMahasiswa(ctx context.Context, input AdminMahasiswaInput) (*models.AdminMahasiswa, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		INSERT INTO mahasiswa (nrp, nama_mahasiswa, email, role, status)
		VALUES (?, ?, ?, ?, ?)
	`, input.NRP, input.NamaMahasiswa, input.Email, input.Role, input.Status)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO mahasiswa_auth (mahasiswa_id, password_hash) VALUES (?, ?)
	`, id, input.PasswordHash); err != nil {
		return nil, err
	}
	if input.Role == "admin" {
		if _, err := tx.ExecContext(ctx, `UPDATE wallet SET jenis_wallet = 'ADMIN' WHERE mahasiswa_id = ?`, id); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetMahasiswa(context.Background(), uint64(id))
}

func (r *AdminRepository) UpdateMahasiswa(ctx context.Context, id uint64, input AdminMahasiswaInput) (*models.AdminMahasiswa, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `
		UPDATE mahasiswa
		SET nrp = ?, nama_mahasiswa = ?, email = ?, role = ?, status = ?
		WHERE id_mahasiswa = ?
	`, input.NRP, input.NamaMahasiswa, input.Email, input.Role, input.Status, id)
	if err != nil {
		return nil, err
	}
	if err := requireAffected(result); err != nil {
		return nil, err
	}
	if input.PasswordHash != "" {
		if _, err := tx.ExecContext(ctx, `UPDATE mahasiswa_auth SET password_hash = ? WHERE mahasiswa_id = ?`, input.PasswordHash, id); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetMahasiswa(context.Background(), id)
}

func (r *AdminRepository) DeleteMahasiswa(ctx context.Context, id uint64) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	result, err := r.db.ExecContext(ctx, `DELETE FROM mahasiswa WHERE id_mahasiswa = ?`, id)
	if err != nil {
		return err
	}
	return requireAffected(result)
}

func (r *AdminRepository) ListAuthRecords(ctx context.Context, params AdminListParams) (*models.PaginatedResult, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	where := " WHERE 1=1"
	args := make([]any, 0)
	if params.Search != "" {
		where += " AND (m.nrp LIKE ? OR m.nama_mahasiswa LIKE ? OR m.email LIKE ? OR a.password_hash LIKE ? OR a.pin_hash LIKE ?)"
		term := "%" + params.Search + "%"
		args = append(args, term, term, term, term, term)
	}
	joins := " FROM mahasiswa_auth a INNER JOIN mahasiswa m ON m.id_mahasiswa = a.mahasiswa_id"
	total, err := r.count(ctx, "SELECT COUNT(*)"+joins+where, args...)
	if err != nil {
		return nil, err
	}
	sort := safeSort(params.Sort, map[string]string{
		"id": "a.id_auth", "student": "m.nama_mahasiswa", "nrp": "m.nrp",
		"email": "m.email", "last_login_at": "a.last_login_at", "created_at": "a.created_at",
	}, "a.created_at")
	query := `SELECT a.id_auth, a.mahasiswa_id, m.nama_mahasiswa, m.nrp, m.email,
		a.password_hash, a.pin_hash, a.last_login_at, a.created_at` + joins + where +
		" ORDER BY " + sort + " " + safeOrder(params.Order) + " LIMIT ? OFFSET ?"
	rows, err := r.db.QueryContext(ctx, query, append(append([]any{}, args...), params.PerPage, offset(params))...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]models.AdminAuthRecord, 0)
	for rows.Next() {
		item, err := scanAdminAuthRecord(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return paginated(items, params, total), nil
}

func (r *AdminRepository) GetAuthRecord(ctx context.Context, id uint64) (*models.AdminAuthRecord, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	return scanAdminAuthRecord(r.db.QueryRowContext(ctx, `
		SELECT a.id_auth, a.mahasiswa_id, m.nama_mahasiswa, m.nrp, m.email,
			a.password_hash, a.pin_hash, a.last_login_at, a.created_at
		FROM mahasiswa_auth a
		INNER JOIN mahasiswa m ON m.id_mahasiswa = a.mahasiswa_id
		WHERE a.id_auth = ?
	`, id))
}

func (r *AdminRepository) ListWallets(ctx context.Context, params AdminListParams) (*models.PaginatedResult, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	where := " WHERE 1=1"
	args := make([]any, 0)
	if params.Search != "" {
		where += " AND (w.id_wallet LIKE ? OR m.nrp LIKE ? OR m.nama_mahasiswa LIKE ?)"
		term := "%" + params.Search + "%"
		args = append(args, term, term, term)
	}
	if kind := params.Filters["jenis_wallet"]; kind != "" {
		where += " AND w.jenis_wallet = ?"
		args = append(args, kind)
	}
	total, err := r.count(ctx, "SELECT COUNT(*) FROM wallet w INNER JOIN mahasiswa m ON m.id_mahasiswa = w.mahasiswa_id"+where, args...)
	if err != nil {
		return nil, err
	}
	sort := safeSort(params.Sort, map[string]string{
		"id": "w.id_wallet", "student": "m.nama_mahasiswa", "type": "w.jenis_wallet",
		"balance": "w.saldo", "created_at": "w.created_at",
	}, "w.created_at")
	query := `SELECT w.id_wallet, w.mahasiswa_id, m.nama_mahasiswa, m.nrp, w.jenis_wallet,
		w.saldo, w.created_at, w.updated_at
		FROM wallet w INNER JOIN mahasiswa m ON m.id_mahasiswa = w.mahasiswa_id` + where +
		" ORDER BY " + sort + " " + safeOrder(params.Order) + " LIMIT ? OFFSET ?"
	rows, err := r.db.QueryContext(ctx, query, append(append([]any{}, args...), params.PerPage, offset(params))...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]models.AdminWallet, 0)
	for rows.Next() {
		item, err := scanAdminWallet(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return paginated(items, params, total), nil
}

func (r *AdminRepository) GetWallet(ctx context.Context, id string) (*models.AdminWallet, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	return scanAdminWallet(r.db.QueryRowContext(ctx, `
		SELECT w.id_wallet, w.mahasiswa_id, m.nama_mahasiswa, m.nrp, w.jenis_wallet,
			w.saldo, w.created_at, w.updated_at
		FROM wallet w INNER JOIN mahasiswa m ON m.id_mahasiswa = w.mahasiswa_id
		WHERE w.id_wallet = ?
	`, id))
}

func (r *AdminRepository) UpdateWalletType(ctx context.Context, id, walletType string) (*models.AdminWallet, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	result, err := r.db.ExecContext(ctx, `UPDATE wallet SET jenis_wallet = ? WHERE id_wallet = ?`, walletType, id)
	if err != nil {
		return nil, err
	}
	if err := requireAffected(result); err != nil {
		return nil, err
	}
	return r.GetWallet(context.Background(), id)
}

func (r *AdminRepository) ListBanks(ctx context.Context, params AdminListParams) (*models.PaginatedResult, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	where := " WHERE 1=1"
	args := make([]any, 0)
	if params.Search != "" {
		where += " AND (b.nama_bank LIKE ? OR b.kode_bank LIKE ?)"
		term := "%" + params.Search + "%"
		args = append(args, term, term)
	}
	if active := params.Filters["is_active"]; active != "" {
		where += " AND b.is_active = ?"
		args = append(args, active == "true")
	}
	total, err := r.count(ctx, "SELECT COUNT(*) FROM bank b"+where, args...)
	if err != nil {
		return nil, err
	}
	sort := safeSort(params.Sort, map[string]string{
		"id": "b.id_bank", "name": "b.nama_bank", "code": "b.kode_bank",
		"fee": "b.biaya_admin", "active": "b.is_active", "created_at": "b.created_at",
	}, "b.created_at")
	query := `SELECT b.id_bank, b.nama_bank, b.kode_bank, b.biaya_admin, b.is_active, b.created_at
		FROM bank b` + where + " ORDER BY " + sort + " " + safeOrder(params.Order) + " LIMIT ? OFFSET ?"
	rows, err := r.db.QueryContext(ctx, query, append(append([]any{}, args...), params.PerPage, offset(params))...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]models.AdminBank, 0)
	for rows.Next() {
		item, err := scanAdminBank(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return paginated(items, params, total), nil
}

func (r *AdminRepository) GetBank(ctx context.Context, id uint64) (*models.AdminBank, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	return scanAdminBank(r.db.QueryRowContext(ctx, `
		SELECT id_bank, nama_bank, kode_bank, biaya_admin, is_active, created_at
		FROM bank WHERE id_bank = ?
	`, id))
}

func (r *AdminRepository) CreateBank(ctx context.Context, input AdminBankInput) (*models.AdminBank, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO bank (nama_bank, kode_bank, biaya_admin, is_active) VALUES (?, ?, ?, ?)
	`, input.NamaBank, input.KodeBank, input.BiayaAdmin, input.IsActive)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.GetBank(context.Background(), uint64(id))
}

func (r *AdminRepository) UpdateBank(ctx context.Context, id uint64, input AdminBankInput) (*models.AdminBank, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	result, err := r.db.ExecContext(ctx, `
		UPDATE bank SET nama_bank = ?, kode_bank = ?, biaya_admin = ?, is_active = ? WHERE id_bank = ?
	`, input.NamaBank, input.KodeBank, input.BiayaAdmin, input.IsActive, id)
	if err != nil {
		return nil, err
	}
	if err := requireAffected(result); err != nil {
		return nil, err
	}
	return r.GetBank(context.Background(), id)
}

func (r *AdminRepository) DeleteBank(ctx context.Context, id uint64) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	result, err := r.db.ExecContext(ctx, `DELETE FROM bank WHERE id_bank = ?`, id)
	if err != nil {
		return err
	}
	return requireAffected(result)
}

func (r *AdminRepository) ListRekening(ctx context.Context, params AdminListParams) (*models.PaginatedResult, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	where := " WHERE 1=1"
	args := make([]any, 0)
	if params.Search != "" {
		where += " AND (r.no_rekening LIKE ? OR r.nama_pemilik LIKE ? OR m.nrp LIKE ? OR m.nama_mahasiswa LIKE ? OR b.nama_bank LIKE ?)"
		term := "%" + params.Search + "%"
		args = append(args, term, term, term, term, term)
	}
	if active := params.Filters["is_active"]; active != "" {
		where += " AND r.is_active = ?"
		args = append(args, active == "true")
	}
	if bankID := params.Filters["bank_id"]; bankID != "" {
		where += " AND r.bank_id_bank = ?"
		args = append(args, bankID)
	}
	joins := ` FROM rekening_mahasiswa r
		INNER JOIN mahasiswa m ON m.id_mahasiswa = r.mahasiswa_id
		INNER JOIN bank b ON b.id_bank = r.bank_id_bank`
	total, err := r.count(ctx, "SELECT COUNT(*)"+joins+where, args...)
	if err != nil {
		return nil, err
	}
	sort := safeSort(params.Sort, map[string]string{
		"id": "r.id_rekening", "student": "m.nama_mahasiswa", "bank": "b.nama_bank",
		"account_number": "r.no_rekening", "owner": "r.nama_pemilik", "active": "r.is_active", "created_at": "r.created_at",
	}, "r.created_at")
	query := `SELECT r.id_rekening, r.mahasiswa_id, m.nama_mahasiswa, m.nrp, r.bank_id_bank,
		b.nama_bank, r.no_rekening, r.nama_pemilik, r.is_active, r.created_at` + joins + where +
		" ORDER BY " + sort + " " + safeOrder(params.Order) + " LIMIT ? OFFSET ?"
	rows, err := r.db.QueryContext(ctx, query, append(append([]any{}, args...), params.PerPage, offset(params))...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]models.AdminRekening, 0)
	for rows.Next() {
		item, err := scanAdminRekening(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return paginated(items, params, total), nil
}

func (r *AdminRepository) GetRekening(ctx context.Context, id uint64) (*models.AdminRekening, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	return scanAdminRekening(r.db.QueryRowContext(ctx, `
		SELECT r.id_rekening, r.mahasiswa_id, m.nama_mahasiswa, m.nrp, r.bank_id_bank,
			b.nama_bank, r.no_rekening, r.nama_pemilik, r.is_active, r.created_at
		FROM rekening_mahasiswa r
		INNER JOIN mahasiswa m ON m.id_mahasiswa = r.mahasiswa_id
		INNER JOIN bank b ON b.id_bank = r.bank_id_bank
		WHERE r.id_rekening = ?
	`, id))
}

func (r *AdminRepository) CreateRekening(ctx context.Context, input AdminRekeningInput) (*models.AdminRekening, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO rekening_mahasiswa (mahasiswa_id, bank_id_bank, no_rekening, nama_pemilik, is_active)
		VALUES (?, ?, ?, ?, ?)
	`, input.MahasiswaID, input.BankID, input.NoRekening, input.NamaPemilik, input.IsActive)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.GetRekening(context.Background(), uint64(id))
}

func (r *AdminRepository) UpdateRekening(ctx context.Context, id uint64, input AdminRekeningInput) (*models.AdminRekening, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	result, err := r.db.ExecContext(ctx, `
		UPDATE rekening_mahasiswa
		SET mahasiswa_id = ?, bank_id_bank = ?, no_rekening = ?, nama_pemilik = ?, is_active = ?
		WHERE id_rekening = ?
	`, input.MahasiswaID, input.BankID, input.NoRekening, input.NamaPemilik, input.IsActive, id)
	if err != nil {
		return nil, err
	}
	if err := requireAffected(result); err != nil {
		return nil, err
	}
	return r.GetRekening(context.Background(), id)
}

func (r *AdminRepository) DeleteRekening(ctx context.Context, id uint64) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	result, err := r.db.ExecContext(ctx, `DELETE FROM rekening_mahasiswa WHERE id_rekening = ?`, id)
	if err != nil {
		return err
	}
	return requireAffected(result)
}

func (r *AdminRepository) ListMerchants(ctx context.Context, params AdminListParams) (*models.PaginatedResult, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	where := " WHERE 1=1"
	args := make([]any, 0)
	if params.Search != "" {
		where += " AND (m.id_merchant LIKE ? OR m.nama_merchant LIKE ? OR m.kategori LIKE ?)"
		term := "%" + params.Search + "%"
		args = append(args, term, term, term)
	}
	if status := params.Filters["status"]; status != "" {
		where += " AND m.status = ?"
		args = append(args, status)
	}
	total, err := r.count(ctx, "SELECT COUNT(*) FROM merchant m"+where, args...)
	if err != nil {
		return nil, err
	}
	sort := safeSort(params.Sort, map[string]string{
		"id": "m.id_merchant", "name": "m.nama_merchant", "category": "m.kategori",
		"balance": "m.saldo_merchant", "status": "m.status", "created_at": "m.created_at",
	}, "m.created_at")
	query := `SELECT m.id_merchant, m.nama_merchant, m.kategori, m.saldo_merchant, m.status,
		m.created_at, m.updated_at FROM merchant m` + where +
		" ORDER BY " + sort + " " + safeOrder(params.Order) + " LIMIT ? OFFSET ?"
	rows, err := r.db.QueryContext(ctx, query, append(append([]any{}, args...), params.PerPage, offset(params))...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]models.AdminMerchant, 0)
	for rows.Next() {
		item, err := scanAdminMerchant(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return paginated(items, params, total), nil
}

func (r *AdminRepository) GetMerchant(ctx context.Context, id string) (*models.AdminMerchant, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	return scanAdminMerchant(r.db.QueryRowContext(ctx, `
		SELECT id_merchant, nama_merchant, kategori, saldo_merchant, status, created_at, updated_at
		FROM merchant WHERE id_merchant = ?
	`, id))
}

func (r *AdminRepository) CreateMerchant(ctx context.Context, input AdminMerchantInput) (*models.AdminMerchant, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO merchant (id_merchant, nama_merchant, kategori, status) VALUES (?, ?, ?, ?)
	`, input.IDMerchant, input.NamaMerchant, input.Kategori, input.Status)
	if err != nil {
		return nil, err
	}
	return r.GetMerchant(context.Background(), input.IDMerchant)
}

func (r *AdminRepository) UpdateMerchant(ctx context.Context, id string, input AdminMerchantInput) (*models.AdminMerchant, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	result, err := r.db.ExecContext(ctx, `
		UPDATE merchant SET nama_merchant = ?, kategori = ?, status = ? WHERE id_merchant = ?
	`, input.NamaMerchant, input.Kategori, input.Status, id)
	if err != nil {
		return nil, err
	}
	if err := requireAffected(result); err != nil {
		return nil, err
	}
	return r.GetMerchant(context.Background(), id)
}

func (r *AdminRepository) DeleteMerchant(ctx context.Context, id string) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	result, err := r.db.ExecContext(ctx, `DELETE FROM merchant WHERE id_merchant = ?`, id)
	if err != nil {
		return err
	}
	return requireAffected(result)
}

func (r *AdminRepository) ListTransactionsPaged(ctx context.Context, params AdminListParams) (*models.PaginatedResult, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	where := " WHERE 1=1"
	args := make([]any, 0)
	if params.Search != "" {
		where += " AND (t.kode_transaksi LIKE ? OR t.keterangan LIKE ? OR w.id_wallet LIKE ? OR m.nrp LIKE ? OR m.nama_mahasiswa LIKE ? OR b.nama_bank LIKE ? OR mc.nama_merchant LIKE ?)"
		term := "%" + params.Search + "%"
		args = append(args, term, term, term, term, term, term, term)
	}
	if transactionType := params.Filters["type"]; transactionType != "" {
		where += " AND t.jenis_transaksi = ?"
		args = append(args, transactionType)
	}
	if status := params.Filters["status"]; status != "" {
		where += " AND t.status = ?"
		args = append(args, status)
	}
	joins := ` FROM transaksi t
		INNER JOIN wallet w ON w.id_wallet = t.wallet_id_wallet
		INNER JOIN mahasiswa m ON m.id_mahasiswa = w.mahasiswa_id
		LEFT JOIN bank b ON b.id_bank = t.bank_id_bank
		LEFT JOIN merchant mc ON mc.id_merchant = t.merchant_id`
	total, err := r.count(ctx, "SELECT COUNT(*)"+joins+where, args...)
	if err != nil {
		return nil, err
	}
	sort := safeSort(params.Sort, map[string]string{
		"id": "t.id_transaksi", "code": "t.kode_transaksi", "type": "t.jenis_transaksi",
		"amount": "t.nominal", "status": "t.status", "student": "m.nama_mahasiswa", "date": "t.waktu",
	}, "t.waktu")
	query := `SELECT t.id_transaksi, t.kode_transaksi, t.jenis_transaksi, t.nominal, t.status,
		t.waktu, t.bank_id_bank, COALESCE(b.nama_bank, ''), COALESCE(t.merchant_id, ''),
		COALESCE(mc.nama_merchant, ''), t.wallet_id_wallet, m.id_mahasiswa, m.nama_mahasiswa,
		m.nrp, COALESCE(t.keterangan, ''), t.created_at` + joins + where +
		" ORDER BY " + sort + " " + safeOrder(params.Order) + " LIMIT ? OFFSET ?"
	rows, err := r.db.QueryContext(ctx, query, append(append([]any{}, args...), params.PerPage, offset(params))...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]models.AdminTransaction, 0)
	for rows.Next() {
		item, err := scanAdminTransaction(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return paginated(items, params, total), nil
}

func (r *AdminRepository) GetTransaction(ctx context.Context, id uint64) (*models.AdminTransaction, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	return scanAdminTransaction(r.db.QueryRowContext(ctx, `
		SELECT t.id_transaksi, t.kode_transaksi, t.jenis_transaksi, t.nominal, t.status,
			t.waktu, t.bank_id_bank, COALESCE(b.nama_bank, ''), COALESCE(t.merchant_id, ''),
			COALESCE(mc.nama_merchant, ''), t.wallet_id_wallet, m.id_mahasiswa, m.nama_mahasiswa,
			m.nrp, COALESCE(t.keterangan, ''), t.created_at
		FROM transaksi t
		INNER JOIN wallet w ON w.id_wallet = t.wallet_id_wallet
		INNER JOIN mahasiswa m ON m.id_mahasiswa = w.mahasiswa_id
		LEFT JOIN bank b ON b.id_bank = t.bank_id_bank
		LEFT JOIN merchant mc ON mc.id_merchant = t.merchant_id
		WHERE t.id_transaksi = ?
	`, id))
}

func (r *AdminRepository) ListAuditLogsPaged(ctx context.Context, params AdminListParams) (*models.PaginatedResult, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	where := " WHERE 1=1"
	args := make([]any, 0)
	if params.Search != "" {
		where += " AND (a.action LIKE ? OR a.description LIKE ? OR t.kode_transaksi LIKE ?)"
		term := "%" + params.Search + "%"
		args = append(args, term, term, term)
	}
	if action := params.Filters["action"]; action != "" {
		where += " AND a.action = ?"
		args = append(args, action)
	}
	joins := " FROM audit_logs a LEFT JOIN transaksi t ON t.id_transaksi = a.transaksi_id"
	total, err := r.count(ctx, "SELECT COUNT(*)"+joins+where, args...)
	if err != nil {
		return nil, err
	}
	sort := safeSort(params.Sort, map[string]string{
		"id": "a.id_audit", "action": "a.action", "transaction": "t.kode_transaksi", "created_at": "a.created_at",
	}, "a.created_at")
	query := `SELECT a.id_audit, a.transaksi_id, COALESCE(t.kode_transaksi, ''), a.action,
		a.description, a.created_at` + joins + where + " ORDER BY " + sort + " " + safeOrder(params.Order) + " LIMIT ? OFFSET ?"
	rows, err := r.db.QueryContext(ctx, query, append(append([]any{}, args...), params.PerPage, offset(params))...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]models.AdminAuditLog, 0)
	for rows.Next() {
		var item models.AdminAuditLog
		var transactionID sql.NullInt64
		if err := rows.Scan(&item.IDAudit, &transactionID, &item.TransactionCode, &item.Action, &item.Description, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.TransaksiID = nullableUint64(transactionID)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return paginated(items, params, total), nil
}

func (r *AdminRepository) ListDailyReportsPaged(ctx context.Context, params AdminListParams) (*models.PaginatedResult, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	where := ""
	args := make([]any, 0)
	if params.Search != "" {
		where = " WHERE DATE_FORMAT(tanggal, '%Y-%m-%d') LIKE ?"
		args = append(args, "%"+params.Search+"%")
	}
	total, err := r.count(ctx, "SELECT COUNT(*) FROM v_laporan_transaksi_harian"+where, args...)
	if err != nil {
		return nil, err
	}
	sort := safeSort(params.Sort, map[string]string{
		"date": "tanggal", "transactions": "total_transaksi", "amount": "total_nominal",
		"topup": "total_topup", "payment": "total_payment",
	}, "tanggal")
	query := `SELECT DATE_FORMAT(tanggal, '%Y-%m-%d'), total_transaksi, total_nominal,
		total_topup, total_payment, total_transaksi_success
		FROM v_laporan_transaksi_harian` + where + " ORDER BY " + sort + " " + safeOrder(params.Order) + " LIMIT ? OFFSET ?"
	rows, err := r.db.QueryContext(ctx, query, append(append([]any{}, args...), params.PerPage, offset(params))...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]models.DailyReport, 0)
	for rows.Next() {
		var item models.DailyReport
		if err := rows.Scan(&item.Tanggal, &item.TotalTransaksi, &item.TotalNominal, &item.TotalTopup, &item.TotalPayment, &item.TotalTransaksiOK); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return paginated(items, params, total), nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanAdminMahasiswa(row scanner) (*models.AdminMahasiswa, error) {
	var item models.AdminMahasiswa
	var lastLogin sql.NullTime
	if err := row.Scan(&item.ID, &item.NRP, &item.NamaMahasiswa, &item.Email, &item.Role, &item.Status,
		&item.WalletID, &item.WalletBalance, &lastLogin, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	if lastLogin.Valid {
		item.LastLoginAt = &lastLogin.Time
	}
	return &item, nil
}

func scanAdminAuthRecord(row scanner) (*models.AdminAuthRecord, error) {
	var item models.AdminAuthRecord
	var pinHash sql.NullString
	var lastLogin sql.NullTime
	if err := row.Scan(&item.IDAuth, &item.MahasiswaID, &item.MahasiswaName, &item.NRP,
		&item.Email, &item.PasswordHash, &pinHash, &lastLogin, &item.CreatedAt); err != nil {
		return nil, err
	}
	if pinHash.Valid {
		item.PINHash = &pinHash.String
	}
	if lastLogin.Valid {
		item.LastLoginAt = &lastLogin.Time
	}
	return &item, nil
}

func scanAdminWallet(row scanner) (*models.AdminWallet, error) {
	var item models.AdminWallet
	if err := row.Scan(&item.IDWallet, &item.MahasiswaID, &item.MahasiswaName, &item.NRP,
		&item.JenisWallet, &item.Saldo, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	return &item, nil
}

func scanAdminBank(row scanner) (*models.AdminBank, error) {
	var item models.AdminBank
	if err := row.Scan(&item.IDBank, &item.NamaBank, &item.KodeBank, &item.BiayaAdmin, &item.IsActive, &item.CreatedAt); err != nil {
		return nil, err
	}
	return &item, nil
}

func scanAdminRekening(row scanner) (*models.AdminRekening, error) {
	var item models.AdminRekening
	if err := row.Scan(&item.IDRekening, &item.MahasiswaID, &item.MahasiswaName, &item.NRP,
		&item.BankID, &item.BankName, &item.NoRekening, &item.NamaPemilik, &item.IsActive, &item.CreatedAt); err != nil {
		return nil, err
	}
	return &item, nil
}

func scanAdminMerchant(row scanner) (*models.AdminMerchant, error) {
	var item models.AdminMerchant
	if err := row.Scan(&item.IDMerchant, &item.NamaMerchant, &item.Kategori, &item.SaldoMerchant,
		&item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	return &item, nil
}

func scanAdminTransaction(row scanner) (*models.AdminTransaction, error) {
	var item models.AdminTransaction
	var bankID sql.NullInt64
	if err := row.Scan(&item.IDTransaksi, &item.KodeTransaksi, &item.JenisTransaksi, &item.Nominal,
		&item.Status, &item.Waktu, &bankID, &item.BankName, &item.MerchantID, &item.MerchantName,
		&item.WalletID, &item.MahasiswaID, &item.MahasiswaName, &item.NRP, &item.Keterangan, &item.CreatedAt); err != nil {
		return nil, err
	}
	item.BankID = nullableUint64(bankID)
	return &item, nil
}

func (r *AdminRepository) count(ctx context.Context, query string, args ...any) (int64, error) {
	var total int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&total)
	return total, err
}

func paginated(items any, params AdminListParams, total int64) *models.PaginatedResult {
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(params.PerPage) - 1) / int64(params.PerPage))
	}
	return &models.PaginatedResult{Items: items, Pagination: models.Pagination{
		Page: params.Page, PerPage: params.PerPage, Total: total, TotalPages: totalPages,
	}}
}

func offset(params AdminListParams) int {
	return (params.Page - 1) * params.PerPage
}

func safeSort(requested string, allowed map[string]string, fallback string) string {
	if column, ok := allowed[requested]; ok {
		return column
	}
	return fallback
}

func safeOrder(order string) string {
	if strings.EqualFold(order, "asc") {
		return "ASC"
	}
	return "DESC"
}

func requireAffected(result sql.Result) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func nullableUint64(value sql.NullInt64) *uint64 {
	if !value.Valid {
		return nil
	}
	converted := uint64(value.Int64)
	return &converted
}
