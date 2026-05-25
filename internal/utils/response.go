package utils

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func Success(c *gin.Context, status int, message string, data any) {
	if data == nil {
		data = gin.H{}
	}

	c.JSON(status, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

func Error(c *gin.Context, status int, message string) {
	if message == "" {
		message = http.StatusText(status)
	}

	c.JSON(status, gin.H{
		"success": false,
		"message": message,
	})
}

func CleanDatabaseError(err error) string {
	if err == nil {
		return ""
	}

	message := err.Error()
	if idx := strings.LastIndex(message, ": "); idx >= 0 && idx+2 < len(message) {
		message = strings.TrimSpace(message[idx+2:])
	}

	if message == "" {
		return "Terjadi kesalahan database"
	}
	return message
}
