// Package pktque provides packet Filter interface and structures used by other components.
package pktque

import (
	"time"

	"github.com/Youngju-Heo/gomedia/core/media/av"
)

// Filter Filter
type Filter interface {
	// Change packet time or drop packet
	ModifyPacket(pkt *av.Packet, streams []av.CodecData, videoidx int, audioidx int) (drop bool, err error)
}

// Filters Combine multiple Filters into one, ModifyPacket will be called in order.
type Filters []Filter

// ModifyPacket ModifyPacket
func (instance Filters) ModifyPacket(pkt *av.Packet, streams []av.CodecData, videoidx int, audioidx int) (drop bool, err error) {
	for _, filter := range instance {
		if drop, err = filter.ModifyPacket(pkt, streams, videoidx, audioidx); err != nil {
			return
		}
		if drop {
			return
		}
	}
	return
}

// FilterDemuxer Wrap origin Demuxer and Filter into a new Demuxer, when read this Demuxer filters will be called.
type FilterDemuxer struct {
	av.Demuxer
	Filter   Filter
	streams  []av.CodecData
	videoidx int
	audioidx int
}

// ReadPacket ReadPacket
func (instance FilterDemuxer) ReadPacket() (pkt av.Packet, err error) {
	if instance.streams == nil {
		if instance.streams, err = instance.Demuxer.Streams(); err != nil {
			return
		}
		for i, stream := range instance.streams {
			if stream.Type().IsVideo() {
				instance.videoidx = i
			} else if stream.Type().IsAudio() {
				instance.audioidx = i
			}
		}
	}

	for {
		if pkt, err = instance.Demuxer.ReadPacket(); err != nil {
			return
		}
		var drop bool
		if drop, err = instance.Filter.ModifyPacket(&pkt, instance.streams, instance.videoidx, instance.audioidx); err != nil {
			return
		}
		if !drop {
			break
		}
	}

	return
}

// WaitKeyFrame Drop packets until first video key frame arrived.
type WaitKeyFrame struct {
	ok bool
}

// ModifyPacket ModifyPacket
func (instance *WaitKeyFrame) ModifyPacket(pkt *av.Packet, streams []av.CodecData, videoidx int, audioidx int) (drop bool, err error) {
	if !instance.ok && pkt.Idx == int8(videoidx) && pkt.IsKeyFrame {
		instance.ok = true
	}
	drop = !instance.ok
	return
}

// FixTime Fix incorrect packet timestamps.
type FixTime struct {
	zerobase      time.Duration
	incrbase      time.Duration
	lasttime      time.Duration
	StartFromZero bool // make timestamp start from zero
	MakeIncrement bool // force timestamp increment
}

// ModifyPacket ModifyPacket
func (instance *FixTime) ModifyPacket(pkt *av.Packet, streams []av.CodecData, videoidx int, audioidx int) (drop bool, err error) {
	if instance.StartFromZero {
		if instance.zerobase == 0 {
			instance.zerobase = pkt.Time
		}
		pkt.Time -= instance.zerobase
	}

	if instance.MakeIncrement {
		pkt.Time -= instance.incrbase
		if instance.lasttime == 0 {
			instance.lasttime = pkt.Time
		}
		if pkt.Time < instance.lasttime || pkt.Time > instance.lasttime+time.Millisecond*500 {
			instance.incrbase += pkt.Time - instance.lasttime
			pkt.Time = instance.lasttime
		}
		instance.lasttime = pkt.Time
	}

	return
}

// AVSync Drop incorrect packets to make A/V sync.
type AVSync struct {
	MaxTimeDiff time.Duration
	time        []time.Duration
}

// ModifyPacket ModifyPacket
func (instance *AVSync) ModifyPacket(pkt *av.Packet, streams []av.CodecData, videoidx int, audioidx int) (drop bool, err error) {
	if instance.time == nil {
		instance.time = make([]time.Duration, len(streams))
		if instance.MaxTimeDiff == 0 {
			instance.MaxTimeDiff = time.Millisecond * 500
		}
	}

	start, end, correctable, correcttime := instance.check(int(pkt.Idx))
	if pkt.Time >= start && pkt.Time < end {
		instance.time[pkt.Idx] = pkt.Time
	} else {
		if correctable {
			pkt.Time = correcttime
			for i := range instance.time {
				instance.time[i] = correcttime
			}
		} else {
			drop = true
		}
	}
	return
}

func (instance *AVSync) check(i int) (start time.Duration, end time.Duration, correctable bool, correcttime time.Duration) {
	minidx := -1
	maxidx := -1
	for j := range instance.time {
		if minidx == -1 || instance.time[j] < instance.time[minidx] {
			minidx = j
		}
		if maxidx == -1 || instance.time[j] > instance.time[maxidx] {
			maxidx = j
		}
	}
	allthesame := instance.time[minidx] == instance.time[maxidx]

	if i == maxidx {
		if allthesame {
			correctable = true
		} else {
			correctable = false
		}
	} else {
		correctable = true
	}

	start = instance.time[minidx]
	end = start + instance.MaxTimeDiff
	correcttime = start + time.Millisecond*40
	return
}

// Walltime Make packets reading speed as same as walltime, effect like ffmpeg -re option.
type Walltime struct {
	firsttime time.Time
}

// ModifyPacket ModifyPacket
func (instance *Walltime) ModifyPacket(pkt *av.Packet, streams []av.CodecData, videoidx int, audioidx int) (drop bool, err error) {
	if pkt.Idx == 0 {
		if instance.firsttime.IsZero() {
			instance.firsttime = time.Now()
		}
		pkttime := instance.firsttime.Add(pkt.Time)
		delta := pkttime.Sub(time.Now())
		if delta > 0 {
			time.Sleep(delta)
		}
	}
	return
}
