package repositories

import (
	"context"
	"database/sql"
)

type FinancialRepository struct {
	db *sql.DB
}

func NewFinancialRepository(db *sql.DB) *FinancialRepository {
	return &FinancialRepository{db: db}
}

func (r *FinancialRepository) Topup(ctx context.Context, walletID string, bankID uint64, nominal float64) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	_, err := r.db.ExecContext(ctx, "CALL sp_topup_wallet(?, ?, ?)", walletID, bankID, nominal)
	return err
}

func (r *FinancialRepository) PayMerchant(ctx context.Context, walletID, merchantID string, nominal float64) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	_, err := r.db.ExecContext(ctx, "CALL sp_bayar_merchant(?, ?, ?)", walletID, merchantID, nominal)
	return err
}
