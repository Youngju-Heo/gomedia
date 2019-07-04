package ts

import (
	"fmt"
	"io"
	"time"

	"github.com/Youngju-Heo/gomedia/core/media/av"
	"github.com/Youngju-Heo/gomedia/core/media/codec/aacparser"
	"github.com/Youngju-Heo/gomedia/core/media/codec/h264parser"
	"github.com/Youngju-Heo/gomedia/core/media/format/ts/tsio"
)

// CodecTypes var
var CodecTypes = []av.CodecType{av.H264, av.AAC}

// Muxer struct
type Muxer struct {
	w                        io.Writer
	streams                  []*Stream
	PaddingToMakeCounterCont bool

	psidata []byte
	peshdr  []byte
	tshdr   []byte
	adtshdr []byte
	datav   [][]byte
	nalus   [][]byte

	tswpat, tswpmt *tsio.TSWriter
}

// NewMuxer func
func NewMuxer(w io.Writer) *Muxer {
	return &Muxer{
		w:       w,
		psidata: make([]byte, 188),
		peshdr:  make([]byte, tsio.MaxPESHeaderLength),
		tshdr:   make([]byte, tsio.MaxTSHeaderLength),
		adtshdr: make([]byte, aacparser.ADTSHeaderLength),
		nalus:   make([][]byte, 16),
		datav:   make([][]byte, 16),
		tswpmt:  tsio.NewTSWriter(tsio.PmtPID),
		tswpat:  tsio.NewTSWriter(tsio.PatPID),
	}
}

func (inst *Muxer) newStream(codec av.CodecData) (err error) {
	ok := false
	for _, c := range CodecTypes {
		if codec.Type() == c {
			ok = true
			break
		}
	}
	if !ok {
		err = fmt.Errorf("ts: codec type=%s is not supported", codec.Type())
		return
	}

	pid := uint16(len(inst.streams) + 0x100)
	stream := &Stream{
		muxer:     inst,
		CodecData: codec,
		pid:       pid,
		tsw:       tsio.NewTSWriter(pid),
	}
	inst.streams = append(inst.streams, stream)
	return
}

func (inst *Muxer) writePaddingTSPackets(tsw *tsio.TSWriter) (err error) {
	for tsw.ContinuityCounter&0xf != 0x0 {
		if err = tsw.WritePackets(inst.w, inst.datav[:0], 0, false, true); err != nil {
			return
		}
	}
	return
}

// WriteTrailer func
func (inst *Muxer) WriteTrailer() (err error) {
	if inst.PaddingToMakeCounterCont {
		for _, stream := range inst.streams {
			if err = inst.writePaddingTSPackets(stream.tsw); err != nil {
				return
			}
		}
	}
	return
}

// SetWriter func
func (inst *Muxer) SetWriter(w io.Writer) {
	inst.w = w
	return
}

// WritePATPMT func
func (inst *Muxer) WritePATPMT() (err error) {
	pat := tsio.PAT{
		Entries: []tsio.PATEntry{
			{ProgramNumber: 1, ProgramMapPID: tsio.PmtPID},
		},
	}
	patlen := pat.Marshal(inst.psidata[tsio.PSIHeaderLength:])
	n := tsio.FillPSI(inst.psidata, tsio.TableIDPAT, tsio.TableExtPAT, patlen)
	inst.datav[0] = inst.psidata[:n]
	if err = inst.tswpat.WritePackets(inst.w, inst.datav[:1], 0, false, true); err != nil {
		return
	}

	var elemStreams []tsio.ElementaryStreamInfo
	for _, stream := range inst.streams {
		switch stream.Type() {
		case av.AAC:
			elemStreams = append(elemStreams, tsio.ElementaryStreamInfo{
				StreamType:    tsio.ElementaryStreamTypeAdtsAAC,
				ElementaryPID: stream.pid,
			})
		case av.H264:
			elemStreams = append(elemStreams, tsio.ElementaryStreamInfo{
				StreamType:    tsio.ElementaryStreamTypeH264,
				ElementaryPID: stream.pid,
			})
		}
	}

	pmt := tsio.PMT{
		PCRPID:                0x100,
		ElementaryStreamInfos: elemStreams,
	}
	pmtlen := pmt.Len()
	if pmtlen+tsio.PSIHeaderLength > len(inst.psidata) {
		err = fmt.Errorf("ts: pmt too large")
		return
	}
	pmt.Marshal(inst.psidata[tsio.PSIHeaderLength:])
	n = tsio.FillPSI(inst.psidata, tsio.TableIDPMT, tsio.TableExtPMT, pmtlen)
	inst.datav[0] = inst.psidata[:n]
	if err = inst.tswpmt.WritePackets(inst.w, inst.datav[:1], 0, false, true); err != nil {
		return
	}

	return
}

// WriteHeader func
func (inst *Muxer) WriteHeader(streams []av.CodecData) (err error) {
	inst.streams = []*Stream{}
	for _, stream := range streams {
		if err = inst.newStream(stream); err != nil {
			return
		}
	}

	if err = inst.WritePATPMT(); err != nil {
		return
	}
	return
}

// WritePacket func
func (inst *Muxer) WritePacket(pkt av.Packet) (err error) {
	stream := inst.streams[pkt.Idx]
	pkt.Time += time.Second

	switch stream.Type() {
	case av.AAC:
		codec := stream.CodecData.(aacparser.CodecData)

		n := tsio.FillPESHeader(inst.peshdr, tsio.StreamIDAAC, len(inst.adtshdr)+len(pkt.Data), pkt.Time, 0)
		inst.datav[0] = inst.peshdr[:n]
		aacparser.FillADTSHeader(inst.adtshdr, codec.Config, 1024, len(pkt.Data))
		inst.datav[1] = inst.adtshdr
		inst.datav[2] = pkt.Data

		if err = stream.tsw.WritePackets(inst.w, inst.datav[:3], pkt.Time, true, false); err != nil {
			return
		}

	case av.H264:
		codec := stream.CodecData.(h264parser.CodecData)

		nalus := inst.nalus[:0]
		if pkt.IsKeyFrame {
			nalus = append(nalus, codec.SPS())
			nalus = append(nalus, codec.PPS())
		}
		pktnalus, _ := h264parser.SplitNALUs(pkt.Data)
		for _, nalu := range pktnalus {
			nalus = append(nalus, nalu)
		}

		datav := inst.datav[:1]
		for i, nalu := range nalus {
			if i == 0 {
				datav = append(datav, h264parser.AUDBytes)
			} else {
				datav = append(datav, h264parser.StartCodeBytes)
			}
			datav = append(datav, nalu)
		}

		n := tsio.FillPESHeader(inst.peshdr, tsio.StreamIDH264, -1, pkt.Time+pkt.CompositionTime, pkt.Time)
		datav[0] = inst.peshdr[:n]

		if err = stream.tsw.WritePackets(inst.w, datav, pkt.Time, pkt.IsKeyFrame, false); err != nil {
			return
		}
	}

	return
}
