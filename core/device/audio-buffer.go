package device

import (
	"github.com/Youngju-Heo/gomedia/core/media/av"
)

// AudioBuffer audio play buffer
type AudioBuffer struct {
	frequency     int
	samples       int
	currentBuffer []byte
	bufferChannel chan []byte
}

// NewAudioBuffer create new audio buffer
func NewAudioBuffer(bufferCount int,
	frequency int,
	samples int,
) *AudioBuffer {

	return &AudioBuffer{
		frequency:     frequency,
		samples:       samples,
		currentBuffer: make([]byte, 0),
		bufferChannel: make(chan []byte, bufferCount),
	}
}

// Frequency freq
func (buffer *AudioBuffer) Frequency() int {
	return buffer.frequency
}

// Format sample format
func (buffer *AudioBuffer) Format() av.SampleFormat {
	return av.S16
}

// Channels channel layout
func (buffer *AudioBuffer) Channels() av.ChannelLayout {
	return av.ChStereo
}

// RecommendSampleCount recommended sample count
func (buffer *AudioBuffer) RecommendSampleCount() int {
	return buffer.samples
}

// Process process buffer
func (buffer *AudioBuffer) Process(byteLength int) (rslt []byte) {

	var remain = byteLength

	for remain > 0 {

		if prepare := len(buffer.currentBuffer); prepare < remain {
			recv := <-buffer.bufferChannel
			buffer.currentBuffer = append(buffer.currentBuffer, recv...)
			continue
		} else {
			if prepare == remain {
				rslt = buffer.currentBuffer
				buffer.currentBuffer = make([]byte, 0)
			} else {
				rslt = buffer.currentBuffer[0:remain]
				buffer.currentBuffer = buffer.currentBuffer[remain:]
			}
			remain = 0
		}
	}

	return
}

// AddBuffer add date to buffer
func (buffer *AudioBuffer) AddBuffer(data []byte) {
	buffer.bufferChannel <- data
}
