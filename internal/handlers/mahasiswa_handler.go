package handlers

import (
	"database/sql"
	"errors"
	"net/http"

	"mbg-backend/internal/middleware"
	"mbg-backend/internal/repositories"
	"mbg-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type MahasiswaHandler struct {
	mahasiswaRepo *repositories.MahasiswaRepository
}

func NewMahasiswaHandler(mahasiswaRepo *repositories.MahasiswaRepository) *MahasiswaHandler {
	return &MahasiswaHandler{mahasiswaRepo: mahasiswaRepo}
}

func (h *MahasiswaHandler) Profile(c *gin.Context) {
	mahasiswaID, _, ok := middleware.CurrentUser(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "Token tidak valid")
		return
	}

	mahasiswa, err := h.mahasiswaRepo.GetByID(c.Request.Context(), mahasiswaID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.Error(c, http.StatusNotFound, "Mahasiswa tidak ditemukan")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil profil mahasiswa")
		return
	}

	utils.Success(c, http.StatusOK, "Profil mahasiswa berhasil diambil", mahasiswa)
}

func (h *MahasiswaHandler) Wallet(c *gin.Context) {
	mahasiswaID, _, ok := middleware.CurrentUser(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "Token tidak valid")
		return
	}

	wallet, err := h.mahasiswaRepo.GetWalletByMahasiswaID(c.Request.Context(), mahasiswaID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.Error(c, http.StatusNotFound, "Wallet tidak ditemukan")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil wallet")
		return
	}

	utils.Success(c, http.StatusOK, "Wallet berhasil diambil", wallet)
}

func (h *MahasiswaHandler) Transactions(c *gin.Context) {
	mahasiswaID, _, ok := middleware.CurrentUser(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "Token tidak valid")
		return
	}

	transactions, err := h.mahasiswaRepo.ListTransactionsByMahasiswa(c.Request.Context(), mahasiswaID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil riwayat transaksi")
		return
	}

	utils.Success(c, http.StatusOK, "Riwayat transaksi berhasil diambil", transactions)
}
