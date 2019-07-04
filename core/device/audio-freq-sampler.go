package device

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/Youngju-Heo/gomedia/core/media/av"
)

const (
	twoPi = math.Pi * 2
)

var littleEndian = binary.LittleEndian

// AudioFrequencySampler generate sampling audio
type AudioFrequencySampler struct {
	layout        av.ChannelLayout
	baseFrequency int
	format        av.SampleFormat
	samples       int
	frequencyData []int
	frequencyBase []int
	bytePerPeriod []int
}

// NewFrequencySampler create new sampler
func NewFrequencySampler(baseFrequency int, layout av.ChannelLayout, freqInfo []int, samples int) (*AudioFrequencySampler, error) {

	channels := layout.Count()
	if channels != len(freqInfo) {
		return nil, fmt.Errorf("layout vs freqInfo mismatch")
	}

	sampler := &AudioFrequencySampler{
		layout:        layout,
		baseFrequency: baseFrequency,
		format:        av.S16,
		samples:       samples,
	}

	sampler.frequencyData = make([]int, channels)

	copy(sampler.frequencyData, freqInfo)
	sampler.bytePerPeriod = make([]int, channels)
	sampler.frequencyBase = make([]int, channels)

	for idx, freq := range freqInfo {
		if freq == 0 {
			sampler.bytePerPeriod[idx] = 0
		} else {
			sampler.bytePerPeriod[idx] = baseFrequency / freq
		}
	}

	return sampler, nil
}

// Process process sampler
func (sampler *AudioFrequencySampler) Process(len int) (rslt []byte) {

	if rslt = make([]byte, len); rslt != nil {
		idx := 0
		for idx < len {
			for c := 0; c < sampler.layout.Count(); c++ {
				if sampler.bytePerPeriod[c] == 0 {
					sampler.frequencyBase[c] = 0
				} else {
					v := int16(32767.0 * math.Sin(float64(sampler.frequencyBase[c])*twoPi/float64(sampler.bytePerPeriod[c])))
					littleEndian.PutUint16(rslt[idx:], uint16(v))
					idx += 2
					sampler.frequencyBase[c] = (sampler.frequencyBase[c] + 1) % sampler.bytePerPeriod[c]
				}
			}
		}
	}

	return
}

// Frequency base frequency
func (sampler *AudioFrequencySampler) Frequency() int {
	return sampler.baseFrequency
}

// Format audio format
func (sampler *AudioFrequencySampler) Format() av.SampleFormat {
	return sampler.format
}

// Channels return channel layout
func (sampler *AudioFrequencySampler) Channels() av.ChannelLayout {
	return sampler.layout
}

// RecommendSampleCount sampling count
func (sampler *AudioFrequencySampler) RecommendSampleCount() int {
	return sampler.samples
}
