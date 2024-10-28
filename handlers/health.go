package handlers

import (
	"audio-server/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetHealth() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP", "version": utils.AppVersion})
	}
}
