package ffmpeg

/*
#cgo CFLAGS: -I../../../deps/include
#include "ffmpeg.h"
*/
import "C"
import (
	"fmt"
	"time"
	"unsafe"

	"github.com/Youngju-Heo/gomedia/core/media/av"
	"github.com/Youngju-Heo/gomedia/core/media/av/avutil"
)

func channelLayoutFF2AV(layout C.uint64_t) (channelLayout av.ChannelLayout) {
	if layout&C.AV_CH_FRONT_CENTER != 0 {
		channelLayout |= av.ChFrontCenter
	}
	if layout&C.AV_CH_FRONT_LEFT != 0 {
		channelLayout |= av.ChFrontLeft
	}
	if layout&C.AV_CH_FRONT_RIGHT != 0 {
		channelLayout |= av.ChFrontRight
	}
	if layout&C.AV_CH_BACK_CENTER != 0 {
		channelLayout |= av.ChBackCenter
	}
	if layout&C.AV_CH_BACK_LEFT != 0 {
		channelLayout |= av.ChBackLeft
	}
	if layout&C.AV_CH_BACK_RIGHT != 0 {
		channelLayout |= av.ChBackRight
	}
	if layout&C.AV_CH_SIDE_LEFT != 0 {
		channelLayout |= av.ChSideLeft
	}
	if layout&C.AV_CH_SIDE_RIGHT != 0 {
		channelLayout |= av.ChSideRight
	}
	if layout&C.AV_CH_LOW_FREQUENCY != 0 {
		channelLayout |= av.ChLowFreq
	}
	return
}

func channelLayoutAV2FF(channelLayout av.ChannelLayout) (layout C.uint64_t) {
	if channelLayout&av.ChFrontCenter != 0 {
		layout |= C.AV_CH_FRONT_CENTER
	}
	if channelLayout&av.ChFrontLeft != 0 {
		layout |= C.AV_CH_FRONT_LEFT
	}
	if channelLayout&av.ChFrontRight != 0 {
		layout |= C.AV_CH_FRONT_RIGHT
	}
	if channelLayout&av.ChBackCenter != 0 {
		layout |= C.AV_CH_BACK_CENTER
	}
	if channelLayout&av.ChBackLeft != 0 {
		layout |= C.AV_CH_BACK_LEFT
	}
	if channelLayout&av.ChBackRight != 0 {
		layout |= C.AV_CH_BACK_RIGHT
	}
	if channelLayout&av.ChSideLeft != 0 {
		layout |= C.AV_CH_SIDE_LEFT
	}
	if channelLayout&av.ChSideRight != 0 {
		layout |= C.AV_CH_SIDE_RIGHT
	}
	if channelLayout&av.ChLowFreq != 0 {
		layout |= C.AV_CH_LOW_FREQUENCY
	}
	return
}

func sampleFormatAV2FF(sampleFormat av.SampleFormat) (ffsamplefmt int32) {
	switch sampleFormat {
	case av.U8:
		ffsamplefmt = C.AV_SAMPLE_FMT_U8
	case av.S16:
		ffsamplefmt = C.AV_SAMPLE_FMT_S16
	case av.S32:
		ffsamplefmt = C.AV_SAMPLE_FMT_S32
	case av.FLT:
		ffsamplefmt = C.AV_SAMPLE_FMT_FLT
	case av.DBL:
		ffsamplefmt = C.AV_SAMPLE_FMT_DBL
	case av.U8P:
		ffsamplefmt = C.AV_SAMPLE_FMT_U8P
	case av.S16P:
		ffsamplefmt = C.AV_SAMPLE_FMT_S16P
	case av.S32P:
		ffsamplefmt = C.AV_SAMPLE_FMT_S32P
	case av.FLTP:
		ffsamplefmt = C.AV_SAMPLE_FMT_FLTP
	case av.DBLP:
		ffsamplefmt = C.AV_SAMPLE_FMT_DBLP
	}
	return
}

func sampleFormatFF2AV(ffsamplefmt int32) (sampleFormat av.SampleFormat) {
	switch ffsamplefmt {
	case C.AV_SAMPLE_FMT_U8: ///< unsigned 8 bits
		sampleFormat = av.U8
	case C.AV_SAMPLE_FMT_S16: ///< signed 16 bits
		sampleFormat = av.S16
	case C.AV_SAMPLE_FMT_S32: ///< signed 32 bits
		sampleFormat = av.S32
	case C.AV_SAMPLE_FMT_FLT: ///< float
		sampleFormat = av.FLT
	case C.AV_SAMPLE_FMT_DBL: ///< double
		sampleFormat = av.DBL
	case C.AV_SAMPLE_FMT_U8P: ///< unsigned 8 bits, planar
		sampleFormat = av.U8P
	case C.AV_SAMPLE_FMT_S16P: ///< signed 16 bits, planar
		sampleFormat = av.S16P
	case C.AV_SAMPLE_FMT_S32P: ///< signed 32 bits, planar
		sampleFormat = av.S32P
	case C.AV_SAMPLE_FMT_FLTP: ///< float, planar
		sampleFormat = av.FLTP
	case C.AV_SAMPLE_FMT_DBLP: ///< double, planar
		sampleFormat = av.DBLP
	}
	return
}

func audioFrameAssignToAVParams(f *C.AVFrame, frame *av.AudioFrame) {
	frame.SampleFormat = sampleFormatFF2AV(int32(f.format))
	frame.ChannelLayout = channelLayoutFF2AV(f.channel_layout)
	frame.SampleRate = int(f.sample_rate)
}

func audioFrameAssignToAVData(f *C.AVFrame, frame *av.AudioFrame) {
	frame.SampleCount = int(f.nb_samples)
	frame.Data = make([][]byte, int(f.channels))
	for i := 0; i < int(f.channels); i++ {
		frame.Data[i] = C.GoBytes(unsafe.Pointer(f.data[i]), f.linesize[0])
	}
}

func audioFrameAssignToAV(f *C.AVFrame, frame *av.AudioFrame) {
	audioFrameAssignToAVParams(f, frame)
	audioFrameAssignToAVData(f, frame)
}

func audioFrameAssignToFFParams(frame av.AudioFrame, f *C.AVFrame) {
	f.format = C.int(sampleFormatAV2FF(frame.SampleFormat))
	f.channel_layout = channelLayoutAV2FF(frame.ChannelLayout)
	f.sample_rate = C.int(frame.SampleRate)
	f.channels = C.int(frame.ChannelLayout.Count())
}

func audioFrameAssignToFFData(frame av.AudioFrame, f *C.AVFrame) {
	f.nb_samples = C.int(frame.SampleCount)
	for i := range frame.Data {
		f.data[i] = (*C.uint8_t)(unsafe.Pointer(&frame.Data[i][0]))
		f.linesize[i] = C.int(len(frame.Data[i]))
	}
}
func audioFrameAssignToFF(frame av.AudioFrame, f *C.AVFrame) {
	audioFrameAssignToFFParams(frame, f)
	audioFrameAssignToFFData(frame, f)
}

type audioCodecData struct {
	codecID       uint32
	sampleFormat  av.SampleFormat
	channelLayout av.ChannelLayout
	sampleRate    int
	extradata     []byte
}

// Type func
func (instance audioCodecData) Type() av.CodecType {
	return av.MakeAudioCodecType(instance.codecID)
}

// SampleRate func
func (instance audioCodecData) SampleRate() int {
	return instance.sampleRate
}

// SampleFormat func
func (instance audioCodecData) SampleFormat() av.SampleFormat {
	return instance.sampleFormat
}

// ChannelLayout func
func (instance audioCodecData) ChannelLayout() av.ChannelLayout {
	return instance.channelLayout
}

// PacketDuration func
func (instance audioCodecData) PacketDuration(data []byte) (dur time.Duration, err error) {
	// TODO: implement it: ffmpeg get_audio_frame_duration
	err = fmt.Errorf("ffmpeg: cannot get packet duration")
	return
}

// AudioCodecHandler func
func AudioCodecHandler(h *avutil.RegisterHandler) {
	var dec av.AudioDecoder
	var err error
	h.AudioDecoder = func(codec av.AudioCodecData) (av.AudioDecoder, error) {
		if dec, err = NewAudioDecoder(codec); err != nil {
			return nil, nil
		}
		return dec, err
	}

	h.AudioEncoder = func(typ av.CodecType) (av.AudioEncoder, error) {
		var enc av.AudioEncoder
		var err error
		if enc, err = NewAudioEncoderByCodecType(typ); err != nil {
			return nil, nil
		}
		return enc, err
	}
}

// NewAudioEncoderByCodecType func
func NewAudioEncoderByCodecType(typ av.CodecType) (enc *AudioEncoder, err error) {
	var id uint32

	switch typ {
	case av.AAC:
		id = C.AV_CODEC_ID_AAC

	default:
		err = fmt.Errorf("ffmpeg: cannot find encoder codecType=%d", typ)
		return
	}

	codec := C.avcodec_find_encoder(id)
	if codec == nil || C.avcodec_get_type(id) != C.AVMEDIA_TYPE_AUDIO {
		err = fmt.Errorf("ffmpeg: cannot find audio encoder codecID=%d", id)
		return
	}

	_enc := &AudioEncoder{}
	if _enc.ff, err = newFFCtxByCodec(codec); err != nil {
		return
	}
	enc = _enc
	return
}

// NewAudioEncoderByName func
func NewAudioEncoderByName(name string) (enc *AudioEncoder, err error) {
	_enc := &AudioEncoder{}

	codec := C.avcodec_find_encoder_by_name(C.CString(name))
	if codec == nil || C.avcodec_get_type(codec.id) != C.AVMEDIA_TYPE_AUDIO {
		err = fmt.Errorf("ffmpeg: cannot find audio encoder name=%s", name)
		return
	}

	if _enc.ff, err = newFFCtxByCodec(codec); err != nil {
		return
	}
	enc = _enc
	return
}
