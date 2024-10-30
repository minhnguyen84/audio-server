package audio

import "github.com/go-audio/audio"

type Audio struct {
	SessionId string
	Data      struct {
		internal *audio.IntBuffer
		external *audio.IntBuffer
	}
}
