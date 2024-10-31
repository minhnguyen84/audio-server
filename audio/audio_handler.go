package audio

import (
	"audio-server/utils"
	"encoding/binary"
	"github.com/go-audio/audio"
)

type Handler struct {
	Dispatcher *AudioDispatcher
}

func NewAudioHandler(dispatcher *AudioDispatcher) *Handler {
	return &Handler{
		Dispatcher: dispatcher,
	}
}

func (ah *Handler) HandleAudioData(sessionId string, binaryData []byte) {
	// Préparer les échantillons PCM pour l'écriture
	numSamples := len(binaryData) / 2 // 2 bytes par échantillon 16 bits

	bufferInternal := newIntBuffer(numSamples)
	bufferExternal := newIntBuffer(numSamples)

	for i := 1; i < numSamples; i++ {
		bufferInternal.Data[i] = int(int16(binary.LittleEndian.Uint16(binaryData[i*2-1:])))
		bufferExternal.Data[i] = int(int16(binary.LittleEndian.Uint16(binaryData[i*2:])))
	}

	ah.Dispatcher.Dispatch(Message{
		SessionId: sessionId,
		IsClosed:  false,
		Data: Data{
			Internal: bufferInternal,
			External: bufferExternal,
		},
	})

}

func (ah *Handler) Close(sessionId string) {
	ah.Dispatcher.Dispatch(Message{
		SessionId: sessionId,
		IsClosed:  true,
	})
	utils.Logger.Info("Fichier WAV enregistré avec succès.")
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
