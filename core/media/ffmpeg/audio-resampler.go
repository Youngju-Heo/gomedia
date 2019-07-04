package ffmpeg

/*
#cgo CFLAGS: -I../../../deps/include
#include "ffmpeg.h"
int resample_convert(SwrContext *ctx, int *out, int outcount, int *in, int incount) {
	return swr_convert(ctx, (void *)out, outcount, (void *)in, incount);
}

*/
import "C"
import (
	"fmt"
	"runtime"
	"unsafe"

	"github.com/Youngju-Heo/gomedia/core/media/av"
)

// Resampler struct
type Resampler struct {
	inSampleFormat, OutSampleFormat   av.SampleFormat
	inChannelLayout, OutChannelLayout av.ChannelLayout
	inSampleRate, OutSampleRate       int
	ctx                               *C.SwrContext
}

// NewResampler create new audio resampler
func NewResampler(fmt av.SampleFormat, layout av.ChannelLayout, rate int) (*Resampler, error) {

	return &Resampler{
		OutSampleFormat:  fmt,
		OutChannelLayout: layout,
		OutSampleRate:    rate,
	}, nil
}

// Resample func
func (sampler *Resampler) Resample(in av.AudioFrame) (out av.AudioFrame, err error) {

	var inSampleFormat = in.SampleFormat
	var inChannelLayout = in.ChannelLayout
	var inSampleRate = in.SampleRate

	var flush av.AudioFrame

	// same situation: no need change
	formatChange := sampler.inSampleFormat != inSampleFormat || sampler.inChannelLayout != inChannelLayout || sampler.inSampleRate != inSampleRate || sampler.ctx == nil

	// need reset resampler
	if formatChange {
		// have already processing
		if sampler.ctx != nil {
			outChannels := sampler.OutChannelLayout.Count()
			if sampler.OutSampleFormat.IsPlanar() {
				outChannels = 1
			}
			outData := make([]*C.uint8_t, outChannels)
			outSampleCount := int(C.swr_get_out_samples(sampler.ctx, C.int(in.SampleCount)))
			outLinesize := outSampleCount * sampler.OutSampleFormat.BytesPerSample()
			flush.Data = make([][]byte, outChannels)
			for i := 0; i < outChannels; i++ {
				flush.Data[i] = make([]byte, outLinesize)
				outData[i] = (*C.uint8_t)(unsafe.Pointer(&flush.Data[i][0]))
			}
			flush.ChannelLayout = sampler.OutChannelLayout
			flush.SampleFormat = sampler.OutSampleFormat
			flush.SampleRate = sampler.OutSampleRate

			convertSamples := int(C.resample_convert(sampler.ctx,
				(*C.int)(unsafe.Pointer(&outData[0])), C.int(outSampleCount), nil, C.int(0)))

			if convertSamples < 0 {
				err = fmt.Errorf("swr_convert failed")
				return
			}
			flush.SampleCount = convertSamples
			if convertSamples < outSampleCount {
				for i := 0; i < outChannels; i++ {
					flush.Data[i] = flush.Data[i][:convertSamples*sampler.OutSampleFormat.BytesPerSample()]
				}
			}
			C.swr_free(&sampler.ctx)
		} else {
			runtime.SetFinalizer(sampler, func(sampler *Resampler) {
				sampler.Close()
			})
		}

		ctx := C.swr_alloc()
		if ctx == nil {
			return
		}

		C.av_opt_set_int(unsafe.Pointer(ctx), C.CString("in_channel_layout"), C.int64_t(channelLayoutAV2FF(inChannelLayout)), 0)
		C.av_opt_set_int(unsafe.Pointer(ctx), C.CString("out_channel_layout"), C.int64_t(channelLayoutAV2FF(sampler.OutChannelLayout)), 0)
		C.av_opt_set_int(unsafe.Pointer(ctx), C.CString("in_sample_rate"), C.int64_t(inSampleRate), 0)
		C.av_opt_set_int(unsafe.Pointer(ctx), C.CString("out_sample_rate"), C.int64_t(sampler.OutSampleRate), 0)
		C.av_opt_set_int(unsafe.Pointer(ctx), C.CString("in_sample_fmt"), C.int64_t(sampleFormatAV2FF(inSampleFormat)), 0)
		C.av_opt_set_int(unsafe.Pointer(ctx), C.CString("out_sample_fmt"), C.int64_t(sampleFormatAV2FF(sampler.OutSampleFormat)), 0)

		if C.swr_init(ctx) < C.int(0) {
			C.swr_free(&ctx)
			return
		}
		sampler.ctx = ctx
		sampler.inSampleFormat = inSampleFormat
		sampler.inChannelLayout = inChannelLayout
		sampler.inSampleRate = inSampleRate
	}

	var inChannels int
	inSampleCount := in.SampleCount
	if !sampler.inSampleFormat.IsPlanar() {
		inChannels = 1
	} else {
		inChannels = sampler.inChannelLayout.Count()
	}
	inData := make([]*C.uint8_t, inChannels)
	for i := 0; i < inChannels; i++ {
		inData[i] = (*C.uint8_t)(unsafe.Pointer(&in.Data[i][0]))
	}

	var outChannels, outLinesize, outBytesPerSample int
	outSampleCount := int(C.swr_get_out_samples(sampler.ctx, C.int(in.SampleCount)))
	if !sampler.OutSampleFormat.IsPlanar() {
		outChannels = 1
		outBytesPerSample = sampler.OutSampleFormat.BytesPerSample() * sampler.OutChannelLayout.Count()
		outLinesize = outSampleCount * outBytesPerSample
	} else {
		outChannels = sampler.OutChannelLayout.Count()
		outBytesPerSample = sampler.OutSampleFormat.BytesPerSample()
		outLinesize = outSampleCount * outBytesPerSample
	}
	outData := make([]*C.uint8_t, outChannels)
	out.Data = make([][]byte, outChannels)
	for i := 0; i < outChannels; i++ {
		out.Data[i] = make([]byte, outLinesize)
		outData[i] = (*C.uint8_t)(unsafe.Pointer(&out.Data[i][0]))
	}
	out.ChannelLayout = sampler.OutChannelLayout
	out.SampleFormat = sampler.OutSampleFormat
	out.SampleRate = sampler.OutSampleRate

	convertSamples := int(C.resample_convert(
		sampler.ctx,
		(*C.int)(unsafe.Pointer(&outData[0])), C.int(outSampleCount),
		(*C.int)(unsafe.Pointer(&inData[0])), C.int(inSampleCount),
	))
	if convertSamples < 0 {
		err = fmt.Errorf("ffmpeg: avresample_convert_frame failed")
		return
	}

	out.SampleCount = convertSamples
	if convertSamples < outSampleCount {
		for i := 0; i < outChannels; i++ {
			out.Data[i] = out.Data[i][:convertSamples*outBytesPerSample]
		}
	}

	if flush.SampleCount > 0 {
		out = flush.Concat(out)
	}

	return
}

// Close func
func (sampler *Resampler) Close() {
	if sampler.ctx != nil {
		C.swr_free(&sampler.ctx)
	}
}
