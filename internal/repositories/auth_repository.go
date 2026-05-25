package repositories

import (
	"context"
	"database/sql"

	"mbg-backend/internal/models"
)

type AuthRepository struct {
	db *sql.DB
}

type RegisterMahasiswaInput struct {
	NRP           string
	NamaMahasiswa string
	Email         string
	PasswordHash  string
	Role          string
}

type AuthRecord struct {
	Mahasiswa    models.Mahasiswa
	PasswordHash string
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) CreateMahasiswaWithAuth(ctx context.Context, input RegisterMahasiswaInput) (*models.Mahasiswa, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		INSERT INTO mahasiswa (nrp, nama_mahasiswa, email, role)
		VALUES (?, ?, ?, ?)
	`, input.NRP, input.NamaMahasiswa, input.Email, input.Role)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO mahasiswa_auth (mahasiswa_id, password_hash)
		VALUES (?, ?)
	`, id, input.PasswordHash)
	if err != nil {
		return nil, err
	}

	var mahasiswa models.Mahasiswa
	err = tx.QueryRowContext(ctx, `
		SELECT id_mahasiswa, nrp, nama_mahasiswa, email, role, status,
			DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s')
		FROM mahasiswa
		WHERE id_mahasiswa = ?
	`, id).Scan(
		&mahasiswa.ID,
		&mahasiswa.NRP,
		&mahasiswa.NamaMahasiswa,
		&mahasiswa.Email,
		&mahasiswa.Role,
		&mahasiswa.Status,
		&mahasiswa.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &mahasiswa, nil
}

func (r *AuthRepository) FindAuthByEmail(ctx context.Context, email string) (*AuthRecord, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	var record AuthRecord
	err := r.db.QueryRowContext(ctx, `
		SELECT m.id_mahasiswa, m.nrp, m.nama_mahasiswa, m.email, m.role, m.status,
			DATE_FORMAT(m.created_at, '%Y-%m-%d %H:%i:%s'),
			a.password_hash
		FROM mahasiswa m
		INNER JOIN mahasiswa_auth a ON a.mahasiswa_id = m.id_mahasiswa
		WHERE m.email = ?
	`, email).Scan(
		&record.Mahasiswa.ID,
		&record.Mahasiswa.NRP,
		&record.Mahasiswa.NamaMahasiswa,
		&record.Mahasiswa.Email,
		&record.Mahasiswa.Role,
		&record.Mahasiswa.Status,
		&record.Mahasiswa.CreatedAt,
		&record.PasswordHash,
	)
	if err != nil {
		return nil, err
	}

	return &record, nil
}

func (r *AuthRepository) UpdateLastLogin(ctx context.Context, mahasiswaID uint64) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	_, err := r.db.ExecContext(ctx, `
		UPDATE mahasiswa_auth
		SET last_login_at = CURRENT_TIMESTAMP
		WHERE mahasiswa_id = ?
	`, mahasiswaID)
	return err
}
