package flv

import (
	"bufio"
	"fmt"
	"io"

	"github.com/Youngju-Heo/gomedia/core/media/av"
	"github.com/Youngju-Heo/gomedia/core/media/av/avutil"
	"github.com/Youngju-Heo/gomedia/core/media/codec/aacparser"
	"github.com/Youngju-Heo/gomedia/core/media/codec/h264parser"
	"github.com/Youngju-Heo/gomedia/core/media/format/flv/flvio"
	"github.com/Youngju-Heo/gomedia/core/media/utils/bits/pio"
)

// MaxProbePacketCount const
var MaxProbePacketCount = 20

// NewMetadataByStreams func
func NewMetadataByStreams(streams []av.CodecData) (metadata flvio.AMFMap, err error) {
	metadata = flvio.AMFMap{}

	for _, _stream := range streams {
		typ := _stream.Type()
		switch {
		case typ.IsVideo():
			stream := _stream.(av.VideoCodecData)
			switch typ {
			case av.H264:
				metadata["videocodecid"] = flvio.VideoH264

			default:
				err = fmt.Errorf("flv: metadata: unsupported video codecType=%v", stream.Type())
				return
			}

			metadata["width"] = stream.Width()
			metadata["height"] = stream.Height()
			metadata["displayWidth"] = stream.Width()
			metadata["displayHeight"] = stream.Height()

		case typ.IsAudio():
			stream := _stream.(av.AudioCodecData)
			switch typ {
			case av.AAC:
				metadata["audiocodecid"] = flvio.SoundAAC

			// case av.SPEEX:
			// 	metadata["audiocodecid"] = flvio.SoundSpeex

			default:
				err = fmt.Errorf("flv: metadata: unsupported audio codecType=%v", stream.Type())
				return
			}

			metadata["audiosamplerate"] = stream.SampleRate()
		}
	}

	return
}

// Prober struct
type Prober struct {
	HasAudio, HasVideo             bool
	GotAudio, GotVideo             bool
	VideoStreamIdx, AudioStreamIdx int
	PushedCount                    int
	Streams                        []av.CodecData
	CachedPkts                     []av.Packet
}

// CacheTag func
func (inst *Prober) CacheTag(_tag flvio.Tag, timestamp int32) {
	pkt, _ := inst.TagToPacket(_tag, timestamp)
	inst.CachedPkts = append(inst.CachedPkts, pkt)
}

// PushTag func
func (inst *Prober) PushTag(tag flvio.Tag, timestamp int32) (err error) {
	inst.PushedCount++

	if inst.PushedCount > MaxProbePacketCount {
		err = fmt.Errorf("flv: max probe packet count reached")
		return
	}

	switch tag.Type {
	case flvio.TagVideo:
		switch tag.AVCPacketType {
		case flvio.AvcSeqhdr:
			if !inst.GotVideo {
				var stream h264parser.CodecData
				if stream, err = h264parser.NewCodecDataFromAVCDecoderConfRecord(tag.Data); err != nil {
					err = fmt.Errorf("flv: h264 seqhdr invalid")
					return
				}
				inst.VideoStreamIdx = len(inst.Streams)
				inst.Streams = append(inst.Streams, stream)
				inst.GotVideo = true
			}

		case flvio.AvcNalu:
			inst.CacheTag(tag, timestamp)
		}

	case flvio.TagAudio:
		switch tag.SoundFormat {
		case flvio.SoundAAC:
			switch tag.AACPacketType {
			case flvio.AACSeqhdr:
				if !inst.GotAudio {
					var stream aacparser.CodecData
					if stream, err = aacparser.NewCodecDataFromMPEG4AudioConfigBytes(tag.Data); err != nil {
						err = fmt.Errorf("flv: aac seqhdr invalid")
						return
					}
					inst.AudioStreamIdx = len(inst.Streams)
					inst.Streams = append(inst.Streams, stream)
					inst.GotAudio = true
				}

			case flvio.AACRaw:
				inst.CacheTag(tag, timestamp)
			}

			// case flvio.SoundSpeex:
			// 	if !inst.GotAudio {
			// 		stream := codec.NewSpeexCodecData(16000, tag.ChannelLayout())
			// 		inst.AudioStreamIdx = len(inst.Streams)
			// 		inst.Streams = append(inst.Streams, stream)
			// 		inst.GotAudio = true
			// 		inst.CacheTag(tag, timestamp)
			// 	}

			// case flvio.SoundNellymoser:
			// 	if !inst.GotAudio {
			// 		stream := fake.CodecData{
			// 			CodecTypeItem:     av.NELLYMOSER,
			// 			SampleRateItem:    16000,
			// 			SampleFormatItem:  av.S16,
			// 			ChannelLayoutItem: tag.ChannelLayout(),
			// 		}
			// 		inst.AudioStreamIdx = len(inst.Streams)
			// 		inst.Streams = append(inst.Streams, stream)
			// 		inst.GotAudio = true
			// 		inst.CacheTag(tag, timestamp)
			// 	}

		}
	}

	return
}

// Probed type
func (inst *Prober) Probed() (ok bool) {
	if inst.HasAudio || inst.HasVideo {
		if inst.HasAudio == inst.GotAudio && inst.HasVideo == inst.GotVideo {
			return true
		}
	} else {
		if inst.PushedCount == MaxProbePacketCount {
			return true
		}
	}
	return
}

// TagToPacket type
func (inst *Prober) TagToPacket(tag flvio.Tag, timestamp int32) (pkt av.Packet, ok bool) {
	switch tag.Type {
	case flvio.TagVideo:
		pkt.Idx = int8(inst.VideoStreamIdx)
		switch tag.AVCPacketType {
		case flvio.AvcNalu:
			ok = true
			pkt.Data = tag.Data
			pkt.CompositionTime = flvio.TsToTime(tag.CompositionTime)
			pkt.IsKeyFrame = tag.FrameType == flvio.FrameKey
		}

	case flvio.TagAudio:
		pkt.Idx = int8(inst.AudioStreamIdx)
		switch tag.SoundFormat {
		case flvio.SoundAAC:
			switch tag.AACPacketType {
			case flvio.AACRaw:
				ok = true
				pkt.Data = tag.Data
			}

		case flvio.SoundSpeex:
			ok = true
			pkt.Data = tag.Data

		case flvio.SoundNellymoser:
			ok = true
			pkt.Data = tag.Data
		}
	}

	pkt.Time = flvio.TsToTime(timestamp)
	return
}

// Empty type
func (inst *Prober) Empty() bool {
	return len(inst.CachedPkts) == 0
}

// PopPacket type
func (inst *Prober) PopPacket() av.Packet {
	pkt := inst.CachedPkts[0]
	inst.CachedPkts = inst.CachedPkts[1:]
	return pkt
}

// CodecDataToTag type
func CodecDataToTag(stream av.CodecData) (_tag flvio.Tag, ok bool, err error) {
	switch stream.Type() {
	case av.H264:
		h264 := stream.(h264parser.CodecData)
		tag := flvio.Tag{
			Type:          flvio.TagVideo,
			AVCPacketType: flvio.AvcSeqhdr,
			CodecID:       flvio.VideoH264,
			Data:          h264.AVCDecoderConfRecordBytes(),
			FrameType:     flvio.FrameKey,
		}
		ok = true
		_tag = tag

	// case av.NELLYMOSER:
	// case av.SPEEX:

	case av.AAC:
		aac := stream.(aacparser.CodecData)
		tag := flvio.Tag{
			Type:          flvio.TagAudio,
			SoundFormat:   flvio.SoundAAC,
			SoundRate:     flvio.Sound44Khz,
			AACPacketType: flvio.AACSeqhdr,
			Data:          aac.MPEG4AudioConfigBytes(),
		}
		switch aac.SampleFormat().BytesPerSample() {
		case 1:
			tag.SoundSize = flvio.Sound8Bit
		default:
			tag.SoundSize = flvio.Sound16Bit
		}
		switch aac.ChannelLayout().Count() {
		case 1:
			tag.SoundType = flvio.SoundMono
		case 2:
			tag.SoundType = flvio.SoundStereo
		}
		ok = true
		_tag = tag

	default:
		err = fmt.Errorf("flv: unspported codecType=%v", stream.Type())
		return
	}
	return
}

// PacketToTag type
func PacketToTag(pkt av.Packet, stream av.CodecData) (tag flvio.Tag, timestamp int32) {
	switch stream.Type() {
	case av.H264:
		tag = flvio.Tag{
			Type:            flvio.TagVideo,
			AVCPacketType:   flvio.AvcNalu,
			CodecID:         flvio.VideoH264,
			Data:            pkt.Data,
			CompositionTime: flvio.TimeToTs(pkt.CompositionTime),
		}
		if pkt.IsKeyFrame {
			tag.FrameType = flvio.FrameKey
		} else {
			tag.FrameType = flvio.FrameInter
		}

	case av.AAC:
		tag = flvio.Tag{
			Type:          flvio.TagAudio,
			SoundFormat:   flvio.SoundAAC,
			SoundRate:     flvio.Sound44Khz,
			AACPacketType: flvio.AACRaw,
			Data:          pkt.Data,
		}
		astream := stream.(av.AudioCodecData)
		switch astream.SampleFormat().BytesPerSample() {
		case 1:
			tag.SoundSize = flvio.Sound8Bit
		default:
			tag.SoundSize = flvio.Sound16Bit
		}
		switch astream.ChannelLayout().Count() {
		case 1:
			tag.SoundType = flvio.SoundMono
		case 2:
			tag.SoundType = flvio.SoundStereo
		}

		// case av.SPEEX:
		// 	tag = flvio.Tag{
		// 		Type:        flvio.TagAudio,
		// 		SoundFormat: flvio.SoundSpeex,
		// 		Data:        pkt.Data,
		// 	}

		// case av.NELLYMOSER:
		// 	tag = flvio.Tag{
		// 		Type:        flvio.TagAudio,
		// 		SoundFormat: flvio.SoundNellymoser,
		// 		Data:        pkt.Data,
		// 	}
	}

	timestamp = flvio.TimeToTs(pkt.Time)
	return
}

// Muxer type
type Muxer struct {
	bufw    writeFlusher
	b       []byte
	streams []av.CodecData
}

type writeFlusher interface {
	io.Writer
	Flush() error
}

// NewMuxerWriteFlusher type
func NewMuxerWriteFlusher(w writeFlusher) *Muxer {
	return &Muxer{
		bufw: w,
		b:    make([]byte, 256),
	}
}

// NewMuxer type
func NewMuxer(w io.Writer) *Muxer {
	return NewMuxerWriteFlusher(bufio.NewWriterSize(w, pio.RecommendBufioSize))
}

// CodecTypes var
var CodecTypes = []av.CodecType{av.H264, av.AAC} //, av.SPEEX}

// WriteHeader type
func (inst *Muxer) WriteHeader(streams []av.CodecData) (err error) {
	var flags uint8
	for _, stream := range streams {
		if stream.Type().IsVideo() {
			flags |= flvio.FileHasVideo
		} else if stream.Type().IsAudio() {
			flags |= flvio.FileHasAudio
		}
	}

	n := flvio.FillFileHeader(inst.b, flags)
	if _, err = inst.bufw.Write(inst.b[:n]); err != nil {
		return
	}

	for _, stream := range streams {
		var tag flvio.Tag
		var ok bool
		if tag, ok, err = CodecDataToTag(stream); err != nil {
			return
		}
		if ok {
			if err = flvio.WriteTag(inst.bufw, tag, 0, inst.b); err != nil {
				return
			}
		}
	}

	inst.streams = streams
	return
}

// WritePacket type
func (inst *Muxer) WritePacket(pkt av.Packet) (err error) {
	stream := inst.streams[pkt.Idx]
	tag, timestamp := PacketToTag(pkt, stream)

	if err = flvio.WriteTag(inst.bufw, tag, timestamp, inst.b); err != nil {
		return
	}
	return
}

// WriteTrailer type
func (inst *Muxer) WriteTrailer() (err error) {
	if err = inst.bufw.Flush(); err != nil {
		return
	}
	return
}

// Demuxer type
type Demuxer struct {
	prober *Prober
	bufr   *bufio.Reader
	b      []byte
	stage  int
}

// NewDemuxer type
func NewDemuxer(r io.Reader) *Demuxer {
	return &Demuxer{
		bufr:   bufio.NewReaderSize(r, pio.RecommendBufioSize),
		prober: &Prober{},
		b:      make([]byte, 256),
	}
}

func (inst *Demuxer) prepare() (err error) {
	for inst.stage < 2 {
		switch inst.stage {
		case 0:
			if _, err = io.ReadFull(inst.bufr, inst.b[:flvio.FileHeaderLength]); err != nil {
				return
			}
			var flags uint8
			var skip int
			if flags, skip, err = flvio.ParseFileHeader(inst.b); err != nil {
				return
			}
			if _, err = inst.bufr.Discard(skip); err != nil {
				return
			}
			if flags&flvio.FileHasAudio != 0 {
				inst.prober.HasAudio = true
			}
			if flags&flvio.FileHasVideo != 0 {
				inst.prober.HasVideo = true
			}
			inst.stage++

		case 1:
			for !inst.prober.Probed() {
				var tag flvio.Tag
				var timestamp int32
				if tag, timestamp, err = flvio.ReadTag(inst.bufr, inst.b); err != nil {
					return
				}
				if err = inst.prober.PushTag(tag, timestamp); err != nil {
					return
				}
			}
			inst.stage++
		}
	}
	return
}

// Streams type
func (inst *Demuxer) Streams() (streams []av.CodecData, err error) {
	if err = inst.prepare(); err != nil {
		return
	}
	streams = inst.prober.Streams
	return
}

// ReadPacket type
func (inst *Demuxer) ReadPacket() (pkt av.Packet, err error) {
	if err = inst.prepare(); err != nil {
		return
	}

	if !inst.prober.Empty() {
		pkt = inst.prober.PopPacket()
		return
	}

	for {
		var tag flvio.Tag
		var timestamp int32
		if tag, timestamp, err = flvio.ReadTag(inst.bufr, inst.b); err != nil {
			return
		}

		var ok bool
		if pkt, ok = inst.prober.TagToPacket(tag, timestamp); ok {
			return
		}
	}

	// return
}

// Handler type
func Handler(h *avutil.RegisterHandler) {
	h.Probe = func(b []byte) bool {
		return b[0] == 'F' && b[1] == 'L' && b[2] == 'V'
	}

	h.Ext = ".flv"

	h.ReaderDemuxer = func(r io.Reader) av.Demuxer {
		return NewDemuxer(r)
	}

	h.WriterMuxer = func(w io.Writer) av.Muxer {
		return NewMuxer(w)
	}

	h.CodecTypes = CodecTypes
}
