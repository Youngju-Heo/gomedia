package mp4

import (
	"bufio"
	"fmt"
	"io"
	"time"

	"github.com/Youngju-Heo/gomedia/core/media/av"
	"github.com/Youngju-Heo/gomedia/core/media/codec/aacparser"
	"github.com/Youngju-Heo/gomedia/core/media/codec/h264parser"
	"github.com/Youngju-Heo/gomedia/core/media/format/mp4/mp4io"
	"github.com/Youngju-Heo/gomedia/core/media/utils/bits/pio"
)

// Muxer type
type Muxer struct {
	w       io.WriteSeeker
	bufw    *bufio.Writer
	wpos    int64
	streams []*Stream
}

// NewMuxer type
func NewMuxer(w io.WriteSeeker) *Muxer {
	return &Muxer{
		w:    w,
		bufw: bufio.NewWriterSize(w, pio.RecommendBufioSize),
	}
}

func (init *Muxer) newStream(codec av.CodecData) (err error) {
	switch codec.Type() {
	case av.H264, av.AAC:

	default:
		err = fmt.Errorf("mp4: codec type=%v is not supported", codec.Type())
		return
	}
	stream := &Stream{CodecData: codec}

	stream.sample = &mp4io.SampleTable{
		SampleDesc:   &mp4io.SampleDesc{},
		TimeToSample: &mp4io.TimeToSample{},
		SampleToChunk: &mp4io.SampleToChunk{
			Entries: []mp4io.SampleToChunkEntry{
				{
					FirstChunk:      1,
					SampleDescID:    1,
					SamplesPerChunk: 1,
				},
			},
		},
		SampleSize:  &mp4io.SampleSize{},
		ChunkOffset: &mp4io.ChunkOffset{},
	}

	stream.trackAtom = &mp4io.Track{
		Header: &mp4io.TrackHeader{
			TrackID:  int32(len(init.streams) + 1),
			Flags:    0x0003, // Track enabled | Track in movie
			Duration: 0,      // fill later
			Matrix:   [9]int32{0x10000, 0, 0, 0, 0x10000, 0, 0, 0, 0x40000000},
		},
		Media: &mp4io.Media{
			Header: &mp4io.MediaHeader{
				TimeScale: 0, // fill later
				Duration:  0, // fill later
				Language:  21956,
			},
			Info: &mp4io.MediaInfo{
				Sample: stream.sample,
				Data: &mp4io.DataInfo{
					Refer: &mp4io.DataRefer{
						URL: &mp4io.DataReferURL{
							Flags: 0x000001, // init reference
						},
					},
				},
			},
		},
	}

	switch codec.Type() {
	case av.H264:
		stream.sample.SyncSample = &mp4io.SyncSample{}
	}

	stream.timeScale = 90000
	stream.muxer = init
	init.streams = append(init.streams, stream)

	return
}

func (init *Stream) fillTrackAtom() (err error) {
	init.trackAtom.Media.Header.TimeScale = int32(init.timeScale)
	init.trackAtom.Media.Header.Duration = int32(init.duration)

	if init.Type() == av.H264 {
		codec := init.CodecData.(h264parser.CodecData)
		width, height := codec.Width(), codec.Height()
		init.sample.SampleDesc.AVC1Desc = &mp4io.AVC1Desc{
			DataRefIdx:           1,
			HorizontalResolution: 72,
			VorizontalResolution: 72,
			Width:                int16(width),
			Height:               int16(height),
			FrameCount:           1,
			Depth:                24,
			ColorTableID:         -1,
			Conf:                 &mp4io.AVC1Conf{Data: codec.AVCDecoderConfRecordBytes()},
		}
		init.trackAtom.Media.Handler = &mp4io.HandlerRefer{
			SubType: [4]byte{'v', 'i', 'd', 'e'},
			Name:    []byte("Video Media Handler"),
		}
		init.trackAtom.Media.Info.Video = &mp4io.VideoMediaInfo{
			Flags: 0x000001,
		}
		init.trackAtom.Header.TrackWidth = float64(width)
		init.trackAtom.Header.TrackHeight = float64(height)

	} else if init.Type() == av.AAC {
		codec := init.CodecData.(aacparser.CodecData)
		init.sample.SampleDesc.MP4ADesc = &mp4io.MP4ADesc{
			DataRefIdx:       1,
			NumberOfChannels: int16(codec.ChannelLayout().Count()),
			SampleSize:       int16(codec.SampleFormat().BytesPerSample()),
			SampleRate:       float64(codec.SampleRate()),
			Conf: &mp4io.ElemStreamDesc{
				DecConfig: codec.MPEG4AudioConfigBytes(),
			},
		}
		init.trackAtom.Header.Volume = 1
		init.trackAtom.Header.AlternateGroup = 1
		init.trackAtom.Media.Handler = &mp4io.HandlerRefer{
			SubType: [4]byte{'s', 'o', 'u', 'n'},
			Name:    []byte("Sound Handler"),
		}
		init.trackAtom.Media.Info.Sound = &mp4io.SoundMediaInfo{}

	} else {
		err = fmt.Errorf("mp4: codec type=%d invalid", init.Type())
	}

	return
}

// WriteHeader type
func (init *Muxer) WriteHeader(streams []av.CodecData) (err error) {
	init.streams = []*Stream{}
	for _, stream := range streams {
		if err = init.newStream(stream); err != nil {
			return
		}
	}

	taghdr := make([]byte, 8)
	pio.PutU32BE(taghdr[4:], uint32(mp4io.MDAT))
	if _, err = init.w.Write(taghdr); err != nil {
		return
	}
	init.wpos += 8

	for _, stream := range init.streams {
		if stream.Type().IsVideo() {
			stream.sample.CompositionOffset = &mp4io.CompositionOffset{}
		}
	}
	return
}

// WritePacket type
func (init *Muxer) WritePacket(pkt av.Packet) (err error) {
	stream := init.streams[pkt.Idx]
	if stream.lastpkt != nil {
		if err = stream.writePacket(*stream.lastpkt, pkt.Time-stream.lastpkt.Time); err != nil {
			return
		}
	}
	stream.lastpkt = &pkt
	return
}

func (init *Stream) writePacket(pkt av.Packet, rawdur time.Duration) (err error) {
	if rawdur < 0 {
		err = fmt.Errorf("mp4: stream#%d time=%v < lasttime=%v", pkt.Idx, pkt.Time, init.lastpkt.Time)
		return
	}

	if _, err = init.muxer.bufw.Write(pkt.Data); err != nil {
		return
	}

	if pkt.IsKeyFrame && init.sample.SyncSample != nil {
		init.sample.SyncSample.Entries = append(init.sample.SyncSample.Entries, uint32(init.sampleIndex+1))
	}

	duration := uint32(init.timeToTs(rawdur))
	if init.sttsEntry == nil || duration != init.sttsEntry.Duration {
		init.sample.TimeToSample.Entries = append(init.sample.TimeToSample.Entries, mp4io.TimeToSampleEntry{Duration: duration})
		init.sttsEntry = &init.sample.TimeToSample.Entries[len(init.sample.TimeToSample.Entries)-1]
	}
	init.sttsEntry.Count++

	if init.sample.CompositionOffset != nil {
		offset := uint32(init.timeToTs(pkt.CompositionTime))
		if init.cttsEntry == nil || offset != init.cttsEntry.Offset {
			table := init.sample.CompositionOffset
			table.Entries = append(table.Entries, mp4io.CompositionOffsetEntry{Offset: offset})
			init.cttsEntry = &table.Entries[len(table.Entries)-1]
		}
		init.cttsEntry.Count++
	}

	init.duration += int64(duration)
	init.sampleIndex++
	init.sample.ChunkOffset.Entries = append(init.sample.ChunkOffset.Entries, uint32(init.muxer.wpos))
	init.sample.SampleSize.Entries = append(init.sample.SampleSize.Entries, uint32(len(pkt.Data)))

	init.muxer.wpos += int64(len(pkt.Data))
	return
}

// WriteTrailer type
func (init *Muxer) WriteTrailer() (err error) {
	for _, stream := range init.streams {
		if stream.lastpkt != nil {
			if err = stream.writePacket(*stream.lastpkt, 0); err != nil {
				return
			}
			stream.lastpkt = nil
		}
	}

	moov := &mp4io.Movie{}
	moov.Header = &mp4io.MovieHeader{
		PreferredRate:   1,
		PreferredVolume: 1,
		Matrix:          [9]int32{0x10000, 0, 0, 0, 0x10000, 0, 0, 0, 0x40000000},
		NextTrackID:     2,
	}

	maxDur := time.Duration(0)
	timeScale := int64(10000)
	for _, stream := range init.streams {
		if err = stream.fillTrackAtom(); err != nil {
			return
		}
		dur := stream.tsToTime(stream.duration)
		stream.trackAtom.Header.Duration = int32(timeToTs(dur, timeScale))
		if dur > maxDur {
			maxDur = dur
		}
		moov.Tracks = append(moov.Tracks, stream.trackAtom)
	}
	moov.Header.TimeScale = int32(timeScale)
	moov.Header.Duration = int32(timeToTs(maxDur, timeScale))

	if err = init.bufw.Flush(); err != nil {
		return
	}

	var mdatsize int64
	if mdatsize, err = init.w.Seek(0, 1); err != nil {
		return
	}
	if _, err = init.w.Seek(0, 0); err != nil {
		return
	}
	taghdr := make([]byte, 4)
	pio.PutU32BE(taghdr, uint32(mdatsize))
	if _, err = init.w.Write(taghdr); err != nil {
		return
	}

	if _, err = init.w.Seek(0, 2); err != nil {
		return
	}
	b := make([]byte, moov.Len())
	moov.Marshal(b)
	if _, err = init.w.Write(b); err != nil {
		return
	}

	return
}
