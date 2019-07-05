// Package transcode implements Transcoder based on Muxer/Demuxer and AudioEncoder/AudioDecoder interface.
package transcode

import (
	"fmt"
	"time"

	"github.com/Youngju-Heo/gomedia/core/media/av"
	"github.com/Youngju-Heo/gomedia/core/media/av/pktque"
)

// Debug type
var Debug bool

type tStream struct {
	codec              av.CodecData
	timeline           *pktque.Timeline
	aencodec, adecodec av.AudioCodecData
	aenc               av.AudioEncoder
	adec               av.AudioDecoder
}

// Options struct
type Options struct {
	// check if transcode is needed, and create the AudioDecoder and AudioEncoder.
	FindAudioDecoderEncoder func(codec av.AudioCodecData, i int) (
		need bool, dec av.AudioDecoder, enc av.AudioEncoder, err error,
	)
}

// Transcoder struct
type Transcoder struct {
	streams []*tStream
}

// NewTranscoder func
func NewTranscoder(streams []av.CodecData, options Options) (_self *Transcoder, err error) {
	instance := &Transcoder{}
	instance.streams = []*tStream{}

	for i, stream := range streams {
		ts := &tStream{codec: stream}
		if stream.Type().IsAudio() {
			if options.FindAudioDecoderEncoder != nil {
				var ok bool
				var enc av.AudioEncoder
				var dec av.AudioDecoder
				ok, dec, enc, err = options.FindAudioDecoderEncoder(stream.(av.AudioCodecData), i)
				if ok {
					if err != nil {
						return
					}
					ts.timeline = &pktque.Timeline{}
					if ts.codec, err = enc.CodecData(); err != nil {
						return
					}
					ts.aencodec = ts.codec.(av.AudioCodecData)
					ts.adecodec = stream.(av.AudioCodecData)
					ts.aenc = enc
					ts.adec = dec
				}
			}
		}
		instance.streams = append(instance.streams, ts)
	}

	_self = instance
	return
}

func (instance *tStream) audioDecodeAndEncode(inpkt av.Packet) (outpkts []av.Packet, err error) {
	var dur time.Duration
	var frame av.AudioFrame
	var ok bool
	if ok, frame, err = instance.adec.Decode(inpkt.Data); err != nil {
		return
	}
	if !ok {
		return
	}

	if dur, err = instance.adecodec.PacketDuration(inpkt.Data); err != nil {
		err = fmt.Errorf("transcode: PacketDuration() failed for input stream #%d", inpkt.Idx)
		return
	}

	if Debug {
		fmt.Println("transcode: push", inpkt.Time, dur)
	}
	instance.timeline.Push(inpkt.Time, dur)

	var _outpkts [][]byte
	if _outpkts, err = instance.aenc.Encode(frame); err != nil {
		return
	}
	for _, _outpkt := range _outpkts {
		if dur, err = instance.aencodec.PacketDuration(_outpkt); err != nil {
			err = fmt.Errorf("transcode: PacketDuration() failed for output stream #%d", inpkt.Idx)
			return
		}
		outpkt := av.Packet{Idx: inpkt.Idx, Data: _outpkt}
		outpkt.Time = instance.timeline.Pop(dur)

		if Debug {
			fmt.Println("transcode: pop", outpkt.Time, dur)
		}

		outpkts = append(outpkts, outpkt)
	}

	return
}

// Do the transcode.
//
// In audio transcoding one Packet may transcode into many Packets
// packet time will be adjusted automatically.
func (instance *Transcoder) Do(pkt av.Packet) (out []av.Packet, err error) {
	stream := instance.streams[pkt.Idx]
	if stream.aenc != nil && stream.adec != nil {
		if out, err = stream.audioDecodeAndEncode(pkt); err != nil {
			return
		}
	} else {
		out = append(out, pkt)
	}
	return
}

// Streams Get CodecDatas after transcoding.
func (instance *Transcoder) Streams() (streams []av.CodecData, err error) {
	for _, stream := range instance.streams {
		streams = append(streams, stream.codec)
	}
	return
}

// Close transcoder, close related encoder and decoders.
func (instance *Transcoder) Close() (err error) {
	for _, stream := range instance.streams {
		if stream.aenc != nil {
			stream.aenc.Close()
			stream.aenc = nil
		}
		if stream.adec != nil {
			stream.adec.Close()
			stream.adec = nil
		}
	}
	instance.streams = nil
	return
}

// Muxer struct Wrap transcoder and origin Muxer into new Muxer.
// Write to new Muxer will do transcoding automatically.
type Muxer struct {
	av.Muxer   // origin Muxer
	Options    // transcode options
	transcoder *Transcoder
}

// WriteHeader func
func (instance *Muxer) WriteHeader(streams []av.CodecData) (err error) {
	if instance.transcoder, err = NewTranscoder(streams, instance.Options); err != nil {
		return
	}
	var newstreams []av.CodecData
	if newstreams, err = instance.transcoder.Streams(); err != nil {
		return
	}
	if err = instance.Muxer.WriteHeader(newstreams); err != nil {
		return
	}
	return
}

// WritePacket func
func (instance *Muxer) WritePacket(pkt av.Packet) (err error) {
	var outpkts []av.Packet
	if outpkts, err = instance.transcoder.Do(pkt); err != nil {
		return
	}
	for _, pkt := range outpkts {
		if err = instance.Muxer.WritePacket(pkt); err != nil {
			return
		}
	}
	return
}

// Close func
func (instance *Muxer) Close() (err error) {
	if instance.transcoder != nil {
		return instance.transcoder.Close()
	}
	return
}

// Demuxer struct Wrap transcoder and origin Demuxer into new Demuxer.
// Read this Demuxer will do transcoding automatically.
type Demuxer struct {
	av.Demuxer
	Options
	transcoder *Transcoder
	outpkts    []av.Packet
}

func (instance *Demuxer) prepare() (err error) {
	if instance.transcoder == nil {
		var streams []av.CodecData
		if streams, err = instance.Demuxer.Streams(); err != nil {
			return
		}
		if instance.transcoder, err = NewTranscoder(streams, instance.Options); err != nil {
			return
		}
	}
	return
}

// ReadPacket func
func (instance *Demuxer) ReadPacket() (pkt av.Packet, err error) {
	if err = instance.prepare(); err != nil {
		return
	}
	for {
		if len(instance.outpkts) > 0 {
			pkt = instance.outpkts[0]
			instance.outpkts = instance.outpkts[1:]
			return
		}
		var rpkt av.Packet
		if rpkt, err = instance.Demuxer.ReadPacket(); err != nil {
			return
		}
		if instance.outpkts, err = instance.transcoder.Do(rpkt); err != nil {
			return
		}
	}
	// return
}

// Streams func
func (instance *Demuxer) Streams() (streams []av.CodecData, err error) {
	if err = instance.prepare(); err != nil {
		return
	}
	return instance.transcoder.Streams()
}

// Close func
func (instance *Demuxer) Close() (err error) {
	if instance.transcoder != nil {
		return instance.transcoder.Close()
	}
	return
}
