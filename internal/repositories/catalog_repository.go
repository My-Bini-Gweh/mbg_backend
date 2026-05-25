package repositories

import (
	"context"
	"database/sql"

	"mbg-backend/internal/models"
)

type CatalogRepository struct {
	db *sql.DB
}

func NewCatalogRepository(db *sql.DB) *CatalogRepository {
	return &CatalogRepository{db: db}
}

func (r *CatalogRepository) ListBanks(ctx context.Context) ([]models.Bank, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, `
		SELECT id_bank, nama_bank, kode_bank, biaya_admin, is_active
		FROM bank
		WHERE is_active = TRUE
		ORDER BY nama_bank ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	banks := make([]models.Bank, 0)
	for rows.Next() {
		var bank models.Bank
		if err := rows.Scan(&bank.IDBank, &bank.NamaBank, &bank.KodeBank, &bank.BiayaAdmin, &bank.IsActive); err != nil {
			return nil, err
		}
		banks = append(banks, bank)
	}

	return banks, rows.Err()
}

func (r *CatalogRepository) ListMerchants(ctx context.Context) ([]models.Merchant, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, `
		SELECT id_merchant, nama_merchant, kategori, saldo_merchant, status
		FROM merchant
		WHERE status = 'ACTIVE'
		ORDER BY nama_merchant ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	merchants := make([]models.Merchant, 0)
	for rows.Next() {
		var merchant models.Merchant
		if err := rows.Scan(
			&merchant.IDMerchant,
			&merchant.NamaMerchant,
			&merchant.Kategori,
			&merchant.SaldoMerchant,
			&merchant.Status,
		); err != nil {
			return nil, err
		}
		merchants = append(merchants, merchant)
	}

	return merchants, rows.Err()
}

func (r *CatalogRepository) GetMerchantByID(ctx context.Context, id string) (*models.Merchant, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	var merchant models.Merchant
	err := r.db.QueryRowContext(ctx, `
		SELECT id_merchant, nama_merchant, kategori, saldo_merchant, status
		FROM merchant
		WHERE id_merchant = ?
	`, id).Scan(
		&merchant.IDMerchant,
		&merchant.NamaMerchant,
		&merchant.Kategori,
		&merchant.SaldoMerchant,
		&merchant.Status,
	)
	if err != nil {
		return nil, err
	}

	return &merchant, nil
}
