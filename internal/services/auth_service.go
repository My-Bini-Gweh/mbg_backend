package services

import (
	"context"
	"database/sql"
	"strings"

	"mbg-backend/internal/models"
	"mbg-backend/internal/repositories"
	"mbg-backend/internal/utils"
)

type AuthService struct {
	authRepo  *repositories.AuthRepository
	jwtSecret string
}

type RegisterInput struct {
	NRP           string
	NamaMahasiswa string
	Email         string
	Password      string
}

type LoginResult struct {
	Token     string           `json:"token"`
	Mahasiswa models.Mahasiswa `json:"mahasiswa"`
}

func NewAuthService(authRepo *repositories.AuthRepository, jwtSecret string) *AuthService {
	return &AuthService{authRepo: authRepo, jwtSecret: jwtSecret}
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*LoginResult, error) {
	passwordHash, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	mahasiswa, err := s.authRepo.CreateMahasiswaWithAuth(ctx, repositories.RegisterMahasiswaInput{
		NRP:           input.NRP,
		NamaMahasiswa: input.NamaMahasiswa,
		Email:         input.Email,
		PasswordHash:  passwordHash,
		Role:          "mahasiswa",
	})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return nil, ErrDuplicateAccount
		}
		return nil, err
	}

	token, err := utils.GenerateToken(s.jwtSecret, mahasiswa.ID, mahasiswa.Role)
	if err != nil {
		return nil, err
	}

	return &LoginResult{Token: token, Mahasiswa: *mahasiswa}, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	record, err := s.authRepo.FindAuthByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !utils.CheckPassword(password, record.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	if err := s.authRepo.UpdateLastLogin(ctx, record.Mahasiswa.ID); err != nil {
		return nil, err
	}

	token, err := utils.GenerateToken(s.jwtSecret, record.Mahasiswa.ID, record.Mahasiswa.Role)
	if err != nil {
		return nil, err
	}

	return &LoginResult{Token: token, Mahasiswa: record.Mahasiswa}, nil
}
