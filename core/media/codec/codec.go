package codec

import (
	"time"

	"github.com/Youngju-Heo/gomedia/core/media/av"
	"github.com/Youngju-Heo/gomedia/core/media/codec/fake"
)

// PCMUCodecData struct
type PCMUCodecData struct {
	typ av.CodecType
}

// Type func
func (codecData PCMUCodecData) Type() av.CodecType {
	return codecData.typ
}

// SampleRate func
func (codecData PCMUCodecData) SampleRate() int {
	return 8000
}

// ChannelLayout func
func (codecData PCMUCodecData) ChannelLayout() av.ChannelLayout {
	return av.ChMono
}

// SampleFormat func
func (codecData PCMUCodecData) SampleFormat() av.SampleFormat {
	return av.S16
}

// PacketDuration func
func (codecData PCMUCodecData) PacketDuration(data []byte) (time.Duration, error) {
	return time.Duration(len(data)) * time.Second / time.Duration(8000), nil
}

// NewPCMMulawCodecData func
func NewPCMMulawCodecData() av.AudioCodecData {
	return PCMUCodecData{
		typ: av.PCMU,
	}
}

// NewPCMAlawCodecData func
func NewPCMAlawCodecData() av.AudioCodecData {
	return PCMUCodecData{
		typ: av.PCMA,
	}
}

// SpeexCodecData struct
type SpeexCodecData struct {
	fake.CodecData
}

// PacketDuration func
func (codecData SpeexCodecData) PacketDuration(data []byte) (time.Duration, error) {
	// libavcodec/libspeexdec.c
	// samples = samplerate/50
	// duration = 0.02s
	return time.Millisecond * 20, nil
}

// // NewSpeexCodecData func
// func NewSpeexCodecData(sr int, cl av.ChannelLayout) SpeexCodecData {
// 	codec := SpeexCodecData{}
// 	codec.CodecTypeItem = av.SPEEX
// 	codec.SampleFormatItem = av.S16
// 	codec.SampleRateItem = sr
// 	codec.ChannelLayoutItem = cl
// 	return codec
// }
