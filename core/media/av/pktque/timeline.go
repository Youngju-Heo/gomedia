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
func (instance *Timeline) Push(tm time.Duration, dur time.Duration) {
	if len(instance.segs) > 0 {
		tail := instance.segs[len(instance.segs)-1]
		diff := tm - (tail.tm + tail.dur)
		if diff < 0 {
			tm -= diff
		}
	}
	instance.segs = append(instance.segs, tlSeg{tm, dur})
}

// Pop func
func (instance *Timeline) Pop(dur time.Duration) (tm time.Duration) {
	if len(instance.segs) == 0 {
		return instance.headtm
	}

	tm = instance.segs[0].tm
	for dur > 0 && len(instance.segs) > 0 {
		seg := &instance.segs[0]
		sub := dur
		if seg.dur < sub {
			sub = seg.dur
		}
		seg.dur -= sub
		dur -= sub
		seg.tm += sub
		instance.headtm += sub
		if seg.dur == 0 {
			copy(instance.segs[0:], instance.segs[1:])
			instance.segs = instance.segs[:len(instance.segs)-1]
		}
	}

	return
}
