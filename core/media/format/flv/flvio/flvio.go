package flvio

import (
	"fmt"
	"io"
	"time"

	"github.com/Youngju-Heo/gomedia/core/media/av"
	"github.com/Youngju-Heo/gomedia/core/media/utils/bits/pio"
)

// TsToTime func
func TsToTime(ts int32) time.Duration {
	return time.Millisecond * time.Duration(ts)
}

// TimeToTs func
func TimeToTs(tm time.Duration) int32 {
	return int32(tm / time.Millisecond)
}

// MaxTagSubHeaderLength const
const MaxTagSubHeaderLength = 16

const (
	// TagAudio const
	TagAudio = 8
	//TagVideo const
	TagVideo = 9
	// TagScriptdata const
	TagScriptdata = 18
)

const (
	// SoundMP3 const
	SoundMP3 = 2
	// SoundNellymoser16KhzMono const
	SoundNellymoser16KhzMono = 4
	// SoundNellymoser8KhzMono const
	SoundNellymoser8KhzMono = 5
	// SoundNellymoser const
	SoundNellymoser = 6
	// SoundAlaw const
	SoundAlaw = 7
	// SoundMulaw const
	SoundMulaw = 8
	// SoundAAC const
	SoundAAC = 10
	// SoundSpeex const
	SoundSpeex = 11

	// Sound5Dot5Khz const
	Sound5Dot5Khz = 0
	// Sound11Khz const
	Sound11Khz = 1
	// Sound22Khz const
	Sound22Khz = 2
	// Sound44Khz const
	Sound44Khz = 3

	// Sound8Bit const
	Sound8Bit = 0
	// Sound16Bit const
	Sound16Bit = 1

	// SoundMono const
	SoundMono = 0
	// SoundStereo const
	SoundStereo = 1

	// AACSeqhdr const
	AACSeqhdr = 0
	// AACRaw const
	AACRaw = 1
)

const (

	// AvcSeqhdr const
	AvcSeqhdr = 0
	// AvcNalu const
	AvcNalu = 1
	//AvcEos const
	AvcEos = 2

	// FrameKey const
	FrameKey = 1
	// FrameInter const
	FrameInter = 2

	// VideoH264 const
	VideoH264 = 7
)

// Tag struct
type Tag struct {
	Type uint8

	/*
		SoundFormat: UB[4]
		0 = Linear PCM, platform endian
		1 = ADPCM
		2 = MP3
		3 = Linear PCM, little endian
		4 = Nellymoser 16-kHz mono
		5 = Nellymoser 8-kHz mono
		6 = Nellymoser
		7 = G.711 A-law logarithmic PCM
		8 = G.711 mu-law logarithmic PCM
		9 = reserved
		10 = AAC
		11 = Speex
		14 = MP3 8-Khz
		15 = Device-specific sound
		Formats 7, 8, 14, and 15 are reserved for internal use
		AAC is supported in Flash Player 9,0,115,0 and higher.
		Speex is supported in Flash Player 10 and higher.
	*/
	SoundFormat uint8

	/*
		SoundRate: UB[2]
		Sampling rate
		0 = 5.5-kHz For AAC: always 3
		1 = 11-kHz
		2 = 22-kHz
		3 = 44-kHz
	*/
	SoundRate uint8

	/*
		SoundSize: UB[1]
		0 = snd8Bit
		1 = snd16Bit
		Size of each sample.
		This parameter only pertains to uncompressed formats.
		Compressed formats always decode to 16 bits internally
	*/
	SoundSize uint8

	/*
		SoundType: UB[1]
		0 = sndMono
		1 = sndStereo
		Mono or stereo sound For Nellymoser: always 0
		For AAC: always 1
	*/
	SoundType uint8

	/*
		0: AAC sequence header
		1: AAC raw
	*/
	AACPacketType uint8

	/*
		1: keyframe (for AVC, a seekable frame)
		2: inter frame (for AVC, a non- seekable frame)
		3: disposable inter frame (H.263 only)
		4: generated keyframe (reserved for server use only)
		5: video info/command frame
	*/
	FrameType uint8

	/*
		1: JPEG (currently unused)
		2: Sorenson H.263
		3: Screen video
		4: On2 VP6
		5: On2 VP6 with alpha channel
		6: Screen video version 2
		7: AVC
	*/
	CodecID uint8

	/*
		0: AVC sequence header
		1: AVC NALU
		2: AVC end of sequence (lower level NALU sequence ender is not required or supported)
	*/
	AVCPacketType uint8

	CompositionTime int32

	Data []byte
}

// ChannelLayout func
func (inst Tag) ChannelLayout() av.ChannelLayout {
	if inst.SoundType == SoundMono {
		return av.ChMono
	}
	return av.ChStereo
}

func (inst *Tag) audioParseHeader(b []byte) (n int, err error) {
	if len(b) < n+1 {
		err = fmt.Errorf("audiodata: parse invalid")
		return
	}

	flags := b[n]
	n++
	inst.SoundFormat = flags >> 4
	inst.SoundRate = (flags >> 2) & 0x3
	inst.SoundSize = (flags >> 1) & 0x1
	inst.SoundType = flags & 0x1

	switch inst.SoundFormat {
	case SoundAAC:
		if len(b) < n+1 {
			err = fmt.Errorf("audiodata: parse invalid")
			return
		}
		inst.AACPacketType = b[n]
		n++
	}

	return
}

func (inst Tag) audioFillHeader(b []byte) (n int) {
	var flags uint8
	flags |= inst.SoundFormat << 4
	flags |= inst.SoundRate << 2
	flags |= inst.SoundSize << 1
	flags |= inst.SoundType
	b[n] = flags
	n++

	switch inst.SoundFormat {
	case SoundAAC:
		b[n] = inst.AACPacketType
		n++
	}

	return
}

func (inst *Tag) videoParseHeader(b []byte) (n int, err error) {
	if len(b) < n+1 {
		err = fmt.Errorf("videodata: parse invalid")
		return
	}
	flags := b[n]
	inst.FrameType = flags >> 4
	inst.CodecID = flags & 0xf
	n++

	if inst.FrameType == FrameInter || inst.FrameType == FrameKey {
		if len(b) < n+4 {
			err = fmt.Errorf("videodata: parse invalid")
			return
		}
		inst.AVCPacketType = b[n]
		n++

		inst.CompositionTime = pio.I24BE(b[n:])
		n += 3
	}

	return
}

func (inst Tag) videoFillHeader(b []byte) (n int) {
	flags := inst.FrameType<<4 | inst.CodecID
	b[n] = flags
	n++
	b[n] = inst.AVCPacketType
	n++
	pio.PutI24BE(b[n:], inst.CompositionTime)
	n += 3
	return
}

// FillHeader func
func (inst Tag) FillHeader(b []byte) (n int) {
	switch inst.Type {
	case TagAudio:
		return inst.audioFillHeader(b)

	case TagVideo:
		return inst.videoFillHeader(b)
	}

	return
}

// ParseHeader func
func (inst *Tag) ParseHeader(b []byte) (n int, err error) {
	switch inst.Type {
	case TagAudio:
		return inst.audioParseHeader(b)

	case TagVideo:
		return inst.videoParseHeader(b)
	}

	return
}

const (
	// TypeFlagsReserved UB[5]
	// TypeFlagsAudio    UB[1] Audio tags are present
	// TypeFlagsReserved UB[1] Must be 0
	// TypeFlagsVideo    UB[1] Video tags are present

	// FileHasAudio const
	FileHasAudio = 0x4
	// FileHasVideo const
	FileHasVideo = 0x1
)

// TagHeaderLength const
const TagHeaderLength = 11

// TagTrailerLength const
const TagTrailerLength = 4

// ParseTagHeader func
func ParseTagHeader(b []byte) (tag Tag, ts int32, datalen int, err error) {
	tagtype := b[0]

	switch tagtype {
	case TagAudio, TagVideo, TagScriptdata:
		tag = Tag{Type: tagtype}

	default:
		err = fmt.Errorf("flvio: ReadTag tagtype=%d invalid", tagtype)
		return
	}

	datalen = int(pio.U24BE(b[1:4]))

	var tslo uint32
	var tshi uint8
	tslo = pio.U24BE(b[4:7])
	tshi = b[7]
	ts = int32(tslo | uint32(tshi)<<24)

	return
}

// ReadTag func
func ReadTag(r io.Reader, b []byte) (tag Tag, ts int32, err error) {
	if _, err = io.ReadFull(r, b[:TagHeaderLength]); err != nil {
		return
	}
	var datalen int
	if tag, ts, datalen, err = ParseTagHeader(b); err != nil {
		return
	}

	data := make([]byte, datalen)
	if _, err = io.ReadFull(r, data); err != nil {
		return
	}

	var n int
	if n, err = (&tag).ParseHeader(data); err != nil {
		return
	}
	tag.Data = data[n:]

	if _, err = io.ReadFull(r, b[:4]); err != nil {
		return
	}
	return
}

// FillTagHeader func
func FillTagHeader(b []byte, tagtype uint8, datalen int, ts int32) (n int) {
	b[n] = tagtype
	n++
	pio.PutU24BE(b[n:], uint32(datalen))
	n += 3
	pio.PutU24BE(b[n:], uint32(ts&0xffffff))
	n += 3
	b[n] = uint8(ts >> 24)
	n++
	pio.PutI24BE(b[n:], 0)
	n += 3
	return
}

// FillTagTrailer func
func FillTagTrailer(b []byte, datalen int) (n int) {
	pio.PutU32BE(b[n:], uint32(datalen+TagHeaderLength))
	n += 4
	return
}

// WriteTag func
func WriteTag(w io.Writer, tag Tag, ts int32, b []byte) (err error) {
	data := tag.Data

	n := tag.FillHeader(b[TagHeaderLength:])
	datalen := len(data) + n

	n += FillTagHeader(b, tag.Type, datalen, ts)

	if _, err = w.Write(b[:n]); err != nil {
		return
	}

	if _, err = w.Write(data); err != nil {
		return
	}

	n = FillTagTrailer(b, datalen)
	if _, err = w.Write(b[:n]); err != nil {
		return
	}

	return
}

// FileHeaderLength const
const FileHeaderLength = 9

// FillFileHeader func
func FillFileHeader(b []byte, flags uint8) (n int) {
	// 'FLV', version 1
	pio.PutU32BE(b[n:], 0x464c5601)
	n += 4

	b[n] = flags
	n++

	// DataOffset: UI32 Offset in bytes from start of file to start of body (that is, size of header)
	// The DataOffset field usually has a value of 9 for FLV version 1.
	pio.PutU32BE(b[n:], 9)
	n += 4

	// PreviousTagSize0: UI32 Always 0
	pio.PutU32BE(b[n:], 0)
	n += 4

	return
}

// ParseFileHeader func
func ParseFileHeader(b []byte) (flags uint8, skip int, err error) {
	flv := pio.U24BE(b[0:3])
	if flv != 0x464c56 { // 'FLV'
		err = fmt.Errorf("flvio: file header cc3 invalid")
		return
	}

	flags = b[4]

	skip = int(pio.U32BE(b[5:9])) - 9 + 4
	if skip < 0 {
		err = fmt.Errorf("flvio: file header datasize invalid")
		return
	}

	return
}
