package app

import (
	"audio-server/audio"
	"audio-server/audio/metadata"
	"audio-server/audiohook_genesys"
	"audio-server/outbound"
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

	audioStorageChan := make(chan audio.Message, 10)
	audioDispatcher := audio.NewAudioDispatcher(audioStorageChan)

	audioHandler := audio.NewAudioHandler(audioDispatcher)

	metadataChan := make(chan metadata.Event, 10)
	audiohookHandler := audiohook_genesys.NewWebSocketHandler(audioHandler, metadataChan)
	metadata.NewMetadataManager(metadataChan)

	uploader, err := utils.NewS3Uploader(*utils.GetConfig())
	if err != nil {
		utils.Logger.Error("Erreur lors de la création du uploader", zap.Error(err))
	}
	if _, err := outbound.NewFileStorage(audioStorageChan, uploader); err != nil {
		utils.Logger.Error("error create NewFileStorage : ", zap.Error(err))
	}
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
