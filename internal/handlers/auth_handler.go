package handlers

import (
	"errors"
	"net/http"
	"strings"

	"mbg-backend/internal/middleware"
	"mbg-backend/internal/repositories"
	"mbg-backend/internal/services"
	"mbg-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService   *services.AuthService
	mahasiswaRepo *repositories.MahasiswaRepository
}

type registerRequest struct {
	NRP           string `json:"nrp"`
	NamaMahasiswa string `json:"nama_mahasiswa"`
	Nama          string `json:"nama"`
	Email         string `json:"email"`
	Password      string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewAuthHandler(authService *services.AuthService, mahasiswaRepo *repositories.MahasiswaRepository) *AuthHandler {
	return &AuthHandler{authService: authService, mahasiswaRepo: mahasiswaRepo}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var request registerRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "Request body tidak valid")
		return
	}

	name := strings.TrimSpace(request.NamaMahasiswa)
	if name == "" {
		name = strings.TrimSpace(request.Nama)
	}

	if strings.TrimSpace(request.NRP) == "" || name == "" || strings.TrimSpace(request.Email) == "" || len(request.Password) < 6 {
		utils.Error(c, http.StatusBadRequest, "NRP, nama, email, dan password minimal 6 karakter wajib diisi")
		return
	}

	result, err := h.authService.Register(c.Request.Context(), services.RegisterInput{
		NRP:           strings.TrimSpace(request.NRP),
		NamaMahasiswa: name,
		Email:         strings.TrimSpace(request.Email),
		Password:      request.Password,
	})
	if err != nil {
		if errors.Is(err, services.ErrDuplicateAccount) {
			utils.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		utils.Error(c, http.StatusInternalServerError, "Gagal register mahasiswa")
		return
	}

	utils.Success(c, http.StatusCreated, "Register berhasil", result)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var request loginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "Request body tidak valid")
		return
	}

	if strings.TrimSpace(request.Email) == "" || request.Password == "" {
		utils.Error(c, http.StatusBadRequest, "Email dan password wajib diisi")
		return
	}

	result, err := h.authService.Login(c.Request.Context(), strings.TrimSpace(request.Email), request.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			utils.Error(c, http.StatusUnauthorized, err.Error())
			return
		}
		utils.Error(c, http.StatusInternalServerError, "Gagal login")
		return
	}

	utils.Success(c, http.StatusOK, "Login berhasil", result)
}

func (h *AuthHandler) Me(c *gin.Context) {
	mahasiswaID, _, ok := middleware.CurrentUser(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "Token tidak valid")
		return
	}

	mahasiswa, err := h.mahasiswaRepo.GetByID(c.Request.Context(), mahasiswaID)
	if err != nil {
		utils.Error(c, http.StatusNotFound, "Mahasiswa tidak ditemukan")
		return
	}

	utils.Success(c, http.StatusOK, "Data user aktif berhasil diambil", mahasiswa)
}
