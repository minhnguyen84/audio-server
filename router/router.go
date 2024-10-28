package router

import (
	"audio-server/audiohook_genesys"
	"audio-server/handlers"
	"audio-server/utils"
	"github.com/gin-gonic/gin"
	"time"
)

func InitializeRouter(audiohookHandler *audiohook_genesys.WebSocketHandler) *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Header("X-App-Version", utils.AppVersion)
		c.Next()
	})

	untracedGroup := r.Group("/")
	untracedGroup.Use(Ginzap(utils.Logger, time.RFC3339, true, false))
	{
		untracedGroup.GET("/", handlers.GetHealth())
		untracedGroup.GET("/health", handlers.GetHealth())
	}

	r.Use(Ginzap(utils.Logger, time.RFC3339, true, true))
	r.GET("/api/v1/audiohook/ws", audiohookHandler.HandleWebSocket)

	r.Use(RecoveryWithZap(utils.Logger, true))
	r.NoRoute(handlers.NoRoute)

	return r
}
