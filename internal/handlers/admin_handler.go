package handlers

import (
	"net/http"

	"mbg-backend/internal/repositories"
	"mbg-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminRepo *repositories.AdminRepository
}

func NewAdminHandler(adminRepo *repositories.AdminRepository) *AdminHandler {
	return &AdminHandler{adminRepo: adminRepo}
}

func (h *AdminHandler) Transactions(c *gin.Context) {
	transactions, err := h.adminRepo.ListTransactions(c.Request.Context())
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil transaksi admin")
		return
	}

	utils.Success(c, http.StatusOK, "Data transaksi berhasil diambil", transactions)
}

func (h *AdminHandler) AuditLogs(c *gin.Context) {
	logs, err := h.adminRepo.ListAuditLogs(c.Request.Context())
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil audit logs")
		return
	}

	utils.Success(c, http.StatusOK, "Audit logs berhasil diambil", logs)
}

func (h *AdminHandler) DailyReports(c *gin.Context) {
	reports, err := h.adminRepo.ListDailyReports(c.Request.Context())
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil laporan harian")
		return
	}

	utils.Success(c, http.StatusOK, "Laporan harian berhasil diambil", reports)
}
