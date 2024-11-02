package audiolab

import (
	"audio-server/utils"
	"github.com/go-audio/audio"
	"sync"
)

type Message struct {
	SessionId string
	IsClosed  bool
	Data      Data
}

type Data struct {
	Internal *audio.IntBuffer
	External *audio.IntBuffer
}

type AudioDispatcher struct {
	outputChannels []chan Message
	mu             sync.RWMutex
}

func NewAudioDispatcher(initialSubscribers ...chan Message) *AudioDispatcher {
	utils.Logger.Info("Init AudioDispatcher")
	dispatcher := &AudioDispatcher{
		outputChannels: make([]chan Message, len(initialSubscribers)),
	}

	dispatcher.mu.Lock()
	defer dispatcher.mu.Unlock()

	for _, ch := range initialSubscribers {
		dispatcher.outputChannels = append(dispatcher.outputChannels, ch)
	}

	return dispatcher
}

func (d *AudioDispatcher) Dispatch(msg Message) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	for _, ch := range d.outputChannels {
		// Utiliser une goroutine pour éviter de bloquer le dispatch si un canal est bloqué.
		go func(c chan Message) {
			c <- msg
		}(ch)
	}
}
