package audiohook_genesys

import (
	"audio-server/audio"
	"audio-server/audio/metadata"
	"audio-server/utils"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"log"
	"net/http"
	"time"
)

// WebSocketHandler gère les connexions WebSocket
type WebSocketHandler struct {
	upgrader     websocket.Upgrader
	metadataChan chan metadata.Event
}

type WebSocketSession struct {
	audioHandler   *audio.Handler
	messageHandler *MessageHandler
}

// NewWebSocketHandler crée une nouvelle instance de WebSocketHandler
func NewWebSocketHandler(metadataChan chan metadata.Event) *WebSocketHandler {
	return &WebSocketHandler{
		upgrader: websocket.Upgrader{
			//  ReadBufferSize:  4096, default value - à ré-évaluer si besoin
			// WriteBufferSize: 4096, default value - à ré-évaluer si besoin
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		metadataChan: metadataChan,
	}
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		utils.Logger.Error("Erreur lors de la mise à niveau de la connexion: ", zap.Error(err))
		return
	}
	defer conn.Close()

	sessionId := c.GetHeader("audiohook-session-id")
	if sessionId == "" {
		utils.Logger.Error("Pas de 'audiohook-session-id' dans header")
		//TODO il faut gérer comme demander audiohook
		return
	}
	utils.Logger.Info("Nouvelle connexion WebSocket établie", zap.String("sessionId", sessionId))

	closeChan := make(chan struct{})
	wsSession := h.newSession(sessionId, closeChan)
	h.createMetadata(sessionId)

	// Goroutine pour gérer la fermeture de la connexion et les logs récapitulatifs
	go func() {
		<-closeChan
		log.Println("Fermeture de la connexion WebSocket")
		// fermer audioHandler = close file and buffer
		wsSession.audioHandler.Close()
		// Fermer la connexion après un court délai pour s'assurer que le message 'closed' est envoyé
		time.Sleep(400 * time.Millisecond)
		conn.Close()
	}()

	// Démarrer l'écoute des messages
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			utils.Logger.Error("Erreur lors de la lecture du message: ", zap.Error(err))
			break
		}

		switch messageType {
		case websocket.BinaryMessage:
			// Gérer les données audio binaires
			utils.Logger.Info("BinaryMessage")
			wsSession.audioHandler.HandleAudioData(message)
		default:
			// Gérer les messages textuels (contrôle) - y compris les messages "keep-alive"
			var msg MessageReceived
			if err := json.Unmarshal(message, &msg); err != nil {
				utils.Logger.Error("Erreur de décodage du message JSON: ", zap.Error(err))
				continue
			}
			messageSent, err := wsSession.messageHandler.handleMessage(msg)

			if err != nil {
				//TODO gérer les errors
			}
			msgBytes, err := json.Marshal(messageSent)
			if err != nil {
				utils.Logger.Error("Erreur lors de la sérialisation du message",
					zap.Any("message", messageSent),
					zap.Error(err))
				continue
			}

			err = conn.WriteMessage(websocket.TextMessage, msgBytes)
			if err != nil {
				utils.Logger.Error("Erreur lors de l'envoi du message", zap.Error(err))
				continue
			}
		}
	}

	utils.Logger.Info("Connexion WebSocket fermée.")
}

func (h *WebSocketHandler) createMetadata(sessionId string) {
	metadataEvent := metadata.Event{
		Type:      metadata.CREATE,
		SessionId: sessionId,
		Parameters: metadata.CreateParameters{
			StartTime:         time.Now(),
			CallerPhoneNumber: nil,
		},
	}
	h.metadataChan <- metadataEvent
}

func (h *WebSocketHandler) closeMetadata(sessionId string) {
	metadataEvent := metadata.Event{
		Type:      metadata.UPDATE,
		SessionId: sessionId,
		Parameters: metadata.UpdateParameters{
			Status: metadata.FINISHED,
		},
	}
	h.metadataChan <- metadataEvent
}

func (h *WebSocketHandler) newSession(sessionId string, closeChan chan struct{}) *WebSocketSession {
	audioHandler, err := audio.NewAudioHandler(sessionId)
	if err != nil {
		//TODO : il faut gérer
	}
	return &WebSocketSession{
		audioHandler:   audioHandler,
		messageHandler: NewMessageHandler(sessionId, closeChan, h.metadataChan),
	}
}
