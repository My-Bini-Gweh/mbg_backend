package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"net/mail"
	"strconv"
	"strings"

	"mbg-backend/internal/middleware"
	"mbg-backend/internal/repositories"
	"mbg-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminRepo *repositories.AdminRepository
}

type adminMahasiswaRequest struct {
	NRP           string `json:"nrp"`
	NamaMahasiswa string `json:"nama_mahasiswa"`
	Email         string `json:"email"`
	Role          string `json:"role"`
	Status        string `json:"status"`
	Password      string `json:"password"`
}

type adminWalletRequest struct {
	JenisWallet string `json:"jenis_wallet"`
}

type adminBankRequest struct {
	NamaBank   string `json:"nama_bank"`
	KodeBank   string `json:"kode_bank"`
	BiayaAdmin string `json:"biaya_admin"`
	IsActive   bool   `json:"is_active"`
}

type adminRekeningRequest struct {
	MahasiswaID uint64 `json:"mahasiswa_id"`
	BankID      uint64 `json:"bank_id_bank"`
	NoRekening  string `json:"no_rekening"`
	NamaPemilik string `json:"nama_pemilik"`
	IsActive    bool   `json:"is_active"`
}

type adminMerchantRequest struct {
	IDMerchant   string `json:"id_merchant"`
	NamaMerchant string `json:"nama_merchant"`
	Kategori     string `json:"kategori"`
	Status       string `json:"status"`
}

func NewAdminHandler(adminRepo *repositories.AdminRepository) *AdminHandler {
	return &AdminHandler{adminRepo: adminRepo}
}

func (h *AdminHandler) Summary(c *gin.Context) {
	summary, err := h.adminRepo.Summary(c.Request.Context())
	if err != nil {
		adminError(c, err, "Gagal mengambil ringkasan admin")
		return
	}
	utils.Success(c, http.StatusOK, "Ringkasan admin berhasil diambil", summary)
}

func (h *AdminHandler) Mahasiswa(c *gin.Context) {
	result, err := h.adminRepo.ListMahasiswa(c.Request.Context(), listParams(c))
	if err != nil {
		adminError(c, err, "Gagal mengambil data mahasiswa")
		return
	}
	utils.Success(c, http.StatusOK, "Data mahasiswa berhasil diambil", result)
}

func (h *AdminHandler) MahasiswaDetail(c *gin.Context) {
	id, ok := numericID(c, "id")
	if !ok {
		return
	}
	item, err := h.adminRepo.GetMahasiswa(c.Request.Context(), id)
	if err != nil {
		adminError(c, err, "Gagal mengambil mahasiswa")
		return
	}
	utils.Success(c, http.StatusOK, "Mahasiswa berhasil diambil", item)
}

func (h *AdminHandler) CreateMahasiswa(c *gin.Context) {
	request, ok := bindMahasiswa(c, true)
	if !ok {
		return
	}
	passwordHash, err := utils.HashPassword(request.Password)
	if err != nil {
		adminError(c, err, "Gagal memproses password")
		return
	}
	item, err := h.adminRepo.CreateMahasiswa(c.Request.Context(), mahasiswaInput(request, passwordHash))
	if err != nil {
		adminError(c, err, "Gagal membuat mahasiswa")
		return
	}
	utils.Success(c, http.StatusCreated, "Mahasiswa berhasil dibuat", item)
}

func (h *AdminHandler) UpdateMahasiswa(c *gin.Context) {
	id, ok := numericID(c, "id")
	if !ok {
		return
	}
	request, ok := bindMahasiswa(c, false)
	if !ok {
		return
	}
	currentID, _, authenticated := middleware.CurrentUser(c)
	if !authenticated {
		utils.Error(c, http.StatusUnauthorized, "Token tidak valid")
		return
	}
	if currentID == id && (request.Role != "admin" || request.Status != "ACTIVE") {
		utils.Error(c, http.StatusBadRequest, "Admin aktif tidak dapat menonaktifkan atau menurunkan role akun sendiri")
		return
	}
	passwordHash := ""
	if request.Password != "" {
		var err error
		passwordHash, err = utils.HashPassword(request.Password)
		if err != nil {
			adminError(c, err, "Gagal memproses password")
			return
		}
	}
	item, err := h.adminRepo.UpdateMahasiswa(c.Request.Context(), id, mahasiswaInput(request, passwordHash))
	if err != nil {
		adminError(c, err, "Gagal memperbarui mahasiswa")
		return
	}
	utils.Success(c, http.StatusOK, "Mahasiswa berhasil diperbarui", item)
}

func (h *AdminHandler) DeleteMahasiswa(c *gin.Context) {
	id, ok := numericID(c, "id")
	if !ok {
		return
	}
	currentID, _, authenticated := middleware.CurrentUser(c)
	if !authenticated {
		utils.Error(c, http.StatusUnauthorized, "Token tidak valid")
		return
	}
	if currentID == id {
		utils.Error(c, http.StatusBadRequest, "Admin tidak dapat menghapus akun sendiri")
		return
	}
	if err := h.adminRepo.DeleteMahasiswa(c.Request.Context(), id); err != nil {
		adminError(c, err, "Gagal menghapus mahasiswa")
		return
	}
	utils.Success(c, http.StatusOK, "Mahasiswa berhasil dihapus", nil)
}

func (h *AdminHandler) AuthRecords(c *gin.Context) {
	result, err := h.adminRepo.ListAuthRecords(c.Request.Context(), listParams(c))
	if err != nil {
		adminError(c, err, "Gagal mengambil data autentikasi mahasiswa")
		return
	}
	utils.Success(c, http.StatusOK, "Data autentikasi mahasiswa berhasil diambil", result)
}

func (h *AdminHandler) AuthRecordDetail(c *gin.Context) {
	id, ok := numericID(c, "id")
	if !ok {
		return
	}
	item, err := h.adminRepo.GetAuthRecord(c.Request.Context(), id)
	if err != nil {
		adminError(c, err, "Gagal mengambil data autentikasi mahasiswa")
		return
	}
	utils.Success(c, http.StatusOK, "Data autentikasi mahasiswa berhasil diambil", item)
}

func (h *AdminHandler) Wallets(c *gin.Context) {
	result, err := h.adminRepo.ListWallets(c.Request.Context(), listParams(c))
	if err != nil {
		adminError(c, err, "Gagal mengambil data wallet")
		return
	}
	utils.Success(c, http.StatusOK, "Data wallet berhasil diambil", result)
}

func (h *AdminHandler) WalletDetail(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		utils.Error(c, http.StatusBadRequest, "ID wallet tidak valid")
		return
	}
	item, err := h.adminRepo.GetWallet(c.Request.Context(), id)
	if err != nil {
		adminError(c, err, "Gagal mengambil wallet")
		return
	}
	utils.Success(c, http.StatusOK, "Wallet berhasil diambil", item)
}

func (h *AdminHandler) UpdateWallet(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	var request adminWalletRequest
	if id == "" || c.ShouldBindJSON(&request) != nil {
		utils.Error(c, http.StatusBadRequest, "Request wallet tidak valid")
		return
	}
	request.JenisWallet = strings.ToUpper(strings.TrimSpace(request.JenisWallet))
	if !oneOf(request.JenisWallet, "REGULAR", "ADMIN") {
		utils.Error(c, http.StatusBadRequest, "Jenis wallet harus REGULAR atau ADMIN")
		return
	}
	item, err := h.adminRepo.UpdateWalletType(c.Request.Context(), id, request.JenisWallet)
	if err != nil {
		adminError(c, err, "Gagal memperbarui wallet")
		return
	}
	utils.Success(c, http.StatusOK, "Wallet berhasil diperbarui", item)
}

func (h *AdminHandler) Banks(c *gin.Context) {
	result, err := h.adminRepo.ListBanks(c.Request.Context(), listParams(c))
	if err != nil {
		adminError(c, err, "Gagal mengambil data bank")
		return
	}
	utils.Success(c, http.StatusOK, "Data bank berhasil diambil", result)
}

func (h *AdminHandler) BankDetail(c *gin.Context) {
	id, ok := numericID(c, "id")
	if !ok {
		return
	}
	item, err := h.adminRepo.GetBank(c.Request.Context(), id)
	if err != nil {
		adminError(c, err, "Gagal mengambil bank")
		return
	}
	utils.Success(c, http.StatusOK, "Bank berhasil diambil", item)
}

func (h *AdminHandler) CreateBank(c *gin.Context) {
	request, ok := bindBank(c)
	if !ok {
		return
	}
	item, err := h.adminRepo.CreateBank(c.Request.Context(), bankInput(request))
	if err != nil {
		adminError(c, err, "Gagal membuat bank")
		return
	}
	utils.Success(c, http.StatusCreated, "Bank berhasil dibuat", item)
}

func (h *AdminHandler) UpdateBank(c *gin.Context) {
	id, ok := numericID(c, "id")
	if !ok {
		return
	}
	request, ok := bindBank(c)
	if !ok {
		return
	}
	item, err := h.adminRepo.UpdateBank(c.Request.Context(), id, bankInput(request))
	if err != nil {
		adminError(c, err, "Gagal memperbarui bank")
		return
	}
	utils.Success(c, http.StatusOK, "Bank berhasil diperbarui", item)
}

func (h *AdminHandler) DeleteBank(c *gin.Context) {
	id, ok := numericID(c, "id")
	if !ok {
		return
	}
	if err := h.adminRepo.DeleteBank(c.Request.Context(), id); err != nil {
		adminError(c, err, "Gagal menghapus bank")
		return
	}
	utils.Success(c, http.StatusOK, "Bank berhasil dihapus", nil)
}

func (h *AdminHandler) Rekening(c *gin.Context) {
	result, err := h.adminRepo.ListRekening(c.Request.Context(), listParams(c))
	if err != nil {
		adminError(c, err, "Gagal mengambil rekening mahasiswa")
		return
	}
	utils.Success(c, http.StatusOK, "Rekening mahasiswa berhasil diambil", result)
}

func (h *AdminHandler) RekeningDetail(c *gin.Context) {
	id, ok := numericID(c, "id")
	if !ok {
		return
	}
	item, err := h.adminRepo.GetRekening(c.Request.Context(), id)
	if err != nil {
		adminError(c, err, "Gagal mengambil rekening mahasiswa")
		return
	}
	utils.Success(c, http.StatusOK, "Rekening mahasiswa berhasil diambil", item)
}

func (h *AdminHandler) CreateRekening(c *gin.Context) {
	request, ok := bindRekening(c)
	if !ok {
		return
	}
	item, err := h.adminRepo.CreateRekening(c.Request.Context(), rekeningInput(request))
	if err != nil {
		adminError(c, err, "Gagal membuat rekening mahasiswa")
		return
	}
	utils.Success(c, http.StatusCreated, "Rekening mahasiswa berhasil dibuat", item)
}

func (h *AdminHandler) UpdateRekening(c *gin.Context) {
	id, ok := numericID(c, "id")
	if !ok {
		return
	}
	request, ok := bindRekening(c)
	if !ok {
		return
	}
	item, err := h.adminRepo.UpdateRekening(c.Request.Context(), id, rekeningInput(request))
	if err != nil {
		adminError(c, err, "Gagal memperbarui rekening mahasiswa")
		return
	}
	utils.Success(c, http.StatusOK, "Rekening mahasiswa berhasil diperbarui", item)
}

func (h *AdminHandler) DeleteRekening(c *gin.Context) {
	id, ok := numericID(c, "id")
	if !ok {
		return
	}
	if err := h.adminRepo.DeleteRekening(c.Request.Context(), id); err != nil {
		adminError(c, err, "Gagal menghapus rekening mahasiswa")
		return
	}
	utils.Success(c, http.StatusOK, "Rekening mahasiswa berhasil dihapus", nil)
}

func (h *AdminHandler) Merchants(c *gin.Context) {
	result, err := h.adminRepo.ListMerchants(c.Request.Context(), listParams(c))
	if err != nil {
		adminError(c, err, "Gagal mengambil data merchant")
		return
	}
	utils.Success(c, http.StatusOK, "Data merchant berhasil diambil", result)
}

func (h *AdminHandler) MerchantDetail(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		utils.Error(c, http.StatusBadRequest, "ID merchant tidak valid")
		return
	}
	item, err := h.adminRepo.GetMerchant(c.Request.Context(), id)
	if err != nil {
		adminError(c, err, "Gagal mengambil merchant")
		return
	}
	utils.Success(c, http.StatusOK, "Merchant berhasil diambil", item)
}

func (h *AdminHandler) CreateMerchant(c *gin.Context) {
	request, ok := bindMerchant(c, true)
	if !ok {
		return
	}
	item, err := h.adminRepo.CreateMerchant(c.Request.Context(), merchantInput(request))
	if err != nil {
		adminError(c, err, "Gagal membuat merchant")
		return
	}
	utils.Success(c, http.StatusCreated, "Merchant berhasil dibuat", item)
}

func (h *AdminHandler) UpdateMerchant(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		utils.Error(c, http.StatusBadRequest, "ID merchant tidak valid")
		return
	}
	request, ok := bindMerchant(c, false)
	if !ok {
		return
	}
	item, err := h.adminRepo.UpdateMerchant(c.Request.Context(), id, merchantInput(request))
	if err != nil {
		adminError(c, err, "Gagal memperbarui merchant")
		return
	}
	utils.Success(c, http.StatusOK, "Merchant berhasil diperbarui", item)
}

func (h *AdminHandler) DeleteMerchant(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		utils.Error(c, http.StatusBadRequest, "ID merchant tidak valid")
		return
	}
	if err := h.adminRepo.DeleteMerchant(c.Request.Context(), id); err != nil {
		adminError(c, err, "Gagal menghapus merchant")
		return
	}
	utils.Success(c, http.StatusOK, "Merchant berhasil dihapus", nil)
}

func (h *AdminHandler) Transactions(c *gin.Context) {
	result, err := h.adminRepo.ListTransactionsPaged(c.Request.Context(), listParams(c))
	if err != nil {
		adminError(c, err, "Gagal mengambil transaksi admin")
		return
	}
	utils.Success(c, http.StatusOK, "Data transaksi berhasil diambil", result)
}

func (h *AdminHandler) TransactionDetail(c *gin.Context) {
	id, ok := numericID(c, "id")
	if !ok {
		return
	}
	item, err := h.adminRepo.GetTransaction(c.Request.Context(), id)
	if err != nil {
		adminError(c, err, "Gagal mengambil transaksi")
		return
	}
	utils.Success(c, http.StatusOK, "Transaksi berhasil diambil", item)
}

func (h *AdminHandler) AuditLogs(c *gin.Context) {
	result, err := h.adminRepo.ListAuditLogsPaged(c.Request.Context(), listParams(c))
	if err != nil {
		adminError(c, err, "Gagal mengambil audit logs")
		return
	}
	utils.Success(c, http.StatusOK, "Audit logs berhasil diambil", result)
}

func (h *AdminHandler) DailyReports(c *gin.Context) {
	result, err := h.adminRepo.ListDailyReportsPaged(c.Request.Context(), listParams(c))
	if err != nil {
		adminError(c, err, "Gagal mengambil laporan harian")
		return
	}
	utils.Success(c, http.StatusOK, "Laporan harian berhasil diambil", result)
}

func listParams(c *gin.Context) repositories.AdminListParams {
	page := positiveInt(c.Query("page"), 1)
	perPage := positiveInt(c.Query("per_page"), 20)
	if perPage > 100 {
		perPage = 100
	}
	filters := map[string]string{}
	for _, key := range []string{"role", "status", "is_active", "bank_id", "jenis_wallet", "type", "action"} {
		filters[key] = strings.TrimSpace(c.Query(key))
	}
	return repositories.AdminListParams{
		Page: page, PerPage: perPage, Search: strings.TrimSpace(c.Query("search")),
		Sort: strings.TrimSpace(c.Query("sort")), Order: strings.TrimSpace(c.Query("order")), Filters: filters,
	}
}

func positiveInt(raw string, fallback int) int {
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return fallback
	}
	return value
}

func numericID(c *gin.Context, key string) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param(key), 10, 64)
	if err != nil || id == 0 {
		utils.Error(c, http.StatusBadRequest, "ID tidak valid")
		return 0, false
	}
	return id, true
}

func bindMahasiswa(c *gin.Context, create bool) (adminMahasiswaRequest, bool) {
	var request adminMahasiswaRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "Request body tidak valid")
		return request, false
	}
	request.NRP = strings.TrimSpace(request.NRP)
	request.NamaMahasiswa = strings.TrimSpace(request.NamaMahasiswa)
	request.Email = strings.ToLower(strings.TrimSpace(request.Email))
	request.Role = strings.ToLower(strings.TrimSpace(request.Role))
	request.Status = strings.ToUpper(strings.TrimSpace(request.Status))
	if request.NRP == "" || len(request.NRP) > 32 || request.NamaMahasiswa == "" || len(request.NamaMahasiswa) > 120 {
		utils.Error(c, http.StatusBadRequest, "NRP dan nama mahasiswa wajib diisi sesuai batas schema")
		return request, false
	}
	if len(request.Email) > 120 {
		utils.Error(c, http.StatusBadRequest, "Email melebihi 120 karakter")
		return request, false
	}
	if _, err := mail.ParseAddress(request.Email); err != nil || !strings.Contains(request.Email, "@") {
		utils.Error(c, http.StatusBadRequest, "Format email tidak valid")
		return request, false
	}
	if !oneOf(request.Role, "mahasiswa", "admin") || !oneOf(request.Status, "ACTIVE", "INACTIVE", "SUSPENDED") {
		utils.Error(c, http.StatusBadRequest, "Role atau status mahasiswa tidak valid")
		return request, false
	}
	if (create && len(request.Password) < 6) || (!create && request.Password != "" && len(request.Password) < 6) {
		utils.Error(c, http.StatusBadRequest, "Password minimal 6 karakter")
		return request, false
	}
	return request, true
}

func mahasiswaInput(request adminMahasiswaRequest, passwordHash string) repositories.AdminMahasiswaInput {
	return repositories.AdminMahasiswaInput{NRP: request.NRP, NamaMahasiswa: request.NamaMahasiswa,
		Email: request.Email, Role: request.Role, Status: request.Status, PasswordHash: passwordHash}
}

func bindBank(c *gin.Context) (adminBankRequest, bool) {
	var request adminBankRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "Request body tidak valid")
		return request, false
	}
	request.NamaBank = strings.TrimSpace(request.NamaBank)
	request.KodeBank = strings.ToUpper(strings.TrimSpace(request.KodeBank))
	request.BiayaAdmin = strings.TrimSpace(request.BiayaAdmin)
	fee, err := strconv.ParseFloat(request.BiayaAdmin, 64)
	if request.NamaBank == "" || len(request.NamaBank) > 80 || request.KodeBank == "" || len(request.KodeBank) > 20 || err != nil || fee < 0 {
		utils.Error(c, http.StatusBadRequest, "Nama, kode, dan biaya admin bank tidak valid")
		return request, false
	}
	return request, true
}

func bankInput(request adminBankRequest) repositories.AdminBankInput {
	return repositories.AdminBankInput{NamaBank: request.NamaBank, KodeBank: request.KodeBank,
		BiayaAdmin: request.BiayaAdmin, IsActive: request.IsActive}
}

func bindRekening(c *gin.Context) (adminRekeningRequest, bool) {
	var request adminRekeningRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "Request body tidak valid")
		return request, false
	}
	request.NoRekening = strings.TrimSpace(request.NoRekening)
	request.NamaPemilik = strings.TrimSpace(request.NamaPemilik)
	if request.MahasiswaID == 0 || request.BankID == 0 || request.NoRekening == "" || len(request.NoRekening) > 40 || request.NamaPemilik == "" || len(request.NamaPemilik) > 120 {
		utils.Error(c, http.StatusBadRequest, "Mahasiswa, bank, nomor rekening, dan nama pemilik wajib valid")
		return request, false
	}
	return request, true
}

func rekeningInput(request adminRekeningRequest) repositories.AdminRekeningInput {
	return repositories.AdminRekeningInput{MahasiswaID: request.MahasiswaID, BankID: request.BankID,
		NoRekening: request.NoRekening, NamaPemilik: request.NamaPemilik, IsActive: request.IsActive}
}

func bindMerchant(c *gin.Context, create bool) (adminMerchantRequest, bool) {
	var request adminMerchantRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, http.StatusBadRequest, "Request body tidak valid")
		return request, false
	}
	request.IDMerchant = strings.ToUpper(strings.TrimSpace(request.IDMerchant))
	request.NamaMerchant = strings.TrimSpace(request.NamaMerchant)
	request.Kategori = strings.TrimSpace(request.Kategori)
	request.Status = strings.ToUpper(strings.TrimSpace(request.Status))
	if create && (request.IDMerchant == "" || len(request.IDMerchant) > 20) {
		utils.Error(c, http.StatusBadRequest, "ID merchant wajib diisi dan maksimal 20 karakter")
		return request, false
	}
	if request.NamaMerchant == "" || len(request.NamaMerchant) > 120 || request.Kategori == "" || len(request.Kategori) > 80 || !oneOf(request.Status, "ACTIVE", "INACTIVE") {
		utils.Error(c, http.StatusBadRequest, "Nama, kategori, atau status merchant tidak valid")
		return request, false
	}
	return request, true
}

func merchantInput(request adminMerchantRequest) repositories.AdminMerchantInput {
	return repositories.AdminMerchantInput{IDMerchant: request.IDMerchant, NamaMerchant: request.NamaMerchant,
		Kategori: request.Kategori, Status: request.Status}
}

func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

func adminError(c *gin.Context, err error, fallback string) {
	if errors.Is(err, sql.ErrNoRows) {
		utils.Error(c, http.StatusNotFound, "Data tidak ditemukan")
		return
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "duplicate") {
		utils.Error(c, http.StatusConflict, "Data dengan identitas unik tersebut sudah digunakan")
		return
	}
	if strings.Contains(message, "foreign key") || strings.Contains(message, "constraint") || strings.Contains(message, "cannot delete") {
		utils.Error(c, http.StatusConflict, "Data tidak dapat diubah karena masih berelasi dengan data lain")
		return
	}
	utils.Error(c, http.StatusInternalServerError, fallback)
}
