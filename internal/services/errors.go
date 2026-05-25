package services

import "errors"

var (
	ErrInvalidCredentials = errors.New("email atau password salah")
	ErrDuplicateAccount   = errors.New("email atau NRP sudah terdaftar")
	ErrWalletForbidden    = errors.New("wallet tidak sesuai dengan mahasiswa login")
)

type BusinessError struct {
	Message string
}

func (e BusinessError) Error() string {
	return e.Message
}
