package mp4

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/Youngju-Heo/gomedia/core/media/av"
	"github.com/Youngju-Heo/gomedia/core/media/codec/aacparser"
	"github.com/Youngju-Heo/gomedia/core/media/codec/h264parser"
	"github.com/Youngju-Heo/gomedia/core/media/format/mp4/mp4io"
)

// Demuxer type
type Demuxer struct {
	r         io.ReadSeeker
	streams   []*Stream
	movieAtom *mp4io.Movie
}

// NewDemuxer type
func NewDemuxer(r io.ReadSeeker) *Demuxer {
	return &Demuxer{
		r: r,
	}
}

// Streams type
func (demuxer *Demuxer) Streams() (streams []av.CodecData, err error) {
	if err = demuxer.probe(); err != nil {
		return
	}
	for _, stream := range demuxer.streams {
		streams = append(streams, stream.CodecData)
	}
	return
}

func (demuxer *Demuxer) readat(pos int64, b []byte) (err error) {
	if _, err = demuxer.r.Seek(pos, 0); err != nil {
		return
	}
	if _, err = io.ReadFull(demuxer.r, b); err != nil {
		return
	}
	return
}

func (demuxer *Demuxer) probe() (err error) {
	if demuxer.movieAtom != nil {
		return
	}

	var moov *mp4io.Movie
	var atoms []mp4io.Atom

	if atoms, err = mp4io.ReadFileAtoms(demuxer.r); err != nil {
		return
	}
	if _, err = demuxer.r.Seek(0, 0); err != nil {
		return
	}

	for _, atom := range atoms {
		if atom.Tag() == mp4io.MOOV {
			moov = atom.(*mp4io.Movie)
		}
	}

	if moov == nil {
		err = fmt.Errorf("mp4: 'moov' atom not found")
		return
	}

	demuxer.streams = []*Stream{}
	for i, atrack := range moov.Tracks {
		stream := &Stream{
			trackAtom: atrack,
			demuxer:   demuxer,
			idx:       i,
		}
		if atrack.Media != nil && atrack.Media.Info != nil && atrack.Media.Info.Sample != nil {
			stream.sample = atrack.Media.Info.Sample
			stream.timeScale = int64(atrack.Media.Header.TimeScale)
		} else {
			err = fmt.Errorf("mp4: sample table not found")
			return
		}

		if avc1 := atrack.GetAVC1Conf(); avc1 != nil {
			if stream.CodecData, err = h264parser.NewCodecDataFromAVCDecoderConfRecord(avc1.Data); err != nil {
				return
			}
			demuxer.streams = append(demuxer.streams, stream)
		} else if esds := atrack.GetElemStreamDesc(); esds != nil {
			if stream.CodecData, err = aacparser.NewCodecDataFromMPEG4AudioConfigBytes(esds.DecConfig); err != nil {
				return
			}
			demuxer.streams = append(demuxer.streams, stream)
		}
	}

	demuxer.movieAtom = moov
	return
}

func (stream *Stream) setSampleIndex(index int) (err error) {
	found := false
	start := 0
	stream.chunkGroupIndex = 0

	for stream.chunkIndex = range stream.sample.ChunkOffset.Entries {
		if stream.chunkGroupIndex+1 < len(stream.sample.SampleToChunk.Entries) &&
			uint32(stream.chunkIndex+1) == stream.sample.SampleToChunk.Entries[stream.chunkGroupIndex+1].FirstChunk {
			stream.chunkGroupIndex++
		}
		n := int(stream.sample.SampleToChunk.Entries[stream.chunkGroupIndex].SamplesPerChunk)
		if index >= start && index < start+n {
			found = true
			stream.sampleIndexInChunk = index - start
			break
		}
		start += n
	}
	if !found {
		err = fmt.Errorf("mp4: stream[%d]: cannot locate sample index in chunk", stream.idx)
		return
	}

	if stream.sample.SampleSize.SampleSize != 0 {
		stream.sampleOffsetInChunk = int64(stream.sampleIndexInChunk) * int64(stream.sample.SampleSize.SampleSize)
	} else {
		if index >= len(stream.sample.SampleSize.Entries) {
			err = fmt.Errorf("mp4: stream[%d]: sample index out of range", stream.idx)
			return
		}
		stream.sampleOffsetInChunk = int64(0)
		for i := index - stream.sampleIndexInChunk; i < index; i++ {
			stream.sampleOffsetInChunk += int64(stream.sample.SampleSize.Entries[i])
		}
	}

	stream.dts = int64(0)
	start = 0
	found = false
	stream.sttsEntryIndex = 0
	for stream.sttsEntryIndex < len(stream.sample.TimeToSample.Entries) {
		entry := stream.sample.TimeToSample.Entries[stream.sttsEntryIndex]
		n := int(entry.Count)
		if index >= start && index < start+n {
			stream.sampleIndexInSttsEntry = index - start
			stream.dts += int64(index-start) * int64(entry.Duration)
			found = true
			break
		}
		start += n
		stream.dts += int64(n) * int64(entry.Duration)
		stream.sttsEntryIndex++
	}
	if !found {
		err = fmt.Errorf("mp4: stream[%d]: cannot locate sample index in stts entry", stream.idx)
		return
	}

	if stream.sample.CompositionOffset != nil && len(stream.sample.CompositionOffset.Entries) > 0 {
		start = 0
		found = false
		stream.cttsEntryIndex = 0
		for stream.cttsEntryIndex < len(stream.sample.CompositionOffset.Entries) {
			n := int(stream.sample.CompositionOffset.Entries[stream.cttsEntryIndex].Count)
			if index >= start && index < start+n {
				stream.sampleIndexInCttsEntry = index - start
				found = true
				break
			}
			start += n
			stream.cttsEntryIndex++
		}
		if !found {
			err = fmt.Errorf("mp4: stream[%d]: cannot locate sample index in ctts entry", stream.idx)
			return
		}
	}

	if stream.sample.SyncSample != nil {
		stream.syncSampleIndex = 0
		for stream.syncSampleIndex < len(stream.sample.SyncSample.Entries)-1 {
			if stream.sample.SyncSample.Entries[stream.syncSampleIndex+1]-1 > uint32(index) {
				break
			}
			stream.syncSampleIndex++
		}
	}

	if false {
		fmt.Printf("mp4: stream[%d]: setSampleIndex chunkGroupIndex=%d chunkIndex=%d sampleOffsetInChunk=%d\n",
			stream.idx, stream.chunkGroupIndex, stream.chunkIndex, stream.sampleOffsetInChunk)
	}

	stream.sampleIndex = index
	return
}

func (stream *Stream) isSampleValid() bool {
	if stream.chunkIndex >= len(stream.sample.ChunkOffset.Entries) {
		return false
	}
	if stream.chunkGroupIndex >= len(stream.sample.SampleToChunk.Entries) {
		return false
	}
	if stream.sttsEntryIndex >= len(stream.sample.TimeToSample.Entries) {
		return false
	}
	if stream.sample.CompositionOffset != nil && len(stream.sample.CompositionOffset.Entries) > 0 {
		if stream.cttsEntryIndex >= len(stream.sample.CompositionOffset.Entries) {
			return false
		}
	}
	if stream.sample.SyncSample != nil {
		if stream.syncSampleIndex >= len(stream.sample.SyncSample.Entries) {
			return false
		}
	}
	if stream.sample.SampleSize.SampleSize != 0 {
		if stream.sampleIndex >= len(stream.sample.SampleSize.Entries) {
			return false
		}
	}
	return true
}

func (stream *Stream) incSampleIndex() (duration int64) {
	if false {
		fmt.Printf("incSampleIndex sampleIndex=%d sampleOffsetInChunk=%d sampleIndexInChunk=%d chunkGroupIndex=%d chunkIndex=%d\n",
			stream.sampleIndex, stream.sampleOffsetInChunk, stream.sampleIndexInChunk, stream.chunkGroupIndex, stream.chunkIndex)
	}

	stream.sampleIndexInChunk++
	if uint32(stream.sampleIndexInChunk) == stream.sample.SampleToChunk.Entries[stream.chunkGroupIndex].SamplesPerChunk {
		stream.chunkIndex++
		stream.sampleIndexInChunk = 0
		stream.sampleOffsetInChunk = int64(0)
	} else {
		if stream.sample.SampleSize.SampleSize != 0 {
			stream.sampleOffsetInChunk += int64(stream.sample.SampleSize.SampleSize)
		} else {
			stream.sampleOffsetInChunk += int64(stream.sample.SampleSize.Entries[stream.sampleIndex])
		}
	}

	if stream.chunkGroupIndex+1 < len(stream.sample.SampleToChunk.Entries) &&
		uint32(stream.chunkIndex+1) == stream.sample.SampleToChunk.Entries[stream.chunkGroupIndex+1].FirstChunk {
		stream.chunkGroupIndex++
	}

	sttsEntry := stream.sample.TimeToSample.Entries[stream.sttsEntryIndex]
	duration = int64(sttsEntry.Duration)
	stream.sampleIndexInSttsEntry++
	stream.dts += duration
	if uint32(stream.sampleIndexInSttsEntry) == sttsEntry.Count {
		stream.sampleIndexInSttsEntry = 0
		stream.sttsEntryIndex++
	}

	if stream.sample.CompositionOffset != nil && len(stream.sample.CompositionOffset.Entries) > 0 {
		stream.sampleIndexInCttsEntry++
		if uint32(stream.sampleIndexInCttsEntry) == stream.sample.CompositionOffset.Entries[stream.cttsEntryIndex].Count {
			stream.sampleIndexInCttsEntry = 0
			stream.cttsEntryIndex++
		}
	}

	if stream.sample.SyncSample != nil {
		entries := stream.sample.SyncSample.Entries
		if stream.syncSampleIndex+1 < len(entries) && entries[stream.syncSampleIndex+1]-1 == uint32(stream.sampleIndex+1) {
			stream.syncSampleIndex++
		}
	}

	stream.sampleIndex++
	return
}

func (stream *Stream) sampleCount() int {
	if stream.sample.SampleSize.SampleSize == 0 {
		chunkGroupIndex := 0
		count := 0
		for chunkIndex := range stream.sample.ChunkOffset.Entries {
			n := int(stream.sample.SampleToChunk.Entries[chunkGroupIndex].SamplesPerChunk)
			count += n
			if chunkGroupIndex+1 < len(stream.sample.SampleToChunk.Entries) &&
				uint32(chunkIndex+1) == stream.sample.SampleToChunk.Entries[chunkGroupIndex+1].FirstChunk {
				chunkGroupIndex++
			}
		}
		return count
	}

	return len(stream.sample.SampleSize.Entries)
}

// ReadPacket type
func (demuxer *Demuxer) ReadPacket() (pkt av.Packet, err error) {
	if err = demuxer.probe(); err != nil {
		return
	}
	if len(demuxer.streams) == 0 {
		err = errors.New("mp4: no streams available while trying to read a packet")
		return
	}

	var chosen *Stream
	var chosenidx int
	for i, stream := range demuxer.streams {
		if chosen == nil || stream.tsToTime(stream.dts) < chosen.tsToTime(chosen.dts) {
			chosen = stream
			chosenidx = i
		}
	}
	if false {
		fmt.Printf("ReadPacket: chosen index=%v time=%v\n", chosen.idx, chosen.tsToTime(chosen.dts))
	}
	tm := chosen.tsToTime(chosen.dts)
	if pkt, err = chosen.readPacket(); err != nil {
		return
	}
	pkt.Time = tm
	pkt.Idx = int8(chosenidx)
	return
}

// CurrentTime type
func (demuxer *Demuxer) CurrentTime() (tm time.Duration) {
	if len(demuxer.streams) > 0 {
		stream := demuxer.streams[0]
		tm = stream.tsToTime(stream.dts)
	}
	return
}

// SeekToTime type
func (demuxer *Demuxer) SeekToTime(tm time.Duration) (err error) {
	for _, stream := range demuxer.streams {
		if stream.Type().IsVideo() {
			if err = stream.seekToTime(tm); err != nil {
				return
			}
			tm = stream.tsToTime(stream.dts)
			break
		}
	}

	for _, stream := range demuxer.streams {
		if !stream.Type().IsVideo() {
			if err = stream.seekToTime(tm); err != nil {
				return
			}
		}
	}

	return
}

func (stream *Stream) readPacket() (pkt av.Packet, err error) {
	if !stream.isSampleValid() {
		err = io.EOF
		return
	}
	//fmt.Println("readPacket", stream.sampleIndex)

	chunkOffset := stream.sample.ChunkOffset.Entries[stream.chunkIndex]
	sampleSize := uint32(0)
	if stream.sample.SampleSize.SampleSize != 0 {
		sampleSize = stream.sample.SampleSize.SampleSize
	} else {
		sampleSize = stream.sample.SampleSize.Entries[stream.sampleIndex]
	}

	sampleOffset := int64(chunkOffset) + stream.sampleOffsetInChunk
	pkt.Data = make([]byte, sampleSize)
	if err = stream.demuxer.readat(sampleOffset, pkt.Data); err != nil {
		return
	}

	if stream.sample.SyncSample != nil {
		if stream.sample.SyncSample.Entries[stream.syncSampleIndex]-1 == uint32(stream.sampleIndex) {
			pkt.IsKeyFrame = true
		}
	}

	//println("pts/dts", stream.ptsEntryIndex, stream.dtsEntryIndex)
	if stream.sample.CompositionOffset != nil && len(stream.sample.CompositionOffset.Entries) > 0 {
		cts := int64(stream.sample.CompositionOffset.Entries[stream.cttsEntryIndex].Offset)
		pkt.CompositionTime = stream.tsToTime(cts)
	}

	stream.incSampleIndex()

	return
}

func (stream *Stream) seekToTime(tm time.Duration) (err error) {
	index := stream.timeToSampleIndex(tm)
	if err = stream.setSampleIndex(index); err != nil {
		return
	}
	if false {
		fmt.Printf("stream[%d]: seekToTime index=%v time=%v cur=%v\n", stream.idx, index, tm, stream.tsToTime(stream.dts))
	}
	return
}

func (stream *Stream) timeToSampleIndex(tm time.Duration) int {
	targetTs := stream.timeToTs(tm)
	targetIndex := 0

	startTs := int64(0)
	endTs := int64(0)
	startIndex := 0
	endIndex := 0
	found := false
	for _, entry := range stream.sample.TimeToSample.Entries {
		endTs = startTs + int64(entry.Count*entry.Duration)
		endIndex = startIndex + int(entry.Count)
		if targetTs >= startTs && targetTs < endTs {
			targetIndex = startIndex + int((targetTs-startTs)/int64(entry.Duration))
			found = true
		}
		startTs = endTs
		startIndex = endIndex
	}
	if !found {
		if targetTs < 0 {
			targetIndex = 0
		} else {
			targetIndex = endIndex - 1
		}
	}

	if stream.sample.SyncSample != nil {
		entries := stream.sample.SyncSample.Entries
		for i := len(entries) - 1; i >= 0; i-- {
			if entries[i]-1 < uint32(targetIndex) {
				targetIndex = int(entries[i] - 1)
				break
			}
		}
	}

	return targetIndex
}
