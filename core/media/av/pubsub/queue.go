// Package pubsub implements publisher-subscribers model used in multi-channel streaming.
package pubsub

import (
	"io"
	"sync"
	"time"

	"../../av"
	"../../av/pktque"
)

//        time
// ----------------->
//
// V-A-V-V-A-V-V-A-V-V
// |                 |
// 0        5        10
// head             tail
// oldest          latest
//

// Queue struct One publisher and multiple subscribers thread-safe packet buffer queue.
type Queue struct {
	buf                      *pktque.Buf
	head, tail               int
	lock                     *sync.RWMutex
	cond                     *sync.Cond
	curgopcount, maxgopcount int
	streams                  []av.CodecData
	videoidx                 int
	closed                   bool
}

// NewQueue func
func NewQueue() *Queue {
	q := &Queue{}
	q.buf = pktque.NewBuf()
	q.maxgopcount = 2
	q.lock = &sync.RWMutex{}
	q.cond = sync.NewCond(q.lock.RLocker())
	q.videoidx = -1
	return q
}

// SetMaxGopCount func
func (instance *Queue) SetMaxGopCount(n int) {
	instance.lock.Lock()
	instance.maxgopcount = n
	instance.lock.Unlock()
	return
}

// WriteHeader func
func (instance *Queue) WriteHeader(streams []av.CodecData) error {
	instance.lock.Lock()

	instance.streams = streams
	for i, stream := range streams {
		if stream.Type().IsVideo() {
			instance.videoidx = i
		}
	}
	instance.cond.Broadcast()

	instance.lock.Unlock()

	return nil
}

// WriteTrailer func
func (instance *Queue) WriteTrailer() error {
	return nil
}

// Close After Close() called, all QueueCursor's ReadPacket will return io.EOF.
func (instance *Queue) Close() (err error) {
	instance.lock.Lock()

	instance.closed = true
	instance.cond.Broadcast()

	instance.lock.Unlock()
	return
}

// WritePacket Put packet into buffer, old packets will be discared.
func (instance *Queue) WritePacket(pkt av.Packet) (err error) {
	instance.lock.Lock()

	instance.buf.Push(pkt)
	if pkt.Idx == int8(instance.videoidx) && pkt.IsKeyFrame {
		instance.curgopcount++
	}

	for instance.curgopcount >= instance.maxgopcount && instance.buf.Count > 1 {
		pkt := instance.buf.Pop()
		if pkt.Idx == int8(instance.videoidx) && pkt.IsKeyFrame {
			instance.curgopcount--
		}
		if instance.curgopcount < instance.maxgopcount {
			break
		}
	}
	//println("shrink", instance.curgopcount, instance.maxgopcount, instance.buf.Head, instance.buf.Tail, "count", instance.buf.Count, "size", instance.buf.Size)

	instance.cond.Broadcast()

	instance.lock.Unlock()
	return
}

// QueueCursor struct
type QueueCursor struct {
	que    *Queue
	pos    pktque.BufPos
	gotpos bool
	init   func(buf *pktque.Buf, videoidx int) pktque.BufPos
}

func (instance *Queue) newCursor() *QueueCursor {
	return &QueueCursor{
		que: instance,
	}
}

// Latest Create cursor position at latest packet.
func (instance *Queue) Latest() *QueueCursor {
	cursor := instance.newCursor()
	cursor.init = func(buf *pktque.Buf, videoidx int) pktque.BufPos {
		return buf.Tail
	}
	return cursor
}

// Oldest Create cursor position at oldest buffered packet.
func (instance *Queue) Oldest() *QueueCursor {
	cursor := instance.newCursor()
	cursor.init = func(buf *pktque.Buf, videoidx int) pktque.BufPos {
		return buf.Head
	}
	return cursor
}

// DelayedTime Create cursor position at specific time in buffered packets.
func (instance *Queue) DelayedTime(dur time.Duration) *QueueCursor {
	cursor := instance.newCursor()
	cursor.init = func(buf *pktque.Buf, videoidx int) pktque.BufPos {
		i := buf.Tail - 1
		if buf.IsValidPos(i) {
			end := buf.Get(i)
			for buf.IsValidPos(i) {
				if end.Time-buf.Get(i).Time > dur {
					break
				}
				i--
			}
		}
		return i
	}
	return cursor
}

// DelayedGopCount Create cursor position at specific delayed GOP count in buffered packets.
func (instance *Queue) DelayedGopCount(n int) *QueueCursor {
	cursor := instance.newCursor()
	cursor.init = func(buf *pktque.Buf, videoidx int) pktque.BufPos {
		i := buf.Tail - 1
		if videoidx != -1 {
			for gop := 0; buf.IsValidPos(i) && gop < n; i-- {
				pkt := buf.Get(i)
				if pkt.Idx == int8(instance.videoidx) && pkt.IsKeyFrame {
					gop++
				}
			}
		}
		return i
	}
	return cursor
}

// Streams func
func (instance *QueueCursor) Streams() (streams []av.CodecData, err error) {
	instance.que.cond.L.Lock()
	for instance.que.streams == nil && !instance.que.closed {
		instance.que.cond.Wait()
	}
	if instance.que.streams != nil {
		streams = instance.que.streams
	} else {
		err = io.EOF
	}
	instance.que.cond.L.Unlock()
	return
}

// ReadPacket will not consume packets in Queue, it's just a cursor.
func (instance *QueueCursor) ReadPacket() (pkt av.Packet, err error) {
	instance.que.cond.L.Lock()
	buf := instance.que.buf
	if !instance.gotpos {
		instance.pos = instance.init(buf, instance.que.videoidx)
		instance.gotpos = true
	}
	for {
		if instance.pos.LT(buf.Head) {
			instance.pos = buf.Head
		} else if instance.pos.GT(buf.Tail) {
			instance.pos = buf.Tail
		}
		if buf.IsValidPos(instance.pos) {
			pkt = buf.Get(instance.pos)
			instance.pos++
			break
		}
		if instance.que.closed {
			err = io.EOF
			break
		}
		instance.que.cond.Wait()
	}
	instance.que.cond.L.Unlock()
	return
}
