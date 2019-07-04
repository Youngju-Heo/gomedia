package ts

import (
	"bufio"
	"fmt"
	"io"
	"time"

	"github.com/Youngju-Heo/gomedia/core/media/av"
	"github.com/Youngju-Heo/gomedia/core/media/codec/aacparser"
	"github.com/Youngju-Heo/gomedia/core/media/codec/h264parser"
	"github.com/Youngju-Heo/gomedia/core/media/format/ts/tsio"
	"github.com/Youngju-Heo/gomedia/core/media/utils/bits/pio"
)

// Demuxer type
type Demuxer struct {
	r *bufio.Reader

	pkts []av.Packet

	pat     *tsio.PAT
	pmt     *tsio.PMT
	streams []*Stream
	tshdr   []byte

	stage int
}

// NewDemuxer func
func NewDemuxer(r io.Reader) *Demuxer {
	return &Demuxer{
		tshdr: make([]byte, 188),
		r:     bufio.NewReaderSize(r, pio.RecommendBufioSize),
	}
}

// Streams func
func (demuxer *Demuxer) Streams() (streams []av.CodecData, err error) {
	if err = demuxer.probe(); err != nil {
		return
	}
	for _, stream := range demuxer.streams {
		streams = append(streams, stream.CodecData)
	}
	return
}

func (demuxer *Demuxer) probe() (err error) {
	if demuxer.stage == 0 {
		for {
			if demuxer.pmt != nil {
				n := 0
				for _, stream := range demuxer.streams {
					if stream.CodecData != nil {
						n++
					}
				}
				if n == len(demuxer.streams) {
					break
				}
			}
			if err = demuxer.poll(); err != nil {
				return
			}
		}
		demuxer.stage++
	}
	return
}

// ReadPacket func
func (demuxer *Demuxer) ReadPacket() (pkt av.Packet, err error) {
	if err = demuxer.probe(); err != nil {
		return
	}

	for len(demuxer.pkts) == 0 {
		if err = demuxer.poll(); err != nil {
			return
		}
	}

	pkt = demuxer.pkts[0]
	demuxer.pkts = demuxer.pkts[1:]
	return
}

func (demuxer *Demuxer) poll() (err error) {
	if err = demuxer.readTSPacket(); err == io.EOF {
		var n int
		if n, err = demuxer.payloadEnd(); err != nil {
			return
		}
		if n == 0 {
			err = io.EOF
		}
	}
	return
}

func (demuxer *Demuxer) initPMT(payload []byte) (err error) {
	var psihdrlen int
	var datalen int
	if _, _, psihdrlen, datalen, err = tsio.ParsePSI(payload); err != nil {
		return
	}
	demuxer.pmt = &tsio.PMT{}
	if _, err = demuxer.pmt.Unmarshal(payload[psihdrlen : psihdrlen+datalen]); err != nil {
		return
	}

	demuxer.streams = []*Stream{}
	for i, info := range demuxer.pmt.ElementaryStreamInfos {
		stream := &Stream{}
		stream.idx = i
		stream.demuxer = demuxer
		stream.pid = info.ElementaryPID
		stream.streamType = info.StreamType
		switch info.StreamType {
		case tsio.ElementaryStreamTypeH264:
			demuxer.streams = append(demuxer.streams, stream)
		case tsio.ElementaryStreamTypeAdtsAAC:
			demuxer.streams = append(demuxer.streams, stream)
		}
	}
	return
}

func (demuxer *Demuxer) payloadEnd() (n int, err error) {
	for _, stream := range demuxer.streams {
		var i int
		if i, err = stream.payloadEnd(); err != nil {
			return
		}
		n += i
	}
	return
}

func (demuxer *Demuxer) readTSPacket() (err error) {
	var hdrlen int
	var pid uint16
	var start bool
	var iskeyframe bool

	if _, err = io.ReadFull(demuxer.r, demuxer.tshdr); err != nil {
		return
	}

	if pid, start, iskeyframe, hdrlen, err = tsio.ParseTSHeader(demuxer.tshdr); err != nil {
		return
	}
	payload := demuxer.tshdr[hdrlen:]

	if demuxer.pat == nil {
		if pid == 0 {
			var psihdrlen int
			var datalen int
			if _, _, psihdrlen, datalen, err = tsio.ParsePSI(payload); err != nil {
				return
			}
			demuxer.pat = &tsio.PAT{}
			if _, err = demuxer.pat.Unmarshal(payload[psihdrlen : psihdrlen+datalen]); err != nil {
				return
			}
		}
	} else if demuxer.pmt == nil {
		for _, entry := range demuxer.pat.Entries {
			if entry.ProgramMapPID == pid {
				if err = demuxer.initPMT(payload); err != nil {
					return
				}
				break
			}
		}
	} else {
		for _, stream := range demuxer.streams {
			if pid == stream.pid {
				if err = stream.handleTSPacket(start, iskeyframe, payload); err != nil {
					return
				}
				break
			}
		}
	}

	return
}

func (stream *Stream) addPacket(payload []byte, timedelta time.Duration) {
	dts := stream.dts
	pts := stream.pts
	if dts == 0 {
		dts = pts
	}

	demuxer := stream.demuxer
	pkt := av.Packet{
		Idx:        int8(stream.idx),
		IsKeyFrame: stream.iskeyframe,
		Time:       dts + timedelta,
		Data:       payload,
	}
	if pts != dts {
		pkt.CompositionTime = pts - dts
	}
	demuxer.pkts = append(demuxer.pkts, pkt)
}

func (stream *Stream) payloadEnd() (n int, err error) {
	payload := stream.data
	if payload == nil {
		return
	}
	if stream.datalen != 0 && len(payload) != stream.datalen {
		err = fmt.Errorf("ts: packet size mismatch size=%d correct=%d", len(payload), stream.datalen)
		return
	}
	stream.data = nil

	switch stream.streamType {
	case tsio.ElementaryStreamTypeAdtsAAC:
		var config aacparser.MPEG4AudioConfig

		delta := time.Duration(0)
		for len(payload) > 0 {
			var hdrlen, framelen, samples int
			if config, hdrlen, framelen, samples, err = aacparser.ParseADTSHeader(payload); err != nil {
				return
			}
			if stream.CodecData == nil {
				if stream.CodecData, err = aacparser.NewCodecDataFromMPEG4AudioConfig(config); err != nil {
					return
				}
			}
			stream.addPacket(payload[hdrlen:framelen], delta)
			n++
			delta += time.Duration(samples) * time.Second / time.Duration(config.SampleRate)
			payload = payload[framelen:]
		}

	case tsio.ElementaryStreamTypeH264:
		nalus, _ := h264parser.SplitNALUs(payload)
		var sps, pps []byte
		for _, nalu := range nalus {
			if len(nalu) > 0 {
				naltype := nalu[0] & 0x1f
				switch {
				case naltype == 7:
					sps = nalu
				case naltype == 8:
					pps = nalu
				case h264parser.IsDataNALU(nalu):
					// raw nalu to avcc
					b := make([]byte, 4+len(nalu))
					pio.PutU32BE(b[0:4], uint32(len(nalu)))
					copy(b[4:], nalu)
					stream.addPacket(b, time.Duration(0))
					n++
				}
			}
		}

		if stream.CodecData == nil && len(sps) > 0 && len(pps) > 0 {
			if stream.CodecData, err = h264parser.NewCodecDataFromSPSAndPPS(sps, pps); err != nil {
				return
			}
		}
	}

	return
}

func (stream *Stream) handleTSPacket(start bool, iskeyframe bool, payload []byte) (err error) {
	if start {
		if _, err = stream.payloadEnd(); err != nil {
			return
		}
		var hdrlen int
		if hdrlen, _, stream.datalen, stream.pts, stream.dts, err = tsio.ParsePESHeader(payload); err != nil {
			return
		}
		stream.iskeyframe = iskeyframe
		if stream.datalen == 0 {
			stream.data = make([]byte, 0, 4096)
		} else {
			stream.data = make([]byte, 0, stream.datalen)
		}
		stream.data = append(stream.data, payload[hdrlen:]...)
	} else {
		stream.data = append(stream.data, payload...)
	}
	return
}
