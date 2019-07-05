package pktque

import (
	"github.com/Youngju-Heo/gomedia/core/media/av"
)

// Buf struct
type Buf struct {
	Head, Tail BufPos
	pkts       []av.Packet
	Size       int
	Count      int
}

// NewBuf NewBuf
func NewBuf() *Buf {
	return &Buf{
		pkts: make([]av.Packet, 64),
	}
}

// Pop Pop
func (instance *Buf) Pop() av.Packet {
	if instance.Count == 0 {
		panic("pktque.Buf: Pop() when count == 0")
	}

	i := int(instance.Head) & (len(instance.pkts) - 1)
	pkt := instance.pkts[i]
	instance.pkts[i] = av.Packet{}
	instance.Size -= len(pkt.Data)
	instance.Head++
	instance.Count--

	return pkt
}

func (instance *Buf) grow() {
	newpkts := make([]av.Packet, len(instance.pkts)*2)
	for i := instance.Head; i.LT(instance.Tail); i++ {
		newpkts[int(i)&(len(newpkts)-1)] = instance.pkts[int(i)&(len(instance.pkts)-1)]
	}
	instance.pkts = newpkts
}

// Push func
func (instance *Buf) Push(pkt av.Packet) {
	if instance.Count == len(instance.pkts) {
		instance.grow()
	}
	instance.pkts[int(instance.Tail)&(len(instance.pkts)-1)] = pkt
	instance.Tail++
	instance.Count++
	instance.Size += len(pkt.Data)
}

// Get packet
func (instance *Buf) Get(pos BufPos) av.Packet {
	return instance.pkts[int(pos)&(len(instance.pkts)-1)]
}

// IsValidPos IsValidPos
func (instance *Buf) IsValidPos(pos BufPos) bool {
	return pos.GE(instance.Head) && pos.LT(instance.Tail)
}

// BufPos type
type BufPos int

// LT check
func (instance BufPos) LT(pos BufPos) bool {
	return instance-pos < 0
}

// GE check
func (instance BufPos) GE(pos BufPos) bool {
	return instance-pos >= 0
}

// GT check
func (instance BufPos) GT(pos BufPos) bool {
	return instance-pos > 0
}
