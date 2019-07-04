package device

import (
	"fmt"
	"log"
	"time"

	"github.com/Youngju-Heo/gomedia/core/media/av"
	"github.com/Youngju-Heo/gomedia/core/media/ffmpeg"
)

// PacketTranscoder struct for transcoder
type PacketTranscoder struct {
	adecodec av.AudioCodecData
	aencodec av.AudioCodecData

	decoder       av.AudioDecoder
	encoder       av.AudioEncoder
	needTranscode bool
}

var debug = false

// NewTranscoder create new transcoder instance
func NewTranscoder(src av.AudioCodecData,
	forceDecoding bool,
	codec av.CodecType,
	sampleRate int,
	layout av.ChannelLayout,
	bitrate int,
	options *map[string]interface{},
) (trans *PacketTranscoder, err error) {

	if src.Type() == codec &&
		src.SampleRate() == sampleRate &&
		src.ChannelLayout() == layout && !forceDecoding {
		// same codec  do not transcoder
		trans = &PacketTranscoder{
			adecodec:      src,
			aencodec:      src,
			needTranscode: false,
		}
	} else {
		// need
		var decoder av.AudioDecoder
		var encoder av.AudioEncoder
		if decoder, err = ffmpeg.NewAudioDecoder(src); err != nil {
			return
		}
		if encoder, err = ffmpeg.NewAudioEncoderByName(codec.CodecName()); err != nil {
			return
		}
		encoder.SetSampleRate(sampleRate)
		encoder.SetChannelLayout(layout)
		encoder.SetBitrate(bitrate)

		if options != nil {
			for k, v := range *options {
				encoder.SetOption(k, v)
			}
		}
		var dst av.AudioCodecData
		if dst, err = encoder.CodecData(); err != nil {
			return
		}

		trans = &PacketTranscoder{
			adecodec:      src,
			aencodec:      dst,
			decoder:       decoder,
			encoder:       encoder,
			needTranscode: true,
		}

		log.Println("need transcode")
	}

	return
}

// Setup Transcoder setup
func (trans *PacketTranscoder) Setup() (err error) {
	return
}

// OutCodecData Output AudioCodecData
func (trans *PacketTranscoder) OutCodecData() (codecData av.AudioCodecData) {
	return trans.aencodec
}

// Do the transcode
func (trans *PacketTranscoder) Do(src av.Packet) (out []av.Packet, err error) {
	if !trans.needTranscode {
		out = []av.Packet{src}
	} else {
		if out, err = trans.decodeAndEncode(src); err != nil {
			return
		}
	}
	return
}

// Close close transcoder
func (trans *PacketTranscoder) Close() (err error) {
	if trans.decoder != nil {
		trans.decoder.Close()
		trans.decoder = nil
	}

	if trans.encoder != nil {
		trans.encoder.Close()
		trans.encoder = nil
	}

	return
}

func (trans *PacketTranscoder) decodeAndEncode(src av.Packet) (out []av.Packet, err error) {
	var dur time.Duration
	var frame av.AudioFrame
	var ok bool
	if ok, frame, err = trans.decoder.Decode(src.Data); err != nil {
		return
	}
	if !ok {
		return
	}

	if dur, err = trans.adecodec.PacketDuration(src.Data); err != nil {
		err = fmt.Errorf("transcode: PacketDuration() failed for input stream #%d", src.Idx)
		return
	}

	if debug {
		fmt.Println("transcode: push", src.Time, dur)
		log.Println("transcode push", src.Time, dur)
	}

	var encodedPkts [][]byte
	if encodedPkts, err = trans.encoder.Encode(frame); err != nil {
		return
	}
	for _, encPkt := range encodedPkts {
		if dur, err = trans.aencodec.PacketDuration(encPkt); err != nil {
			err = fmt.Errorf("transcode: PacketDuration() failed for output stream #%d", src.Idx)
			return
		}
		outpkt := av.Packet{Idx: src.Idx, Data: encPkt}
		outpkt.Time = src.Time

		if debug {
			fmt.Println("transcode: pop ", outpkt.Time, dur)
		}

		out = append(out, outpkt)
	}

	return
}
