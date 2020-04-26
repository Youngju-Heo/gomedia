// Package av defines basic interfaces and data structures of container demux/mux and audio encode/decode.
package av

import (
	"fmt"
	"time"
)

// SampleFormat Audio sample format.
type SampleFormat uint8

const (
	// U8 data
	U8 = SampleFormat(iota + 1) // U8 8-bit unsigned integer
	// S16 data
	S16 // signed 16-bit integer
	// S32 data
	S32 // signed 32-bit integer
	// FLT data
	FLT // 32-bit float
	// DBL data
	DBL // 64-bit float
	// U8P data
	U8P // 8-bit unsigned integer in planar
	// S16P data
	S16P // signed 16-bit integer in planar
	// S32P data
	S32P // signed 32-bit integer in planar
	// FLTP data
	FLTP // 32-bit float in planar
	// DBLP data
	DBLP // 64-bit float in planar
	// U32 data
	U32 // unsigned 32-bit integer
)

// BytesPerSample BytesPerSample
func (format SampleFormat) BytesPerSample() int {
	switch format {
	case U8, U8P:
		return 1
	case S16, S16P:
		return 2
	case FLT, FLTP, S32, S32P, U32:
		return 4
	case DBL, DBLP:
		return 8
	default:
		return 0
	}
}

func (format SampleFormat) String() string {
	switch format {
	case U8:
		return "U8"
	case S16:
		return "S16"
	case S32:
		return "S32"
	case FLT:
		return "FLT"
	case DBL:
		return "DBL"
	case U8P:
		return "U8P"
	case S16P:
		return "S16P"
	case FLTP:
		return "FLTP"
	case DBLP:
		return "DBLP"
	case U32:
		return "U32"
	default:
		return "?"
	}
}

// IsPlanar Check if this sample format is in planar.
func (format SampleFormat) IsPlanar() bool {
	switch format {
	case S16P, S32P, FLTP, DBLP:
		return true
	default:
		return false
	}
}

// ChannelLayout Audio channel layout.
type ChannelLayout uint16

// String channel layout string
func (layout ChannelLayout) String() string {
	return fmt.Sprintf("%dch", layout.Count())
}

// define Audio channel layout
const (
	ChFrontCenter = ChannelLayout(1 << iota)
	ChFrontLeft
	ChFrontRight
	ChBackCenter
	ChBackLeft
	ChBackRight
	ChSideLeft
	ChSideRight
	ChLowFreq
	ChNr

	ChMono     = ChannelLayout(ChFrontCenter)
	ChStereo   = ChannelLayout(ChFrontLeft | ChFrontRight)
	Ch21       = ChannelLayout(ChStereo | ChBackCenter)
	Ch2Point1  = ChannelLayout(ChStereo | ChLowFreq)
	ChSurround = ChannelLayout(ChStereo | ChFrontCenter)
	Ch3Point1  = ChannelLayout(ChSurround | ChLowFreq)
	// TODO: add all channel_layout in ffmpeg
)

// Count Count
func (layout ChannelLayout) Count() (n int) {
	for layout != 0 {
		n++
		layout = (layout - 1) & layout
	}
	return n
}

// CodecType Video/Audio codec type. can be H264/AAC/SPEEX/...
type CodecType uint32

// Define Codec type
var (
	TEXT = CodecType(0)
	// define video
	H264 = MakeVideoCodecType(avCodecTypeMagic + 1)
	JPEG = MakeVideoCodecType(avCodecTypeMagic + 2)
	HEVC = MakeVideoCodecType(avCodecTypeMagic + 3)

	// define audio
	AAC  = MakeAudioCodecType(avCodecTypeMagic + 1)
	PCMU = MakeAudioCodecType(avCodecTypeMagic + 2)
	PCMA = MakeAudioCodecType(avCodecTypeMagic + 3)
	MP3  = MakeAudioCodecType(avCodecTypeMagic + 4)
)

const codecTypeAudioBit = 0x1
const codecTypeOtherBits = 1

// String CodecType to string name
func (ctype CodecType) String() string {
	switch ctype {
	// from video
	case H264:
		return "H264"
	case JPEG:
		return "JPEG"
	case HEVC:
		return "HEVC"

	// from audio
	case AAC:
		return "AAC"
	case PCMU:
		return "PCMU"
	case PCMA:
		return "PCMA"
	case MP3:
		return "MP3"
	}
	return ""
}

// CodecName CodecType to string name
func (ctype CodecType) CodecName() string {
	switch ctype {
	// from video
	case H264:
		return "h264"
	case JPEG:
		return "jpeg"
	case HEVC:
		return "hevc"

	// from audio
	case AAC:
		return "aac"
	case PCMU:
		return "pcm_mulaw"
	case PCMA:
		return "pcm_alaw"
	case MP3:
		return "libmp3lame"
	}
	return ""
}

// IsAudio IsAudio
func (ctype CodecType) IsAudio() bool {
	return ctype&codecTypeAudioBit != 0
}

// IsVideo IsVideo
func (ctype CodecType) IsVideo() bool {
	return ctype&codecTypeAudioBit == 0
}

// MakeAudioCodecType Make a new audio codec type.
func MakeAudioCodecType(base uint32) (c CodecType) {
	c = CodecType(base)<<codecTypeOtherBits | CodecType(codecTypeAudioBit)
	return
}

// MakeVideoCodecType Make a new video codec type.
func MakeVideoCodecType(base uint32) (c CodecType) {
	c = CodecType(base) << codecTypeOtherBits
	return
}

const avCodecTypeMagic = 0x10000

// CodecData is some important bytes for initializing audio/video decoder,
// can be converted to VideoCodecData or AudioCodecData using:
//
//     codecdata.(AudioCodecData) or codecdata.(VideoCodecData)
//
// for H264, CodecData is AVCDecoderConfigure bytes, includes SPS/PPS.
type CodecData interface {
	Type() CodecType // Video/Audio codec type
}

// VideoCodecData VideoCodecData
type VideoCodecData interface {
	CodecData
	Width() int  // Video width
	Height() int // Video height
}

// H264VideoCodecData for h264 info
type H264VideoCodecData interface {
	VideoCodecData
	SPS() []byte
	PPS() []byte
}

// H265VideoCodecData for h264 info
type H265VideoCodecData interface {
	VideoCodecData
	VPS() []byte
	SPS() []byte
	PPS() []byte
}

// AudioCodecData AudioCodecData
type AudioCodecData interface {
	CodecData
	SampleFormat() SampleFormat                   // audio sample format
	SampleRate() int                              // audio sample rate
	ChannelLayout() ChannelLayout                 // audio channel layout
	PacketDuration([]byte) (time.Duration, error) // get audio compressed packet duration
}

// MPEG4AudioCodecData audio codec data for aac
type MPEG4AudioCodecData interface {
	AudioCodecData
	MPEG4AudioConfigBytes() []byte
}

// PacketWriter PacketWriter
type PacketWriter interface {
	WritePacket(Packet) error
}

// PacketReader PacketReader
type PacketReader interface {
	ReadPacket() (Packet, error)
}

// Muxer describes the steps of writing compressed audio/video packets into container formats like MP4/FLV/MPEG-TS.
//
// Container formats, rtmp.Conn, and transcode.Muxer implements Muxer interface.
type Muxer interface {
	WriteHeader([]CodecData) error // write the file header
	PacketWriter                   // write compressed audio/video packets
	WriteTrailer() error           // finish writing file, this func can be called only once
}

// MuxCloser Muxer with Close() method
type MuxCloser interface {
	Muxer
	Close() error
}

// Demuxer can read compressed audio/video packets from container formats like MP4/FLV/MPEG-TS.
type Demuxer interface {
	PacketReader                   // read compressed audio/video packets
	Streams() ([]CodecData, error) // reads the file header, contains video/audio meta infomations
}

// DemuxCloser Demuxer with Close() method
type DemuxCloser interface {
	Demuxer
	Close() error
}

// Packet stores compressed audio/video data.
type Packet struct {
	IsKeyFrame      bool          // video packet is key frame
	Idx             int8          // stream index in container format
	CompositionTime time.Duration // packet presentation time minus decode time for H264 B-Frame
	Time            time.Duration // packet decode time
	Data            []byte        // packet data
}

// AudioFrame Raw audio frame.
type AudioFrame struct {
	SampleFormat  SampleFormat  // audio sample format, e.g: S16,FLTP,...
	ChannelLayout ChannelLayout // audio channel layout, e.g: ChMono,ChStereo,...
	SampleCount   int           // sample count in this frame
	SampleRate    int           // sample rate
	Data          [][]byte      // data array for planar format len(Data) > 1
}

// Duration Duration
func (frame AudioFrame) Duration() time.Duration {
	return time.Second * time.Duration(frame.SampleCount) / time.Duration(frame.SampleRate)
}

// HasSameFormat Check this audio frame has same format as other audio frame.
func (frame AudioFrame) HasSameFormat(other AudioFrame) bool {
	if frame.SampleRate != other.SampleRate {
		return false
	}
	if frame.ChannelLayout != other.ChannelLayout {
		return false
	}
	if frame.SampleFormat != other.SampleFormat {
		return false
	}
	return true
}

// Slice Split sample audio sample from this frame.
func (frame AudioFrame) Slice(start int, end int) (out AudioFrame) {
	if start > end {
		panic(fmt.Sprintf("av: AudioFrame split failed start=%d end=%d invalid", start, end))
	}
	out = frame
	out.Data = append([][]byte(nil), out.Data...)
	out.SampleCount = end - start
	size := frame.SampleFormat.BytesPerSample()
	for i := range out.Data {
		out.Data[i] = out.Data[i][start*size : end*size]
	}
	return
}

// Concat two audio frames.
func (frame AudioFrame) Concat(in AudioFrame) AudioFrame {
	var out = frame
	out.Data = append([][]byte(nil), out.Data...)
	out.SampleCount += in.SampleCount
	for i := range out.Data {
		out.Data[i] = append(out.Data[i], in.Data[i]...)
	}
	return out
}

// AudioEncoder can encode raw audio frame into compressed audio packets.
// cgo/ffmpeg inplements AudioEncoder, using ffmpeg.NewAudioEncoder to create it.
type AudioEncoder interface {
	CodecData() (AudioCodecData, error)   // encoder's codec data can put into container
	Encode(AudioFrame) ([][]byte, error)  // encode raw audio frame into compressed pakcet(s)
	Close()                               // close encoder, free cgo contexts
	SetSampleRate(int) error              // set encoder sample rate
	SetChannelLayout(ChannelLayout) error // set encoder channel layout
	SetSampleFormat(SampleFormat) error   // set encoder sample format
	SetBitrate(int) error                 // set encoder bitrate
	SetOption(string, interface{}) error  // encoder setopt, in ffmpeg is av_opt_set_dict()
	GetOption(string, interface{}) error  // encoder getopt
}

// AudioDecoder can decode compressed audio packets into raw audio frame.
// use ffmpeg.NewAudioDecoderParam to create it.
type AudioDecoder interface {
	Decode([]byte) (bool, AudioFrame, error) // decode one compressed audio packet
	Close()                                  // close decode, free cgo contexts
}

// AudioResampler can convert raw audio frames in different sample rate/format/channel layout.
type AudioResampler interface {
	Resample(AudioFrame) (AudioFrame, error) // convert raw audio frames
}
