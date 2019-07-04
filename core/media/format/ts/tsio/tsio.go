package tsio

import (
	"fmt"
	"io"
	"time"

	"github.com/Youngju-Heo/gomedia/core/media/utils/bits/pio"
)

const (
	// StreamIDH264 const
	StreamIDH264 = 0xe0
	// StreamIDAAC const
	StreamIDAAC = 0xc0
)

const (
	// PatPID const
	PatPID = 0
	// PmtPID const
	PmtPID = 0x1000
)

// TableIDPMT var
const TableIDPMT = 2

// TableExtPMT var
const TableExtPMT = 1

// TableIDPAT var
const TableIDPAT = 0

// TableExtPAT var
const TableExtPAT = 1

// MaxPESHeaderLength var
const MaxPESHeaderLength = 19

// MaxTSHeaderLength var
const MaxTSHeaderLength = 12

// ErrPESHeader func
var ErrPESHeader = fmt.Errorf("invalid PES header")

// ErrPSIHeader func
var ErrPSIHeader = fmt.Errorf("invalid PSI header")

// ErrParsePMT func
var ErrParsePMT = fmt.Errorf("invalid PMT")

// ErrParsePAT func
var ErrParsePAT = fmt.Errorf("invalid PAT")

const (
	// ElementaryStreamTypeH264 const
	ElementaryStreamTypeH264 = 0x1B
	// ElementaryStreamTypeAdtsAAC const
	ElementaryStreamTypeAdtsAAC = 0x0F
)

// PATEntry func
type PATEntry struct {
	ProgramNumber uint16
	NetworkPID    uint16
	ProgramMapPID uint16
}

// PAT func
type PAT struct {
	Entries []PATEntry
}

// Len func
func (inst PAT) Len() (n int) {
	return len(inst.Entries) * 4
}

// Marshal func
func (inst PAT) Marshal(b []byte) (n int) {
	for _, entry := range inst.Entries {
		pio.PutU16BE(b[n:], entry.ProgramNumber)
		n += 2
		if entry.ProgramNumber == 0 {
			pio.PutU16BE(b[n:], entry.NetworkPID&0x1fff|7<<13)
			n += 2
		} else {
			pio.PutU16BE(b[n:], entry.ProgramMapPID&0x1fff|7<<13)
			n += 2
		}
	}
	return
}

// Unmarshal func
func (inst *PAT) Unmarshal(b []byte) (n int, err error) {
	for n < len(b) {
		if n+4 <= len(b) {
			var entry PATEntry
			entry.ProgramNumber = pio.U16BE(b[n:])
			n += 2
			if entry.ProgramNumber == 0 {
				entry.NetworkPID = pio.U16BE(b[n:]) & 0x1fff
				n += 2
			} else {
				entry.ProgramMapPID = pio.U16BE(b[n:]) & 0x1fff
				n += 2
			}
			inst.Entries = append(inst.Entries, entry)
		} else {
			break
		}
	}
	if n < len(b) {
		err = ErrParsePAT
		return
	}
	return
}

// Descriptor func
type Descriptor struct {
	Tag  uint8
	Data []byte
}

// ElementaryStreamInfo func
type ElementaryStreamInfo struct {
	StreamType    uint8
	ElementaryPID uint16
	Descriptors   []Descriptor
}

// PMT func
type PMT struct {
	PCRPID                uint16
	ProgramDescriptors    []Descriptor
	ElementaryStreamInfos []ElementaryStreamInfo
}

// Len func
func (inst PMT) Len() (n int) {
	// 111(3)
	// PCRPID(13)
	n += 2

	// desclen(16)
	n += 2

	for _, desc := range inst.ProgramDescriptors {
		n += 2 + len(desc.Data)
	}

	for _, info := range inst.ElementaryStreamInfos {
		// streamType
		n++

		// Reserved(3)
		// Elementary PID(13)
		n += 2

		// Reserved(6)
		// ES Info length length(10)
		n += 2

		for _, desc := range info.Descriptors {
			n += 2 + len(desc.Data)
		}
	}

	return
}

func (inst PMT) fillDescs(b []byte, descs []Descriptor) (n int) {
	for _, desc := range descs {
		b[n] = desc.Tag
		n++
		b[n] = uint8(len(desc.Data))
		n++
		copy(b[n:], desc.Data)
		n += len(desc.Data)
	}
	return
}

// Marshal func
func (inst PMT) Marshal(b []byte) (n int) {
	// 111(3)
	// PCRPID(13)
	pio.PutU16BE(b[n:], inst.PCRPID|7<<13)
	n += 2

	hold := n
	n += 2
	pos := n
	n += inst.fillDescs(b[n:], inst.ProgramDescriptors)
	desclen := n - pos
	pio.PutU16BE(b[hold:], uint16(desclen)|0xf<<12)

	for _, info := range inst.ElementaryStreamInfos {
		b[n] = info.StreamType
		n++

		// Reserved(3)
		// Elementary PID(13)
		pio.PutU16BE(b[n:], info.ElementaryPID|7<<13)
		n += 2

		hold := n
		n += 2
		pos := n
		n += inst.fillDescs(b[n:], info.Descriptors)
		desclen := n - pos
		pio.PutU16BE(b[hold:], uint16(desclen)|0x3c<<10)
	}

	return
}

func (inst PMT) parseDescs(b []byte) (descs []Descriptor, err error) {
	n := 0
	for n < len(b) {
		if n+2 <= len(b) {
			desc := Descriptor{}
			desc.Tag = b[n]
			desc.Data = make([]byte, b[n+1])
			n += 2
			if n+len(desc.Data) < len(b) {
				copy(desc.Data, b[n:])
				descs = append(descs, desc)
				n += len(desc.Data)
			} else {
				break
			}
		} else {
			break
		}
	}
	if n < len(b) {
		err = ErrParsePMT
		return
	}
	return
}

// Unmarshal func
func (inst *PMT) Unmarshal(b []byte) (n int, err error) {
	if len(b) < n+4 {
		err = ErrParsePMT
		return
	}

	// 111(3)
	// PCRPID(13)
	inst.PCRPID = pio.U16BE(b[0:2]) & 0x1fff
	n += 2

	// Reserved(4)=0xf
	// Reserved(2)=0x0
	// Program info length(10)
	desclen := int(pio.U16BE(b[2:4]) & 0x3ff)
	n += 2

	if desclen > 0 {
		if len(b) < n+desclen {
			err = ErrParsePMT
			return
		}
		if inst.ProgramDescriptors, err = inst.parseDescs(b[n : n+desclen]); err != nil {
			return
		}
		n += desclen
	}

	for n < len(b) {
		if len(b) < n+5 {
			err = ErrParsePMT
			return
		}

		var info ElementaryStreamInfo
		info.StreamType = b[n]
		n++

		// Reserved(3)
		// Elementary PID(13)
		info.ElementaryPID = pio.U16BE(b[n:]) & 0x1fff
		n += 2

		// Reserved(6)
		// ES Info length(10)
		desclen := int(pio.U16BE(b[n:]) & 0x3ff)
		n += 2

		if desclen > 0 {
			if len(b) < n+desclen {
				err = ErrParsePMT
				return
			}
			if info.Descriptors, err = inst.parseDescs(b[n : n+desclen]); err != nil {
				return
			}
			n += desclen
		}

		inst.ElementaryStreamInfos = append(inst.ElementaryStreamInfos, info)
	}

	return
}

// ParsePSI func
func ParsePSI(h []byte) (tableid uint8, tableext uint16, hdrlen int, datalen int, err error) {
	if len(h) < 8 {
		err = ErrPSIHeader
		return
	}

	// pointer(8)
	pointer := h[0]
	hdrlen++
	if pointer > 0 {
		hdrlen += int(pointer)
		if len(h) < hdrlen {
			err = ErrPSIHeader
			return
		}
	}

	if len(h) < hdrlen+12 {
		err = ErrPSIHeader
		return
	}

	// table_id(8)
	tableid = h[hdrlen]
	hdrlen++

	// section_syntax_indicator(1)=1,private_bit(1)=0,reserved(2)=3,unused(2)=0,section_length(10)
	datalen = int(pio.U16BE(h[hdrlen:]))&0x3ff - 9
	hdrlen += 2

	if datalen < 0 {
		err = ErrPSIHeader
		return
	}

	// Table ID extension(16)
	tableext = pio.U16BE(h[hdrlen:])
	hdrlen += 2

	// resverd(2)=3
	// version(5)
	// Current_next_indicator(1)
	hdrlen++

	// section_number(8)
	hdrlen++

	// last_section_number(8)
	hdrlen++

	// data

	// crc(32)

	return
}

// PSIHeaderLength var
const PSIHeaderLength = 9

// FillPSI func
func FillPSI(h []byte, tableid uint8, tableext uint16, datalen int) (n int) {
	// pointer(8)
	h[n] = 0
	n++

	// table_id(8)
	h[n] = tableid
	n++

	// section_syntax_indicator(1)=1,private_bit(1)=0,reserved(2)=3,unused(2)=0,section_length(10)
	pio.PutU16BE(h[n:], uint16(0xa<<12|2+3+4+datalen))
	n += 2

	// Table ID extension(16)
	pio.PutU16BE(h[n:], tableext)
	n += 2

	// resverd(2)=3,version(5)=0,Current_next_indicator(1)=1
	h[n] = 0x3<<6 | 1
	n++

	// section_number(8)
	h[n] = 0
	n++

	// last_section_number(8)
	h[n] = 0
	n++

	n += datalen

	crc := calcCRC32(0xffffffff, h[1:n])
	pio.PutU32LE(h[n:], crc)
	n += 4

	return
}

// TimeToPCR func
func TimeToPCR(tm time.Duration) (pcr uint64) {
	// base(33)+resverd(6)+ext(9)
	ts := uint64(tm * PcrHZ / time.Second)
	base := ts / 300
	ext := ts % 300
	pcr = base<<15 | 0x3f<<9 | ext
	return
}

// PCRToTime func
func PCRToTime(pcr uint64) (tm time.Duration) {
	base := pcr >> 15
	ext := pcr & 0x1ff
	ts := base*300 + ext
	tm = time.Duration(ts) * time.Second / time.Duration(PcrHZ)
	return
}

// TimeToTs func
func TimeToTs(tm time.Duration) (v uint64) {
	ts := uint64(tm * PtsHZ / time.Second)
	// 0010	PTS 32..30 1	PTS 29..15 1 PTS 14..00 1
	v = ((ts>>30)&0x7)<<33 | ((ts>>15)&0x7fff)<<17 | (ts&0x7fff)<<1 | 0x100010001
	return
}

// TsToTime func
func TsToTime(v uint64) (tm time.Duration) {
	// 0010	PTS 32..30 1	PTS 29..15 1 PTS 14..00 1
	ts := (((v >> 33) & 0x7) << 30) | (((v >> 17) & 0x7fff) << 15) | ((v >> 1) & 0x7fff)
	tm = time.Duration(ts) * time.Second / time.Duration(PtsHZ)
	return
}

const (
	// PtsHZ const
	PtsHZ = 90000
	// PcrHZ const
	PcrHZ = 27000000
)

// ParsePESHeader func
func ParsePESHeader(h []byte) (hdrlen int, streamid uint8, datalen int, pts, dts time.Duration, err error) {
	if h[0] != 0 || h[1] != 0 || h[2] != 1 {
		err = ErrPESHeader
		return
	}
	streamid = h[3]

	flags := h[7]
	hdrlen = int(h[8]) + 9

	datalen = int(pio.U16BE(h[4:6]))
	if datalen > 0 {
		datalen -= int(h[8]) + 3
	}

	const PTS = 1 << 7
	const DTS = 1 << 6

	if flags&PTS != 0 {
		if len(h) < 14 {
			err = ErrPESHeader
			return
		}
		pts = TsToTime(pio.U40BE(h[9:14]))
		if flags&DTS != 0 {
			if len(h) < 19 {
				err = ErrPESHeader
				return
			}
			dts = TsToTime(pio.U40BE(h[14:19]))
		}
	}

	return
}

// FillPESHeader func
func FillPESHeader(h []byte, streamid uint8, datalen int, pts, dts time.Duration) (n int) {
	h[0] = 0
	h[1] = 0
	h[2] = 1
	h[3] = streamid

	const PTS = 1 << 7
	const DTS = 1 << 6

	var flags uint8
	if pts != 0 {
		flags |= PTS
		if dts != 0 {
			flags |= DTS
		}
	}

	if flags&PTS != 0 {
		n += 5
	}
	if flags&DTS != 0 {
		n += 5
	}

	// packet_length(16) if zero then variable length
	// Specifies the number of bytes remaining in the packet after this field. Can be zero.
	// If the PES packet length is set to zero, the PES packet can be of any length.
	// A value of zero for the PES packet length can be used only when the PES packet payload is a **video** elementary stream.
	var pktlen uint16
	if datalen >= 0 {
		pktlen = uint16(datalen + n + 3)
	}
	pio.PutU16BE(h[4:6], pktlen)

	h[6] = 2<<6 | 1 // resverd(6,2)=2,original_or_copy(0,1)=1
	h[7] = flags
	h[8] = uint8(n)

	// pts(40)?
	// dts(40)?
	if flags&PTS != 0 {
		if flags&DTS != 0 {
			pio.PutU40BE(h[9:14], TimeToTs(pts)|3<<36)
			pio.PutU40BE(h[14:19], TimeToTs(dts)|1<<36)
		} else {
			pio.PutU40BE(h[9:14], TimeToTs(pts)|2<<36)
		}
	}

	n += 9
	return
}

// TSWriter func
type TSWriter struct {
	w                 io.Writer
	ContinuityCounter uint
	tshdr             []byte
}

// NewTSWriter func
func NewTSWriter(pid uint16) *TSWriter {
	w := &TSWriter{}
	w.tshdr = make([]byte, 188)
	w.tshdr[0] = 0x47
	pio.PutU16BE(w.tshdr[1:3], pid&0x1fff)
	for i := 6; i < 188; i++ {
		w.tshdr[i] = 0xff
	}
	return w
}

// WritePackets func
func (inst *TSWriter) WritePackets(w io.Writer, datav [][]byte, pcr time.Duration, sync bool, paddata bool) (err error) {
	datavlen := pio.VecLen(datav)
	writev := make([][]byte, len(datav))
	writepos := 0

	for writepos < datavlen {
		inst.tshdr[1] = inst.tshdr[1] & 0x1f
		inst.tshdr[3] = byte(inst.ContinuityCounter)&0xf | 0x30
		inst.tshdr[5] = 0 // flags
		hdrlen := 6
		inst.ContinuityCounter++

		if writepos == 0 {
			inst.tshdr[1] = 0x40 | inst.tshdr[1] // Payload Unit Start Indicator
			if pcr != 0 {
				hdrlen += 6
				inst.tshdr[5] = 0x10 | inst.tshdr[5] // PCR flag (Discontinuity indicator 0x80)
				pio.PutU48BE(inst.tshdr[6:12], TimeToPCR(pcr))
			}
			if sync {
				inst.tshdr[5] = 0x40 | inst.tshdr[5] // Random Access indicator
			}
		}

		padtail := 0
		end := writepos + 188 - hdrlen
		if end > datavlen {
			if paddata {
				padtail = end - datavlen
			} else {
				hdrlen += end - datavlen
			}
			end = datavlen
		}
		n := pio.VecSliceTo(datav, writev, writepos, end)

		inst.tshdr[4] = byte(hdrlen) - 5 // length
		if _, err = w.Write(inst.tshdr[:hdrlen]); err != nil {
			return
		}
		for i := 0; i < n; i++ {
			if _, err = w.Write(writev[i]); err != nil {
				return
			}
		}
		if padtail > 0 {
			if _, err = w.Write(inst.tshdr[188-padtail : 188]); err != nil {
				return
			}
		}

		writepos = end
	}

	return
}

// ParseTSHeader func
func ParseTSHeader(tshdr []byte) (pid uint16, start bool, iskeyframe bool, hdrlen int, err error) {
	// https://en.wikipedia.org/wiki/MPEG_transport_stream
	if tshdr[0] != 0x47 {
		err = fmt.Errorf("tshdr sync invalid")
		return
	}
	pid = uint16((tshdr[1]&0x1f))<<8 | uint16(tshdr[2])
	start = tshdr[1]&0x40 != 0
	hdrlen += 4
	if tshdr[3]&0x20 != 0 {
		hdrlen += int(tshdr[4]) + 1
		iskeyframe = tshdr[5]&0x40 != 0
	}
	return
}
