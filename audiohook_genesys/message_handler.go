package audiohook_genesys

import (
	"audio-server/utils"
	"encoding/json"
	"go.uber.org/zap"
)

type MessageHandlerFunc func(msg MessageReceived) (*MessageSent, error)

type MessageHandler struct {
	seq       int
	clientseq int
	sessionId string
	handlers  map[string]MessageHandlerFunc
	closeChan chan struct{}
}

func NewMessageHandler(sessionId string, closeChan chan struct{}) *MessageHandler {
	messageHandler := &MessageHandler{
		seq:       0,
		clientseq: 0,
		sessionId: sessionId,
		handlers:  make(map[string]MessageHandlerFunc),
		closeChan: closeChan,
	}
	messageHandler.handlers[MessageTypeOpen] = messageHandler.handleOpen
	messageHandler.handlers[MessageTypePing] = messageHandler.handlePing
	messageHandler.handlers[MessageTypeClose] = messageHandler.handleClose
	return messageHandler
}

func (mh *MessageHandler) handleMessage(msg MessageReceived) (*MessageSent, error) {
	handlerFunc, exists := mh.handlers[msg.Type]
	if !exists {
		utils.Logger.Info("Type de message non géré", zap.Any("message", msg))
		return nil, nil
	}
	mh.updateClientSeq(msg)
	return handlerFunc(msg)
}

func (mh *MessageHandler) handleOpen(msg MessageReceived) (*MessageSent, error) {
	var params OpenParameters
	if err := json.Unmarshal(msg.Parameters, &params); err != nil {
		utils.Logger.Error("Erreur de décodage des paramètres 'open': %v", zap.Error(err))
		return nil, err
	}

	utils.Logger.Info("Reçu 'open'",
		zap.String("sessionID-message", msg.ID),
		zap.String("sessionID", mh.sessionId))

	//TODO : ajouter la gestion des metadata

	// Envoyer la réponse 'opened'
	openedMsg := MessageSent{
		Version:   "2",
		Type:      MessageTypeOpened,
		Seq:       mh.nextSequence(),
		ClientSeq: msg.Seq,
		ID:        mh.sessionId,
		Parameters: OpenedParameters{
			Media: []MediaParameter{
				{
					Type:     MediaTypeAudio,
					Format:   MediaFormatPCMU,
					Channels: []string{MediaChannelExternal, MediaChannelInternal},
					Rate:     MediaRate,
				},
			},
			StartPaused: false,
			DiscardTo:   nil,
		},
	}
	return &openedMsg, nil
}

func (mh *MessageHandler) handlePing(msg MessageReceived) (*MessageSent, error) {
	var params PingParameters
	if err := json.Unmarshal(msg.Parameters, &params); err != nil {
		utils.Logger.Error("Erreur de décodage des paramètres 'ping': %v", zap.Error(err))
		return nil, err
	}

	utils.Logger.Info("Reçu 'ping'",
		zap.String("sessionID-message", msg.ID),
		zap.String("sessionID", mh.sessionId))

	pongMsg := MessageSent{
		Version:    "2",
		Type:       MessageTypePong,
		Seq:        mh.nextSequence(),
		ClientSeq:  msg.Seq,
		ID:         mh.sessionId,
		Parameters: PongParameters{},
	}
	return &pongMsg, nil
}

func (mh *MessageHandler) handleClose(msg MessageReceived) (*MessageSent, error) {
	var params ClosedParameters
	if err := json.Unmarshal(msg.Parameters, &params); err != nil {
		utils.Logger.Error("Erreur de décodage des paramètres 'close': %v", zap.Error(err))
		return nil, err
	}

	utils.Logger.Info("Reçu 'close'",
		zap.String("sessionID-message", msg.ID),
		zap.String("sessionID", mh.sessionId))

	// Envoyer la réponse 'closed'
	closedMsg := MessageSent{
		Version:    "2",
		Type:       MessageTypeClosed,
		Seq:        mh.nextSequence(),
		ClientSeq:  msg.Seq,
		ID:         mh.sessionId,
		Parameters: PongParameters{},
	}

	// Signaler la fermeture de la connexion
	mh.closeChan <- struct{}{}

	return &closedMsg, nil
}

// TODO : implementer update, error message

func (mh *MessageHandler) updateClientSeq(msg MessageReceived) {
	mh.clientseq = mh.seq
}

func (mh *MessageHandler) nextSequence() int {
	mh.seq++
	return mh.seq
}
