package aacparser

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/Youngju-Heo/gomedia/core/media/av"
	"github.com/Youngju-Heo/gomedia/core/media/utils/bits"
)

// copied from libavcodec/mpeg4audio.h
const (
	AotAACMain       = 1 + iota  ///< Y                       Main
	AotAACLc                     ///< Y                       Low Complexity
	AotAACSsr                    ///< N (code in SoC repo)    Scalable Sample Rate
	AotAACLtp                    ///< Y                       Long Term Prediction
	AotSBR                       ///< Y                       Spectral Band Replication
	AotAACScalable               ///< N                       Scalable
	AotTWINVQ                    ///< N                       Twin Vector Quantizer
	AotCELP                      ///< N                       Code Excited Linear Prediction
	AotHVXC                      ///< N                       Harmonic Vector eXcitation Coding
	AotTTSI          = 12 + iota ///< N                       Text-To-Speech Interface
	AotMAINSYNTH                 ///< N                       Main Synthesis
	AotWAVESYNTH                 ///< N                       Wavetable Synthesis
	AotMIDI                      ///< N                       General MIDI
	AotSAFX                      ///< N                       Algorithmic Synthesis and Audio Effects
	AotErAACLc                   ///< N                       Error Resilient Low Complexity
	AotErAACLtp      = 19 + iota ///< N                       Error Resilient Long Term Prediction
	AotErAACScalable             ///< N                       Error Resilient Scalable
	AotErTWINVQ                  ///< N                       Error Resilient Twin Vector Quantizer
	AotErBSAC                    ///< N                       Error Resilient Bit-Sliced Arithmetic Coding
	AotErAACLd                   ///< N                       Error Resilient Low Delay
	AotErCELP                    ///< N                       Error Resilient Code Excited Linear Prediction
	AotErHVXC                    ///< N                       Error Resilient Harmonic Vector eXcitation Coding
	AotErHILN                    ///< N                       Error Resilient Harmonic and Individual Lines plus Noise
	AotErPARAM                   ///< N                       Error Resilient Parametric
	AotSSC                       ///< N                       SinuSoidal Coding
	AotPS                        ///< N                       Parametric Stereo
	AotSURROUND                  ///< N                       MPEG Surround
	AotESCAPE                    ///< Y                       Escape Value
	AotL1                        ///< Y                       Layer 1
	AotL2                        ///< Y                       Layer 2
	AotL3                        ///< Y                       Layer 3
	AotDST                       ///< N                       Direct Stream Transfer
	AotALS                       ///< Y                       Audio LosslesS
	AotSLS                       ///< N                       Scalable LosslesS
	AotSlsNonCore                ///< N                       Scalable LosslesS (non core)
	AotErAACEld                  ///< N                       Error Resilient Enhanced Low Delay
	AotSmrSimple                 ///< N                       Symbolic Music Representation Simple
	AotSmrMain                   ///< N                       Symbolic Music Representation Main
	AotUsacNosbr                 ///< N                       Unified Speech and Audio Coding (no SBR)
	AotSAOC                      ///< N                       Spatial Audio Object Coding
	AotLdSurround                ///< N                       Low Delay MPEG Surround
	AotUsac                      ///< N                       Unified Speech and Audio Coding
)

// MPEG4AudioConfig struct
type MPEG4AudioConfig struct {
	SampleRate      int
	ChannelLayout   av.ChannelLayout
	ObjectType      uint
	SampleRateIndex uint
	ChannelConfig   uint
}

var sampleRateTable = []int{
	96000, 88200, 64000, 48000, 44100, 32000,
	24000, 22050, 16000, 12000, 11025, 8000, 7350,
}

/*
These are the channel configurations:
0: Defined in AOT Specifc Config
1: 1 channel: front-center
2: 2 channels: front-left, front-right
3: 3 channels: front-center, front-left, front-right
4: 4 channels: front-center, front-left, front-right, back-center
5: 5 channels: front-center, front-left, front-right, back-left, back-right
6: 6 channels: front-center, front-left, front-right, back-left, back-right, LFE-channel
7: 8 channels: front-center, front-left, front-right, side-left, side-right, back-left, back-right, LFE-channel
8-15: Reserved
*/
var chanConfigTable = []av.ChannelLayout{
	0,
	av.ChFrontCenter,
	av.ChFrontLeft | av.ChFrontRight,
	av.ChFrontCenter | av.ChFrontLeft | av.ChFrontRight,
	av.ChFrontCenter | av.ChFrontLeft | av.ChFrontRight | av.ChBackCenter,
	av.ChFrontCenter | av.ChFrontLeft | av.ChFrontRight | av.ChBackLeft | av.ChBackRight,
	av.ChFrontCenter | av.ChFrontLeft | av.ChFrontRight | av.ChBackLeft | av.ChBackRight | av.ChLowFreq,
	av.ChFrontCenter | av.ChFrontLeft | av.ChFrontRight | av.ChSideLeft | av.ChSideRight | av.ChBackLeft | av.ChBackRight | av.ChLowFreq,
}

// ParseADTSHeader func
func ParseADTSHeader(frame []byte) (config MPEG4AudioConfig, hdrlen int, framelen int, samples int, err error) {
	if frame[0] != 0xff || frame[1]&0xf6 != 0xf0 {
		err = fmt.Errorf("aacparser: not adts header")
		return
	}
	config.ObjectType = uint(frame[2]>>6) + 1
	config.SampleRateIndex = uint(frame[2] >> 2 & 0xf)
	config.ChannelConfig = uint(frame[2]<<2&0x4 | frame[3]>>6&0x3)
	if config.ChannelConfig == uint(0) {
		err = fmt.Errorf("aacparser: adts channel count invalid")
		return
	}
	(&config).Complete()
	framelen = int(frame[3]&0x3)<<11 | int(frame[4])<<3 | int(frame[5]>>5)
	samples = (int(frame[6]&0x3) + 1) * 1024
	hdrlen = 7
	if frame[1]&0x1 == 0 {
		hdrlen = 9
	}
	if framelen < hdrlen {
		err = fmt.Errorf("aacparser: adts framelen < hdrlen")
		return
	}
	return
}

// ADTSHeaderLength const
const ADTSHeaderLength = 7

// FillADTSHeader func
func FillADTSHeader(header []byte, config MPEG4AudioConfig, samples int, payloadLength int) {
	payloadLength += 7
	//AAAAAAAA AAAABCCD EEFFFFGH HHIJKLMM MMMMMMMM MMMOOOOO OOOOOOPP (QQQQQQQQ QQQQQQQQ)
	header[0] = 0xff
	header[1] = 0xf1
	header[2] = 0x50
	header[3] = 0x80
	header[4] = 0x43
	header[5] = 0xff
	header[6] = 0xcd
	//config.ObjectType = uint(frames[2]>>6)+1
	//config.SampleRateIndex = uint(frames[2]>>2&0xf)
	//config.ChannelConfig = uint(frames[2]<<2&0x4|frames[3]>>6&0x3)
	header[2] = (byte(config.ObjectType-1)&0x3)<<6 | (byte(config.SampleRateIndex)&0xf)<<2 | byte(config.ChannelConfig>>2)&0x1
	header[3] = header[3]&0x3f | byte(config.ChannelConfig&0x3)<<6
	header[3] = header[3]&0xfc | byte(payloadLength>>11)&0x3
	header[4] = byte(payloadLength >> 3)
	header[5] = header[5]&0x1f | (byte(payloadLength)&0x7)<<5
	header[6] = header[6]&0xfc | byte(samples/1024-1)
	return
}

func readObjectType(r *bits.Reader) (objectType uint, err error) {
	if objectType, err = r.ReadBits(5); err != nil {
		return
	}
	if objectType == AotESCAPE {
		var i uint
		if i, err = r.ReadBits(6); err != nil {
			return
		}
		objectType = 32 + i
	}
	return
}

func writeObjectType(w *bits.Writer, objectType uint) (err error) {
	if objectType >= 32 {
		if err = w.WriteBits(AotESCAPE, 5); err != nil {
			return
		}
		if err = w.WriteBits(objectType-32, 6); err != nil {
			return
		}
	} else {
		if err = w.WriteBits(objectType, 5); err != nil {
			return
		}
	}
	return
}

func readSampleRateIndex(r *bits.Reader) (index uint, err error) {
	if index, err = r.ReadBits(4); err != nil {
		return
	}
	if index == 0xf {
		if index, err = r.ReadBits(24); err != nil {
			return
		}
	}
	return
}

func writeSampleRateIndex(w *bits.Writer, index uint) (err error) {
	if index >= 0xf {
		if err = w.WriteBits(0xf, 4); err != nil {
			return
		}
		if err = w.WriteBits(index, 24); err != nil {
			return
		}
	} else {
		if err = w.WriteBits(index, 4); err != nil {
			return
		}
	}
	return
}

// IsValid func
func (inst MPEG4AudioConfig) IsValid() bool {
	return inst.ObjectType > 0
}

// Complete func
func (inst *MPEG4AudioConfig) Complete() {
	if int(inst.SampleRateIndex) < len(sampleRateTable) {
		inst.SampleRate = sampleRateTable[inst.SampleRateIndex]
	}
	if int(inst.ChannelConfig) < len(chanConfigTable) {
		inst.ChannelLayout = chanConfigTable[inst.ChannelConfig]
	}
	return
}

// ParseMPEG4AudioConfigBytes func
func ParseMPEG4AudioConfigBytes(data []byte) (config MPEG4AudioConfig, err error) {
	// copied from libavcodec/mpeg4audio.c avpriv_mpeg4audio_get_config()
	r := bytes.NewReader(data)
	br := &bits.Reader{R: r}
	if config.ObjectType, err = readObjectType(br); err != nil {
		return
	}
	if config.SampleRateIndex, err = readSampleRateIndex(br); err != nil {
		return
	}
	if config.ChannelConfig, err = br.ReadBits(4); err != nil {
		return
	}
	(&config).Complete()
	return
}

// WriteMPEG4AudioConfig func
func WriteMPEG4AudioConfig(w io.Writer, config MPEG4AudioConfig) (err error) {
	bw := &bits.Writer{W: w}
	if err = writeObjectType(bw, config.ObjectType); err != nil {
		return
	}

	if config.SampleRateIndex == 0 {
		for i, rate := range sampleRateTable {
			if rate == config.SampleRate {
				config.SampleRateIndex = uint(i)
			}
		}
	}
	if err = writeSampleRateIndex(bw, config.SampleRateIndex); err != nil {
		return
	}

	if config.ChannelConfig == 0 {
		for i, layout := range chanConfigTable {
			if layout == config.ChannelLayout {
				config.ChannelConfig = uint(i)
			}
		}
	}
	if err = bw.WriteBits(config.ChannelConfig, 4); err != nil {
		return
	}

	if err = bw.FlushBits(); err != nil {
		return
	}
	return
}

// CodecData struct
type CodecData struct {
	ConfigBytes []byte
	Config      MPEG4AudioConfig
}

// Type func
func (inst CodecData) Type() av.CodecType {
	return av.AAC
}

// MPEG4AudioConfigBytes func
func (inst CodecData) MPEG4AudioConfigBytes() []byte {
	return inst.ConfigBytes
}

// ChannelLayout func
func (inst CodecData) ChannelLayout() av.ChannelLayout {
	return inst.Config.ChannelLayout
}

// SampleRate func
func (inst CodecData) SampleRate() int {
	return inst.Config.SampleRate
}

// SampleFormat func
func (inst CodecData) SampleFormat() av.SampleFormat {
	return av.FLTP
}

// PacketDuration func
func (inst CodecData) PacketDuration(data []byte) (dur time.Duration, err error) {
	dur = time.Duration(1024) * time.Second / time.Duration(inst.Config.SampleRate)
	return
}

// NewCodecDataFromMPEG4AudioConfig func
func NewCodecDataFromMPEG4AudioConfig(config MPEG4AudioConfig) (inst CodecData, err error) {
	b := &bytes.Buffer{}
	WriteMPEG4AudioConfig(b, config)
	return NewCodecDataFromMPEG4AudioConfigBytes(b.Bytes())
}

// NewCodecDataFromMPEG4AudioConfigBytes func
func NewCodecDataFromMPEG4AudioConfigBytes(config []byte) (inst CodecData, err error) {
	inst.ConfigBytes = config
	if inst.Config, err = ParseMPEG4AudioConfigBytes(config); err != nil {
		err = fmt.Errorf("aacparser: parse MPEG4AudioConfig failed(%s)", err)
		return
	}
	return
}
