package metadata

import (
	"audio-server/utils"
	"go.uber.org/zap"
	"time"
)

type EventHandlerFunc func(msg Event)

type Metadata struct {
	SessionId         string
	StartTime         time.Time
	StopTime          *time.Time
	Status            StatusType
	CallerPhoneNumber *string
	Language          *string
}

type MetadataManager struct {
	metadataMap map[string]*Metadata
	eventChan   chan Event
	handlers    map[EventType]EventHandlerFunc
}

func NewMetadataManager(eventChan chan Event) *MetadataManager {

	metadataManager := MetadataManager{
		metadataMap: make(map[string]*Metadata),
		eventChan:   eventChan,
		handlers:    make(map[EventType]EventHandlerFunc),
	}
	metadataManager.handlers[UPDATE] = metadataManager.handleUpdate
	metadataManager.handlers[CREATE] = metadataManager.handleCreate

	go metadataManager.run()
	return &metadataManager
}

func (m *MetadataManager) run() {
	utils.Logger.Info("Start MetadataManager")
	for event := range m.eventChan {
		m.handleEvent(event)
	}
}

func (m *MetadataManager) handleEvent(event Event) {
	utils.Logger.Info("Reçu", zap.Any("Event", event))
	handlerFunc, exists := m.handlers[event.Type]
	if !exists {
		utils.Logger.Warn("Type de event non géré", zap.Any("event", event))
		return
	}

	handlerFunc(event)
	return
}

func (m *MetadataManager) handleCreate(event Event) {
	params, ok := event.Parameters.(CreateParameters)
	if !ok {
		// Gérer l'erreur de type
		return
	}
	metadata := &Metadata{
		SessionId:         event.SessionId,
		StartTime:         params.StartTime,
		Status:            CREATED,
		CallerPhoneNumber: params.CallerPhoneNumber,
	}
	m.metadataMap[event.SessionId] = metadata
}

func (m *MetadataManager) handleUpdate(event Event) {
	params, ok := event.Parameters.(UpdateParameters)
	if !ok {
		// Gérer l'erreur de type
		return
	}
	metadata, exists := m.metadataMap[event.SessionId]
	metadata.Status = params.Status
	if !exists {
		// Gérer le cas où la session n'existe pas
		return
	}
	if params.StopTime != nil {
		metadata.StopTime = params.StopTime
	}
	if params.CallerPhoneNumber != nil {
		metadata.CallerPhoneNumber = params.CallerPhoneNumber
	}
	if params.Language != nil {
		metadata.Language = params.Language
	}
}
