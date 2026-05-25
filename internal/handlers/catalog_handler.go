package handlers

import (
	"database/sql"
	"errors"
	"net/http"

	"mbg-backend/internal/repositories"
	"mbg-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type CatalogHandler struct {
	catalogRepo *repositories.CatalogRepository
}

func NewCatalogHandler(catalogRepo *repositories.CatalogRepository) *CatalogHandler {
	return &CatalogHandler{catalogRepo: catalogRepo}
}

func (h *CatalogHandler) Banks(c *gin.Context) {
	banks, err := h.catalogRepo.ListBanks(c.Request.Context())
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil data bank")
		return
	}

	utils.Success(c, http.StatusOK, "Data bank berhasil diambil", banks)
}

func (h *CatalogHandler) Merchants(c *gin.Context) {
	merchants, err := h.catalogRepo.ListMerchants(c.Request.Context())
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil data merchant")
		return
	}

	utils.Success(c, http.StatusOK, "Data merchant berhasil diambil", merchants)
}

func (h *CatalogHandler) MerchantDetail(c *gin.Context) {
	merchant, err := h.catalogRepo.GetMerchantByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.Error(c, http.StatusNotFound, "Merchant tidak ditemukan")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil detail merchant")
		return
	}

	utils.Success(c, http.StatusOK, "Detail merchant berhasil diambil", merchant)
}
