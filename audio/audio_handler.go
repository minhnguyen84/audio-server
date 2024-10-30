package audio

import (
	"audio-server/utils"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"time"
)

type Handler struct {
	encoderInternal *wav.Encoder
	encoderExternal *wav.Encoder
	uploader        *utils.S3Uploader
	fileInternal    *os.File
	fileExternal    *os.File
	tempFilePath    string
	sessionID       string
}

func NewAudioHandler(sessionId string) (*Handler, error) {
	// Générer un chemin temporaire unique pour chaque session
	fileInternal, err := os.Create(fileName("internal", sessionId))
	if err != nil {
		return nil, fmt.Errorf("Erreur lors de la création du fichier temporaire", zap.Error(err))
	}

	// Initialisation de l'encodeur WAV avec les paramètres PCM = 1
	encoderInternal := wav.NewEncoder(fileInternal, 8000, 16, 1, 1)

	fileExternal, err := os.Create(fileName("external", sessionId))
	if err != nil {
		return nil, fmt.Errorf("Erreur lors de la création du fichier temporaire", zap.Error(err))
	}

	// Initialisation de l'encodeur WAV avec les paramètres PCM = 1
	encoderExternal := wav.NewEncoder(fileExternal, 8000, 16, 1, 1)

	uploader, err := utils.NewS3Uploader(*utils.GetConfig())
	if err != nil {
		return nil, fmt.Errorf("Erreur lors de la création du uploader", zap.Error(err))
	}

	return &Handler{
		encoderInternal: encoderInternal,
		encoderExternal: encoderExternal,
		fileInternal:    fileInternal,
		fileExternal:    fileExternal,
		sessionID:       sessionId,
		uploader:        uploader,
	}, nil
}

func (ah *Handler) HandleAudioData(binaryData []byte) {
	// Préparer les échantillons PCM pour l'écriture
	numSamples := len(binaryData) / 2 // 2 bytes par échantillon 16 bits

	bufferInternal := newIntBuffer(numSamples)
	bufferExternal := newIntBuffer(numSamples)

	for i := 1; i < numSamples; i++ {
		bufferInternal.Data[i] = int(int16(binary.LittleEndian.Uint16(binaryData[i*2-1:])))
		bufferExternal.Data[i] = int(int16(binary.LittleEndian.Uint16(binaryData[i*2:])))
	}

	// Écrire les échantillons PCM dans l'encodeur WAV
	if err := ah.encoderInternal.Write(bufferInternal); err != nil {
		utils.Logger.Error("Erreur lors de l'écriture des données PCM dans le WAV", zap.Error(err))
	}

	// Écrire les échantillons PCM dans l'encodeur WAV
	if err := ah.encoderExternal.Write(bufferExternal); err != nil {
		utils.Logger.Error("Erreur lors de l'écriture des données PCM dans le WAV", zap.Error(err))
	}

}

func (ah *Handler) Close() {
	// Fermer l'encodeur et le fichier local si ce n'est pas déjà fait
	if ah.encoderInternal != nil {
		if err := ah.encoderInternal.Close(); err != nil {
			utils.Logger.Error("Erreur lors de la fermeture de l'encodeur WAV", zap.Error(err))
		}
	}
	// Fermer l'encodeur et le fichier local si ce n'est pas déjà fait
	if ah.encoderExternal != nil {
		if err := ah.encoderExternal.Close(); err != nil {
			utils.Logger.Error("Erreur lors de la fermeture de l'encodeur WAV", zap.Error(err))
		}
	}

	// upload sur S3

	// delete local fil

	if ah.fileInternal != nil {
		ah.uploadFile(ah.fileInternal)

		if err := ah.fileInternal.Close(); err != nil {
			utils.Logger.Error("Erreur lors de la fermeture du fichier WAV", zap.Error(err))
		}

		deleteLocalFile(ah.fileInternal)

	}
	if ah.fileExternal != nil {
		ah.uploadFile(ah.fileExternal)
		if err := ah.fileExternal.Close(); err != nil {
			utils.Logger.Error("Erreur lors de la fermeture du fichier WAV", zap.Error(err))
		}
		deleteLocalFile(ah.fileExternal)
	}

	utils.Logger.Info("Fichier WAV enregistré avec succès.")
}

func (ah *Handler) uploadFile(file *os.File) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := ah.uploader.UploadFile(ctx, file, ah.fileExternal.Name())
	if err != nil {
		utils.Logger.Error("Erreur lors de la uploade file ",
			zap.String("fileName", ah.fileExternal.Name()),
			zap.Error(err))
		return
	}
}

func deleteLocalFile(file *os.File) {
	// Supprimer le fichier local
	err := os.Remove(file.Name())
	if err != nil {
		utils.Logger.Error("Erreur lors de la suppression du fichier local",
			zap.String("fileName", file.Name()),
			zap.Error(err))
		return
	}
}

func newIntBuffer(numSamples int) *audio.IntBuffer {
	return &audio.IntBuffer{
		Format: &audio.Format{
			SampleRate:  48000,
			NumChannels: 1,
		},
		Data:           make([]int, numSamples),
		SourceBitDepth: 16,
	}
}

func fileName(source, sessionId string) string {
	utils.Logger.Info("TempDir : " + os.TempDir())
	return filepath.Join(os.TempDir(), fmt.Sprintf("%s_audio_%s.wav", source, sessionId))
}
