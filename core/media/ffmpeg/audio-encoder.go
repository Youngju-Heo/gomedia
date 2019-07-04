package ffmpeg

/*
#cgo CFLAGS: -I../../deps/include
#include "ffmpeg.h"

*/
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/Youngju-Heo/gomedia/core/media/av"
	"github.com/Youngju-Heo/gomedia/core/media/codec/aacparser"
)

const debug = false

// AudioEncoder struct
type AudioEncoder struct {
	ff               *ffctx
	SampleRate       int
	Bitrate          int
	ChannelLayout    av.ChannelLayout
	SampleFormat     av.SampleFormat
	FrameSampleCount int
	framebuf         av.AudioFrame
	codecData        av.AudioCodecData
	resampler        *Resampler
}

// SetSampleFormat func
func (encoder *AudioEncoder) SetSampleFormat(fmt av.SampleFormat) (err error) {
	encoder.SampleFormat = fmt
	return
}

// SetSampleRate func
func (encoder *AudioEncoder) SetSampleRate(rate int) (err error) {
	encoder.SampleRate = rate
	return
}

// SetChannelLayout func
func (encoder *AudioEncoder) SetChannelLayout(ch av.ChannelLayout) (err error) {
	encoder.ChannelLayout = ch
	return
}

// SetBitrate func
func (encoder *AudioEncoder) SetBitrate(bitrate int) (err error) {
	encoder.Bitrate = bitrate
	return
}

// SetOption func
func (encoder *AudioEncoder) SetOption(key string, val interface{}) (err error) {
	ff := &encoder.ff.ff

	sval := fmt.Sprint(val)
	if key == "profile" {
		ff.profile = C.avcodec_profile_name_to_int(ff.codec, C.CString(sval))
		if ff.profile == C.FF_PROFILE_UNKNOWN {
			err = fmt.Errorf("ffmpeg: profile `%s` invalid", sval)
			return
		}
		return
	}

	C.av_dict_set(&ff.options, C.CString(key), C.CString(sval), 0)
	return
}

// GetOption func
func (encoder *AudioEncoder) GetOption(key string, val interface{}) (err error) {
	ff := &encoder.ff.ff
	entry := C.av_dict_get(ff.options, C.CString(key), nil, 0)
	if entry == nil {
		err = fmt.Errorf("ffmpeg: GetOption failed: `%s` not exists", key)
		return
	}
	switch p := val.(type) {
	case *string:
		*p = C.GoString(entry.value)
	case *int:
		fmt.Sscanf(C.GoString(entry.value), "%d", p)
	default:
		err = fmt.Errorf("ffmpeg: GetOption failed: val must be *string or *int receiver")
		return
	}
	return
}

// Setup func
func (encoder *AudioEncoder) Setup() (err error) {
	ff := &encoder.ff.ff

	ff.frame = C.av_frame_alloc()

	if encoder.SampleFormat == av.SampleFormat(0) {
		encoder.SampleFormat = sampleFormatFF2AV(*ff.codec.sample_fmts)
	}

	//if encoder.Bitrate == 0 {
	//	encoder.Bitrate = 80000
	//}
	if encoder.SampleRate == 0 {
		encoder.SampleRate = 44100
	}
	if encoder.ChannelLayout == av.ChannelLayout(0) {
		encoder.ChannelLayout = av.ChStereo
	}

	ff.codecCtx.sample_fmt = sampleFormatAV2FF(encoder.SampleFormat)
	ff.codecCtx.sample_rate = C.int(encoder.SampleRate)
	ff.codecCtx.bit_rate = C.int64_t(encoder.Bitrate)
	ff.codecCtx.channel_layout = channelLayoutAV2FF(encoder.ChannelLayout)
	ff.codecCtx.strict_std_compliance = C.FF_COMPLIANCE_EXPERIMENTAL
	ff.codecCtx.flags = C.AV_CODEC_FLAG_GLOBAL_HEADER
	ff.codecCtx.profile = ff.profile

	if C.avcodec_open2(ff.codecCtx, ff.codec, nil) != 0 {
		err = fmt.Errorf("ffmpeg: encoder: avcodec_open2 failed")
		return
	}
	encoder.SampleFormat = sampleFormatFF2AV(ff.codecCtx.sample_fmt)
	encoder.FrameSampleCount = int(ff.codecCtx.frame_size)

	extradata := C.GoBytes(unsafe.Pointer(ff.codecCtx.extradata), ff.codecCtx.extradata_size)

	switch ff.codecCtx.codec_id {
	case C.AV_CODEC_ID_AAC:
		if encoder.codecData, err = aacparser.NewCodecDataFromMPEG4AudioConfigBytes(extradata); err != nil {
			return
		}

	default:
		encoder.codecData = audioCodecData{
			channelLayout: encoder.ChannelLayout,
			sampleFormat:  encoder.SampleFormat,
			sampleRate:    encoder.SampleRate,
			codecID:       ff.codecCtx.codec_id,
			extradata:     extradata,
		}
	}

	return
}

func (encoder *AudioEncoder) prepare() (err error) {
	ff := &encoder.ff.ff

	if ff.frame == nil {
		if err = encoder.Setup(); err != nil {
			return
		}
	}

	return
}

// CodecData func
func (encoder *AudioEncoder) CodecData() (codec av.AudioCodecData, err error) {
	if err = encoder.prepare(); err != nil {
		return
	}
	codec = encoder.codecData
	return
}

func (encoder *AudioEncoder) encodeOne(frame av.AudioFrame) (gotpkt bool, pkt []byte, err error) {
	if err = encoder.prepare(); err != nil {
		return
	}

	ff := &encoder.ff.ff

	cpkt := C.AVPacket{}
	audioFrameAssignToFF(frame, ff.frame)

	if false {
		farr := []string{}
		for i := 0; i < len(frame.Data[0])/4; i++ {
			var f *float64 = (*float64)(unsafe.Pointer(&frame.Data[0][i*4]))
			farr = append(farr, fmt.Sprintf("%.8f", *f))
		}
		fmt.Println(farr)
	}
	//cerr := C.avcodec_send_frame(ff.codecCtx, &cpkt, ff.frame, &cgotpkt)
	cerr := C.avcodec_send_frame(ff.codecCtx, ff.frame)
	if cerr < C.int(0) {
		err = fmt.Errorf("ffmpeg: avcodec_encode_audio2 failed: %d", cerr)
		return
	}
	if cerr >= C.int(0) {
		cerr = C.avcodec_receive_packet(ff.codecCtx, &cpkt)
		if cerr == (-C.EAGAIN) || cerr == C.AVERROR_EOF {
			return
		} else if cerr >= C.int(0) {

			gotpkt = true
			pkt = C.GoBytes(unsafe.Pointer(cpkt.data), cpkt.size)
			C.av_packet_unref(&cpkt)

			if debug {
				fmt.Println("ffmpeg: Encode", frame.SampleCount, frame.SampleRate, frame.ChannelLayout, frame.SampleFormat, "len", len(pkt))
			}
		}
	}

	return
}

func (encoder *AudioEncoder) resample(in av.AudioFrame) (out av.AudioFrame, err error) {
	if encoder.resampler == nil {
		encoder.resampler = &Resampler{
			OutSampleFormat:  encoder.SampleFormat,
			OutSampleRate:    encoder.SampleRate,
			OutChannelLayout: encoder.ChannelLayout,
		}
	}
	if out, err = encoder.resampler.Resample(in); err != nil {
		return
	}
	return
}

// Encode func
func (encoder *AudioEncoder) Encode(frame av.AudioFrame) (pkts [][]byte, err error) {
	var gotpkt bool
	var pkt []byte

	if frame.SampleFormat != encoder.SampleFormat || frame.ChannelLayout != encoder.ChannelLayout || frame.SampleRate != encoder.SampleRate {
		if frame, err = encoder.resample(frame); err != nil {
			return
		}
	}

	if encoder.FrameSampleCount != 0 {
		if encoder.framebuf.SampleCount == 0 {
			encoder.framebuf = frame
		} else {
			encoder.framebuf = encoder.framebuf.Concat(frame)
		}
		for encoder.framebuf.SampleCount >= encoder.FrameSampleCount {
			frame := encoder.framebuf.Slice(0, encoder.FrameSampleCount)
			if gotpkt, pkt, err = encoder.encodeOne(frame); err != nil {
				return
			}
			if gotpkt {
				pkts = append(pkts, pkt)
			}
			encoder.framebuf = encoder.framebuf.Slice(encoder.FrameSampleCount, encoder.framebuf.SampleCount)
		}
	} else {
		if gotpkt, pkt, err = encoder.encodeOne(frame); err != nil {
			return
		}
		if gotpkt {
			pkts = append(pkts, pkt)
		}
	}

	return
}

// Close func
func (encoder *AudioEncoder) Close() {
	freeFFCtx(encoder.ff)
	if encoder.resampler != nil {
		encoder.resampler.Close()
		encoder.resampler = nil
	}
}
