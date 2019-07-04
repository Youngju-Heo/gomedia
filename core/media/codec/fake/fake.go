package fake

import (
	"github.com/Youngju-Heo/gomedia/core/media/av"
)

// CodecData struct
type CodecData struct {
	CodecTypeItem     av.CodecType
	SampleRateItem    int
	SampleFormatItem  av.SampleFormat
	ChannelLayoutItem av.ChannelLayout
}

// Type func
func (inst CodecData) Type() av.CodecType {
	return inst.CodecTypeItem
}

// SampleFormat func
func (inst CodecData) SampleFormat() av.SampleFormat {
	return inst.SampleFormatItem
}

// ChannelLayout func
func (inst CodecData) ChannelLayout() av.ChannelLayout {
	return inst.ChannelLayoutItem
}

// SampleRate func
func (inst CodecData) SampleRate() int {
	return inst.SampleRateItem
}
