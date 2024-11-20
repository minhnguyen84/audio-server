package router

import (
	"audio-server/audiohook_genesys"
	"audio-server/handlers"
	"audio-server/utils"
	"github.com/gin-gonic/gin"
	"time"
)

func InitializeRouter(audiohookHandler *audiohook_genesys.WebSocketHandler, inferenceHandler *handlers.InferenceHandler) *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Header("X-App-Version", utils.AppVersion)

		// Ajouter les en-têtes pour gérer CORS
		c.Header("Access-Control-Allow-Origin", "*") // Changez "*" par une origine spécifique si nécessaire
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")

		// Répondre directement aux requêtes OPTIONS
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

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
	r.GET("/inference", inferenceHandler.HandlerInference)

	r.Use(RecoveryWithZap(utils.Logger, true))
	r.NoRoute(handlers.NoRoute)

	return r
}
