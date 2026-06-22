package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"mbg-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

const (
	ContextMahasiswaID = "id_mahasiswa"
	ContextRole        = "role"
)

func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			utils.Error(c, http.StatusUnauthorized, "Token tidak ditemukan")
			c.Abort()
			return
		}

		tokenString, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || strings.TrimSpace(tokenString) == "" {
			utils.Error(c, http.StatusUnauthorized, "Format token tidak valid")
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(secret, strings.TrimSpace(tokenString))
		if err != nil {
			utils.Error(c, http.StatusUnauthorized, "Token tidak valid atau sudah kadaluarsa")
			c.Abort()
			return
		}

		c.Set(ContextMahasiswaID, claims.MahasiswaID)
		c.Set(ContextRole, claims.Role)
		c.Next()
	}
}

func Role(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		role, _ := c.Get(ContextRole)
		roleString, _ := role.(string)
		if _, ok := allowed[roleString]; !ok {
			utils.Error(c, http.StatusForbidden, "Akses tidak diizinkan")
			c.Abort()
			return
		}

		c.Next()
	}
}

func ActiveAccount(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		mahasiswaID, _, ok := CurrentUser(c)
		if !ok {
			utils.Error(c, http.StatusUnauthorized, "Token tidak valid")
			c.Abort()
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()
		var active bool
		err := db.QueryRowContext(ctx, `
			SELECT status = 'ACTIVE' FROM mahasiswa WHERE id_mahasiswa = ?
		`, mahasiswaID).Scan(&active)
		if err != nil {
			if err == sql.ErrNoRows {
				utils.Error(c, http.StatusUnauthorized, "Akun tidak ditemukan")
			} else {
				utils.Error(c, http.StatusInternalServerError, "Gagal memverifikasi status akun")
			}
			c.Abort()
			return
		}
		if !active {
			utils.Error(c, http.StatusForbidden, "Akun tidak aktif")
			c.Abort()
			return
		}
		c.Next()
	}
}

func CurrentUser(c *gin.Context) (uint64, string, bool) {
	idValue, hasID := c.Get(ContextMahasiswaID)
	roleValue, hasRole := c.Get(ContextRole)
	if !hasID || !hasRole {
		return 0, "", false
	}

	id, ok := idValue.(uint64)
	if !ok {
		return 0, "", false
	}

	role, ok := roleValue.(string)
	if !ok {
		return 0, "", false
	}

	return id, role, true
}
