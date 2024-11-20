package handlers

import (
	"audio-server/audiolab/metadata"
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"time"
)

type Inference struct {
	SessionId         string  `json:"session_id"`
	CallerPhoneNumber string  `json:"caller_phone_number"`
	Status            string  `json:"call_status"`
	StartTime         string  `json:"start_time"`
	EndTime           *string `json:"end_time,omitempty"` // Peut être null si pas terminé
	InfarctusProba    int     `json:"infarctus_proba"`
	AVCProba          int     `json:"avc_proba"`
}

type InferenceHandler struct {
	metadataManager *metadata.MetadataManager
}

func NewInferenceHandler(metadataManager *metadata.MetadataManager) *InferenceHandler {
	return &InferenceHandler{
		metadataManager: metadataManager,
	}
}

func (i *InferenceHandler) HandlerInference(c *gin.Context) {
	c.JSON(http.StatusOK, transformMetadataToInference(i.metadataManager.GetAllMetadata()))
}

func transformMetadataToInference(metadataMap map[string]*metadata.Metadata) []Inference {
	var inferences []Inference

	for _, metadataList := range metadataMap {
		// Générer les probabilités
		infarctusProba := getRandomProba()
		avcProba := getRandomProba()

		// Transformer les métadonnées en inference
		inference := Inference{
			SessionId:         metadataList.SessionId,
			CallerPhoneNumber: safeString(metadataList.CallerPhoneNumber, "Unknown"),
			Status:            string(metadataList.Status),
			StartTime:         metadataList.StartTime.Format(time.RFC3339),
			InfarctusProba:    infarctusProba,
			AVCProba:          avcProba,
		}

		// Ajouter StopTime si disponible
		if metadataList.StopTime != nil {
			endTime := metadataList.StopTime.Format(time.RFC3339)
			inference.EndTime = &endTime
		}

		// Ajouter à la liste des inferences
		inferences = append(inferences, inference)
	}

	return inferences
}

func getRandomProba() int {
	weights := []int{0, 3, 4, 0, 3, 4, 0, 3, 4, 1, 2, 5, 6, 7, 8, 9, 10}
	return weights[rand.Intn(len(weights))]
}

func safeString(ptr *string, defaultValue string) string {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}
