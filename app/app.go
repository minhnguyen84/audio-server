package app

import (
	"audio-server/audiohook_genesys"
	"audio-server/audiolab"
	"audio-server/audiolab/metadata"
	"audio-server/handlers"
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

	audioStorageChan := make(chan audiolab.Message, 10)
	audioDispatcher := audiolab.NewAudioDispatcher(audioStorageChan)

	metadataChan := make(chan metadata.Event, 10)
	metadatManage := metadata.NewMetadataManager(metadataChan)

	audioHandler := audiohook_genesys.NewAudioHandler(audioDispatcher)
	audiohookHandler := audiohook_genesys.NewWebSocketHandler(audioHandler, metadataChan)

	inference := handlers.NewInferenceHandler(metadatManage)

	uploader, err := utils.NewS3Uploader(*utils.GetConfig())
	if err != nil {
		utils.Logger.Error("Erreur lors de la création du uploader", zap.Error(err))
		panicWhenSetup("S3Uploader")
	}
	if _, err := outbound.NewFileStorage(audioStorageChan, uploader, *utils.GetConfig()); err != nil {
		utils.Logger.Error("error create NewFileStorage : ",
			zap.Any("con", utils.GetConfig()),
			zap.Error(err))
		panicWhenSetup("FileStorage")
	}
	r := router.InitializeRouter(audiohookHandler, inference)
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

func panicWhenSetup(module string) {
	panic(fmt.Sprintf("Could not setup %s", module))
}
