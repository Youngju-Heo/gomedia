// Package pktque provides packet Filter interface and structures used by other components.
package pktque

import (
	"time"

	"../av"
)

// Filter Filter
type Filter interface {
	// Change packet time or drop packet
	ModifyPacket(pkt *av.Packet, streams []av.CodecData, videoidx int, audioidx int) (drop bool, err error)
}

// Filters Combine multiple Filters into one, ModifyPacket will be called in order.
type Filters []Filter

// ModifyPacket ModifyPacket
func (inst Filters) ModifyPacket(pkt *av.Packet, streams []av.CodecData, videoidx int, audioidx int) (drop bool, err error) {
	for _, filter := range inst {
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
func (inst FilterDemuxer) ReadPacket() (pkt av.Packet, err error) {
	if inst.streams == nil {
		if inst.streams, err = inst.Demuxer.Streams(); err != nil {
			return
		}
		for i, stream := range inst.streams {
			if stream.Type().IsVideo() {
				inst.videoidx = i
			} else if stream.Type().IsAudio() {
				inst.audioidx = i
			}
		}
	}

	for {
		if pkt, err = inst.Demuxer.ReadPacket(); err != nil {
			return
		}
		var drop bool
		if drop, err = inst.Filter.ModifyPacket(&pkt, inst.streams, inst.videoidx, inst.audioidx); err != nil {
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
func (inst *WaitKeyFrame) ModifyPacket(pkt *av.Packet, streams []av.CodecData, videoidx int, audioidx int) (drop bool, err error) {
	if !inst.ok && pkt.Idx == int8(videoidx) && pkt.IsKeyFrame {
		inst.ok = true
	}
	drop = !inst.ok
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
func (inst *FixTime) ModifyPacket(pkt *av.Packet, streams []av.CodecData, videoidx int, audioidx int) (drop bool, err error) {
	if inst.StartFromZero {
		if inst.zerobase == 0 {
			inst.zerobase = pkt.Time
		}
		pkt.Time -= inst.zerobase
	}

	if inst.MakeIncrement {
		pkt.Time -= inst.incrbase
		if inst.lasttime == 0 {
			inst.lasttime = pkt.Time
		}
		if pkt.Time < inst.lasttime || pkt.Time > inst.lasttime+time.Millisecond*500 {
			inst.incrbase += pkt.Time - inst.lasttime
			pkt.Time = inst.lasttime
		}
		inst.lasttime = pkt.Time
	}

	return
}

// AVSync Drop incorrect packets to make A/V sync.
type AVSync struct {
	MaxTimeDiff time.Duration
	time        []time.Duration
}

// ModifyPacket ModifyPacket
func (inst *AVSync) ModifyPacket(pkt *av.Packet, streams []av.CodecData, videoidx int, audioidx int) (drop bool, err error) {
	if inst.time == nil {
		inst.time = make([]time.Duration, len(streams))
		if inst.MaxTimeDiff == 0 {
			inst.MaxTimeDiff = time.Millisecond * 500
		}
	}

	start, end, correctable, correcttime := inst.check(int(pkt.Idx))
	if pkt.Time >= start && pkt.Time < end {
		inst.time[pkt.Idx] = pkt.Time
	} else {
		if correctable {
			pkt.Time = correcttime
			for i := range inst.time {
				inst.time[i] = correcttime
			}
		} else {
			drop = true
		}
	}
	return
}

func (inst *AVSync) check(i int) (start time.Duration, end time.Duration, correctable bool, correcttime time.Duration) {
	minidx := -1
	maxidx := -1
	for j := range inst.time {
		if minidx == -1 || inst.time[j] < inst.time[minidx] {
			minidx = j
		}
		if maxidx == -1 || inst.time[j] > inst.time[maxidx] {
			maxidx = j
		}
	}
	allthesame := inst.time[minidx] == inst.time[maxidx]

	if i == maxidx {
		if allthesame {
			correctable = true
		} else {
			correctable = false
		}
	} else {
		correctable = true
	}

	start = inst.time[minidx]
	end = start + inst.MaxTimeDiff
	correcttime = start + time.Millisecond*40
	return
}

// Walltime Make packets reading speed as same as walltime, effect like ffmpeg -re option.
type Walltime struct {
	firsttime time.Time
}

// ModifyPacket ModifyPacket
func (inst *Walltime) ModifyPacket(pkt *av.Packet, streams []av.CodecData, videoidx int, audioidx int) (drop bool, err error) {
	if pkt.Idx == 0 {
		if inst.firsttime.IsZero() {
			inst.firsttime = time.Now()
		}
		pkttime := inst.firsttime.Add(pkt.Time)
		delta := pkttime.Sub(time.Now())
		if delta > 0 {
			time.Sleep(delta)
		}
	}
	return
}
