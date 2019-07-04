package aac

import (
	"bufio"
	"fmt"
	"io"
	"time"

	"github.com/Youngju-Heo/gomedia/core/media/av"
	"github.com/Youngju-Heo/gomedia/core/media/av/avutil"
	"github.com/Youngju-Heo/gomedia/core/media/codec/aacparser"
)

// Muxer struct
type Muxer struct {
	w       io.Writer
	config  aacparser.MPEG4AudioConfig
	adtshdr []byte
}

// NewMuxer func
func NewMuxer(w io.Writer) *Muxer {
	return &Muxer{
		adtshdr: make([]byte, aacparser.ADTSHeaderLength),
		w:       w,
	}
}

// WriteHeader func
func (inst *Muxer) WriteHeader(streams []av.CodecData) (err error) {
	if len(streams) > 1 || streams[0].Type() != av.AAC {
		err = fmt.Errorf("aac: must be only one aac stream")
		return
	}
	inst.config = streams[0].(aacparser.CodecData).Config
	if inst.config.ObjectType > aacparser.AotAACLtp {
		err = fmt.Errorf("aac: AOT %d is not allowed in ADTS", inst.config.ObjectType)
	}
	return
}

// WritePacket func
func (inst *Muxer) WritePacket(pkt av.Packet) (err error) {
	aacparser.FillADTSHeader(inst.adtshdr, inst.config, 1024, len(pkt.Data))
	if _, err = inst.w.Write(inst.adtshdr); err != nil {
		return
	}
	if _, err = inst.w.Write(pkt.Data); err != nil {
		return
	}
	return
}

// WriteTrailer func
func (inst *Muxer) WriteTrailer() (err error) {
	return
}

// Demuxer struct
type Demuxer struct {
	r         *bufio.Reader
	config    aacparser.MPEG4AudioConfig
	codecdata av.CodecData
	ts        time.Duration
}

// NewDemuxer func
func NewDemuxer(r io.Reader) *Demuxer {
	return &Demuxer{
		r: bufio.NewReader(r),
	}
}

// Streams func
func (inst *Demuxer) Streams() (streams []av.CodecData, err error) {
	if inst.codecdata == nil {
		var adtshdr []byte
		var config aacparser.MPEG4AudioConfig
		if adtshdr, err = inst.r.Peek(9); err != nil {
			return
		}
		if config, _, _, _, err = aacparser.ParseADTSHeader(adtshdr); err != nil {
			return
		}
		if inst.codecdata, err = aacparser.NewCodecDataFromMPEG4AudioConfig(config); err != nil {
			return
		}
	}
	streams = []av.CodecData{inst.codecdata}
	return
}

// ReadPacket func
func (inst *Demuxer) ReadPacket() (pkt av.Packet, err error) {
	var adtshdr []byte
	var config aacparser.MPEG4AudioConfig
	var hdrlen, framelen, samples int
	if adtshdr, err = inst.r.Peek(9); err != nil {
		return
	}
	if config, hdrlen, framelen, samples, err = aacparser.ParseADTSHeader(adtshdr); err != nil {
		return
	}

	pkt.Data = make([]byte, framelen)
	if _, err = io.ReadFull(inst.r, pkt.Data); err != nil {
		return
	}
	pkt.Data = pkt.Data[hdrlen:]

	pkt.Time = inst.ts
	inst.ts += time.Duration(samples) * time.Second / time.Duration(config.SampleRate)
	return
}

// Handler func
func Handler(h *avutil.RegisterHandler) {
	h.Ext = ".aac"

	h.ReaderDemuxer = func(r io.Reader) av.Demuxer {
		return NewDemuxer(r)
	}

	h.WriterMuxer = func(w io.Writer) av.Muxer {
		return NewMuxer(w)
	}

	h.Probe = func(b []byte) bool {
		_, _, _, _, err := aacparser.ParseADTSHeader(b)
		return err == nil
	}

	h.CodecTypes = []av.CodecType{av.AAC}
}
