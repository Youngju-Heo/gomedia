package device

import (
	"fmt"
	"log"

	"github.com/Youngju-Heo/gomedia/core/media/av"
	"github.com/Youngju-Heo/gomedia/core/media/ffmpeg"
)

// PacketDecoder audio packet decoder
type PacketDecoder struct {
	// primitive item
	srcCodec     av.AudioCodecData
	dstLayout    av.ChannelLayout
	dstFrequency int

	// generated item
	decoder   *ffmpeg.AudioDecoder
	resampler *ffmpeg.Resampler
	packets   chan av.AudioFrame
	current   []byte
}

// NewPacketDecoder make new packet decoder
func NewPacketDecoder(srcCodec av.AudioCodecData,
	dstLayout av.ChannelLayout,
	dstFrequency int,
	bufferCount int,
) *PacketDecoder {
	return &PacketDecoder{
		srcCodec:     srcCodec,
		dstLayout:    dstLayout,
		dstFrequency: dstFrequency,
		packets:      make(chan av.AudioFrame, bufferCount),
	}
}

// Initialize initialize decoder
func (decoder *PacketDecoder) Initialize() (err error) {

	if decoder.decoder, err = ffmpeg.NewAudioDecoder(decoder.srcCodec); err != nil {
		log.Println("codec initialize failed", err)
		return
	}

	// setup
	if err = decoder.decoder.Setup(); err != nil {
		log.Println("decoder setup failed", err)
		return
	}

	return
}

// Decode decode packet
func (decoder *PacketDecoder) Decode(pkt av.Packet) (err error) {

	if decoder.decoder == nil {
		return fmt.Errorf("decoder not initialized")
	}

	var hasFrame bool
	var frame av.AudioFrame

	if hasFrame, frame, err = decoder.decoder.Decode(pkt.Data); err != nil {
		log.Println("decode failed", err)
		decoder.decoder.Close()
		decoder.Initialize()
		return
	}

	// no frame
	if !hasFrame {
		return
	}

	if frame.SampleFormat != av.S16 ||
		frame.SampleRate != decoder.dstFrequency ||
		frame.ChannelLayout != decoder.dstLayout {
		// need resampler
		if decoder.resampler == nil {
			if decoder.resampler, err = ffmpeg.NewResampler(av.S16, decoder.dstLayout, decoder.dstFrequency); err != nil {
				log.Println("resampler init failed", err)
				return
			}
		}

		// perform resample
		if frame, err = decoder.resampler.Resample(frame); err != nil {
			log.Println("audio resample failed", err)
			return
		}

		// add decoded to buffer
		decoder.packets <- frame
	}

	return
}

// Frequency freq
func (decoder *PacketDecoder) Frequency() int {
	return decoder.dstFrequency
}

// Format sample format
func (decoder *PacketDecoder) Format() av.SampleFormat {
	return av.S16
}

// Channels channel layout
func (decoder *PacketDecoder) Channels() av.ChannelLayout {
	return decoder.dstLayout
}

// RecommendSampleCount recommended sample count
func (decoder *PacketDecoder) RecommendSampleCount() (samples int) {
	switch decoder.srcCodec.Type() {
	case av.AAC:
		samples = 1024
	case av.PCMU, av.PCMA:
		samples = 640
	default:
		samples = 1024
	}
	return
}

// Process process buffer
func (decoder *PacketDecoder) Process(byteLength int) (rslt []byte) {

	var remain = byteLength

	for remain > 0 {

		var prepare = 0
		if decoder.current != nil {
			prepare = len(decoder.current)
		}

		if prepare < remain {
			recv := <-decoder.packets
			if decoder.current != nil && len(decoder.current) > 0 {
				decoder.current = append(decoder.current, recv.Data[0]...)
			} else {
				decoder.current = recv.Data[0]
			}
			continue
		} else {
			if prepare == remain {
				rslt = decoder.current
				decoder.current = nil
			} else {
				rslt = decoder.current[0:remain]
				decoder.current = decoder.current[remain:]
			}
			remain = 0
		}
	}

	return
}
