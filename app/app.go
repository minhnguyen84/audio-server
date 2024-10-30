package app

import (
	"audio-server/audio/metadata"
	"audio-server/audiohook_genesys"
	"audio-server/router"
	"audio-server/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type App struct {
	Config utils.AppConfig
	Router *gin.Engine
}

func New() *App {
	app := &App{}
	app.setup()
	return app
}

func (app *App) setup() {
	config := utils.GetConfig()
	metadataChan := make(chan metadata.Event)
	audiohookHandler := audiohook_genesys.NewWebSocketHandler(metadataChan)
	metadata.NewMetadataManager(metadataChan)
	r := router.InitializeRouter(audiohookHandler)
	app.Config = *config
	app.Router = r
}

func (app *App) Run() {
	// Serving application
	port := app.Config.Port
	utils.Logger.Info(fmt.Sprintf("RUN APP on PORT %d", port))
	if err := app.Router.Run(fmt.Sprintf(":%d", port)); err != nil {
		utils.Logger.Error("error on running : ", zap.Error(err))
	}
}
