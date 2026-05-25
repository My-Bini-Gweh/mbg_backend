package handlers

import (
	"errors"
	"net/http"
	"strings"

	"mbg-backend/internal/middleware"
	"mbg-backend/internal/services"
	"mbg-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type FinancialHandler struct {
	financialService *services.FinancialService
}

type topupRequest struct {
	WalletID string  `json:"wallet_id"`
	BankID   uint64  `json:"bank_id"`
	Nominal  float64 `json:"nominal"`
}

type paymentRequest struct {
	WalletID   string  `json:"wallet_id"`
	MerchantID string  `json:"merchant_id"`
	Nominal    float64 `json:"nominal"`
}

func NewFinancialHandler(financialService *services.FinancialService) *FinancialHandler {
	return &FinancialHandler{financialService: financialService}
}

func (h *FinancialHandler) Topup(c *gin.Context) {
	mahasiswaID, _, ok := middleware.CurrentUser(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "Token tidak valid")
		return
	}

	var request topupRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "Request body tidak valid")
		return
	}

	if strings.TrimSpace(request.WalletID) == "" || request.BankID == 0 || request.Nominal <= 0 {
		utils.Error(c, http.StatusBadRequest, "wallet_id, bank_id, dan nominal wajib valid")
		return
	}

	wallet, err := h.financialService.Topup(c.Request.Context(), services.TopupInput{
		MahasiswaID: mahasiswaID,
		WalletID:    strings.TrimSpace(request.WalletID),
		BankID:      request.BankID,
		Nominal:     request.Nominal,
	})
	if err != nil {
		if errors.Is(err, services.ErrWalletForbidden) {
			utils.Error(c, http.StatusForbidden, err.Error())
			return
		}
		var businessErr services.BusinessError
		if errors.As(err, &businessErr) {
			utils.Error(c, http.StatusBadRequest, businessErr.Message)
			return
		}
		utils.Error(c, http.StatusInternalServerError, "Gagal memproses top up")
		return
	}

	utils.Success(c, http.StatusOK, "Top up berhasil", gin.H{"wallet": wallet})
}

func (h *FinancialHandler) PayMerchant(c *gin.Context) {
	mahasiswaID, _, ok := middleware.CurrentUser(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "Token tidak valid")
		return
	}

	var request paymentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "Request body tidak valid")
		return
	}

	if strings.TrimSpace(request.WalletID) == "" || strings.TrimSpace(request.MerchantID) == "" || request.Nominal <= 0 {
		utils.Error(c, http.StatusBadRequest, "wallet_id, merchant_id, dan nominal wajib valid")
		return
	}

	wallet, err := h.financialService.PayMerchant(c.Request.Context(), services.PaymentInput{
		MahasiswaID: mahasiswaID,
		WalletID:    strings.TrimSpace(request.WalletID),
		MerchantID:  strings.TrimSpace(request.MerchantID),
		Nominal:     request.Nominal,
	})
	if err != nil {
		if errors.Is(err, services.ErrWalletForbidden) {
			utils.Error(c, http.StatusForbidden, err.Error())
			return
		}
		var businessErr services.BusinessError
		if errors.As(err, &businessErr) {
			utils.Error(c, http.StatusBadRequest, businessErr.Message)
			return
		}
		utils.Error(c, http.StatusInternalServerError, "Gagal memproses pembayaran")
		return
	}

	utils.Success(c, http.StatusOK, "Pembayaran merchant berhasil", gin.H{"wallet": wallet})
}
