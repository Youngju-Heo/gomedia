package pktque

import (
	"time"
)

/*
pop                                   push

     seg                 seg        seg
  |--------|         |---------|   |---|
     20ms                40ms       5ms
----------------- time -------------------->
headtm                               tailtm
*/

type tlSeg struct {
	tm, dur time.Duration
}

// Timeline struct
type Timeline struct {
	segs   []tlSeg
	headtm time.Duration
}

// Push func
func (inst *Timeline) Push(tm time.Duration, dur time.Duration) {
	if len(inst.segs) > 0 {
		tail := inst.segs[len(inst.segs)-1]
		diff := tm - (tail.tm + tail.dur)
		if diff < 0 {
			tm -= diff
		}
	}
	inst.segs = append(inst.segs, tlSeg{tm, dur})
}

// Pop func
func (inst *Timeline) Pop(dur time.Duration) (tm time.Duration) {
	if len(inst.segs) == 0 {
		return inst.headtm
	}

	tm = inst.segs[0].tm
	for dur > 0 && len(inst.segs) > 0 {
		seg := &inst.segs[0]
		sub := dur
		if seg.dur < sub {
			sub = seg.dur
		}
		seg.dur -= sub
		dur -= sub
		seg.tm += sub
		inst.headtm += sub
		if seg.dur == 0 {
			copy(inst.segs[0:], inst.segs[1:])
			inst.segs = inst.segs[:len(inst.segs)-1]
		}
	}

	return
}
