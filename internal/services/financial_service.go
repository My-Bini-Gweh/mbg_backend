package services

import (
	"context"

	"mbg-backend/internal/models"
	"mbg-backend/internal/repositories"
	"mbg-backend/internal/utils"
)

type FinancialService struct {
	financialRepo *repositories.FinancialRepository
	mahasiswaRepo *repositories.MahasiswaRepository
}

type TopupInput struct {
	MahasiswaID uint64
	WalletID    string
	BankID      uint64
	Nominal     float64
}

type PaymentInput struct {
	MahasiswaID uint64
	WalletID    string
	MerchantID  string
	Nominal     float64
}

func NewFinancialService(financialRepo *repositories.FinancialRepository, mahasiswaRepo *repositories.MahasiswaRepository) *FinancialService {
	return &FinancialService{financialRepo: financialRepo, mahasiswaRepo: mahasiswaRepo}
}

func (s *FinancialService) Topup(ctx context.Context, input TopupInput) (*models.Wallet, error) {
	allowed, err := s.mahasiswaRepo.WalletBelongsToMahasiswa(ctx, input.WalletID, input.MahasiswaID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrWalletForbidden
	}

	if err := s.financialRepo.Topup(ctx, input.WalletID, input.BankID, input.Nominal); err != nil {
		return nil, BusinessError{Message: utils.CleanDatabaseError(err)}
	}

	return s.mahasiswaRepo.GetWalletByMahasiswaID(ctx, input.MahasiswaID)
}

func (s *FinancialService) PayMerchant(ctx context.Context, input PaymentInput) (*models.Wallet, error) {
	allowed, err := s.mahasiswaRepo.WalletBelongsToMahasiswa(ctx, input.WalletID, input.MahasiswaID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrWalletForbidden
	}

	if err := s.financialRepo.PayMerchant(ctx, input.WalletID, input.MerchantID, input.Nominal); err != nil {
		return nil, BusinessError{Message: utils.CleanDatabaseError(err)}
	}

	return s.mahasiswaRepo.GetWalletByMahasiswaID(ctx, input.MahasiswaID)
}
