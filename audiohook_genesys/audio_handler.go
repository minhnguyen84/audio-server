package audiohook_genesys

import (
	"audio-server/audiolab"
	"audio-server/utils"
	"encoding/binary"
	"github.com/go-audio/audio"
)

type AudioHandler struct {
	Dispatcher *audiolab.AudioDispatcher
}

func NewAudioHandler(dispatcher *audiolab.AudioDispatcher) *AudioHandler {
	utils.Logger.Info("Init AudioHandler")
	return &AudioHandler{
		Dispatcher: dispatcher,
	}
}

func (ah *AudioHandler) HandleAudioData(sessionId string, binaryData []byte) {
	// Préparer les échantillons PCM pour l'écriture
	numSamples := len(binaryData) / 2 // 2 bytes par échantillon 16 bits

	bufferInternal := newIntBuffer(numSamples)
	bufferExternal := newIntBuffer(numSamples)

	for i := 1; i < numSamples; i++ {
		bufferInternal.Data[i] = int(int16(binary.LittleEndian.Uint16(binaryData[i*2-1:])))
		bufferExternal.Data[i] = int(int16(binary.LittleEndian.Uint16(binaryData[i*2:])))
	}

	ah.Dispatcher.Dispatch(audiolab.Message{
		SessionId: sessionId,
		IsClosed:  false,
		Data: audiolab.Data{
			Internal: bufferInternal,
			External: bufferExternal,
		},
	})

}

func (ah *AudioHandler) Close(sessionId string) {
	ah.Dispatcher.Dispatch(audiolab.Message{
		SessionId: sessionId,
		IsClosed:  true,
	})
	utils.Logger.Debug("Fichier WAV enregistré avec succès.")
}

func newIntBuffer(numSamples int) *audio.IntBuffer {
	return &audio.IntBuffer{
		Format: &audio.Format{
			SampleRate:  8000,
			NumChannels: 1,
		},
		Data:           make([]int, numSamples),
		SourceBitDepth: 16,
	}
}
