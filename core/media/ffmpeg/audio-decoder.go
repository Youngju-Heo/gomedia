package ffmpeg

/*
#cgo CFLAGS: -I../../../deps/include
#include "ffmpeg.h"
int decode_audio(AVCodecContext *ctx, AVFrame *frame, void *data, int size, int *got) {
  struct AVPacket pkt = {.data = data, .size = size};
  int result;
  *got = 0;
  if ((result = avcodec_send_packet(ctx, &pkt)) == 0 && (result = avcodec_receive_frame(ctx, frame)) == 0) {
    *got = 1;
  }

  return result;
}

*/
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/Youngju-Heo/gomedia/core/media/av"
	"github.com/Youngju-Heo/gomedia/core/media/codec/aacparser"
)

// AudioDecoder type
type AudioDecoder struct {
	ff            *ffctx
	ChannelLayout av.ChannelLayout
	SampleFormat  av.SampleFormat
	SampleRate    int
	Extradata     []byte
}

// NewAudioDecoder func
func NewAudioDecoder(codec av.AudioCodecData) (dec *AudioDecoder, err error) {
	_dec := &AudioDecoder{}
	var id uint32

	switch codec.Type() {
	case av.AAC:
		if aaccodec, ok := codec.(aacparser.CodecData); ok {
			_dec.Extradata = aaccodec.MPEG4AudioConfigBytes()
			id = C.AV_CODEC_ID_AAC
		} else {
			err = fmt.Errorf("ffmpeg: aac CodecData must be aacparser.CodecData")
			return
		}

	// case av.SPEEX:
	// 	id = C.AV_CODEC_ID_SPEEX

	case av.PCMU:
		id = C.AV_CODEC_ID_PCM_MULAW

	case av.PCMA:
		id = C.AV_CODEC_ID_PCM_ALAW

	default:
		if ffcodec, ok := codec.(audioCodecData); ok {
			_dec.Extradata = ffcodec.extradata
			id = ffcodec.codecID
		} else {
			err = fmt.Errorf("ffmpeg: invalid CodecData for ffmpeg to decode")
			return
		}
	}

	c := C.avcodec_find_decoder(id)
	if c == nil || C.avcodec_get_type(c.id) != C.AVMEDIA_TYPE_AUDIO {
		err = fmt.Errorf("ffmpeg: cannot find audio decoder id=%d", id)
		return
	}

	if _dec.ff, err = newFFCtxByCodec(c); err != nil {
		return
	}

	_dec.SampleFormat = codec.SampleFormat()
	_dec.SampleRate = codec.SampleRate()
	_dec.ChannelLayout = codec.ChannelLayout()
	if err = _dec.Setup(); err != nil {
		return
	}

	dec = _dec
	return
}

// Setup audio codec
func (decoder *AudioDecoder) Setup() error {
	ff := &decoder.ff.ff

	ff.frame = C.av_frame_alloc()
	if len(decoder.Extradata) > 0 {
		ff.codecCtx.extradata = (*C.uint8_t)(unsafe.Pointer(&decoder.Extradata[0]))
		ff.codecCtx.extradata_size = C.int(len(decoder.Extradata))
	}

	ff.codecCtx.sample_rate = C.int(decoder.SampleRate)
	ff.codecCtx.channel_layout = channelLayoutAV2FF(decoder.ChannelLayout)
	ff.codecCtx.channels = C.int(decoder.ChannelLayout.Count())
	if C.avcodec_open2(ff.codecCtx, ff.codec, nil) != 0 {
		return fmt.Errorf("decoder: avcodec open failed")
	}

	decoder.SampleFormat = sampleFormatFF2AV(ff.codecCtx.sample_fmt)
	decoder.ChannelLayout = channelLayoutFF2AV(ff.codecCtx.channel_layout)
	if decoder.SampleRate == 0 {
		decoder.SampleRate = int(ff.codecCtx.sample_rate)
	}

	return nil
}

// Decode frame
func (decoder *AudioDecoder) Decode(pkt []byte) (bool, av.AudioFrame, error) {
	ff := &decoder.ff.ff
	var frame av.AudioFrame
	var gotFrame bool

	cgotFrame := C.int(0)
	cerr := C.decode_audio(ff.codecCtx, ff.frame, unsafe.Pointer(&pkt[0]), C.int(len(pkt)), &cgotFrame)
	if cerr < C.int(0) {
		msg := GetFFErrorMessage(int(cerr))
		return false, frame, fmt.Errorf("%08x/%v", uint32(C.int(cerr)), msg)
	}

	if cgotFrame != C.int(0) {
		gotFrame = true
		audioFrameAssignToAV(ff.frame, &frame)
		frame.SampleRate = decoder.SampleRate
	}

	return gotFrame, frame, nil
}

// Close decoder
func (decoder *AudioDecoder) Close() {
	freeFFCtx(decoder.ff)
}
