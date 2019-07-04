package pktque

import (
	"../av"
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
func (inst *Buf) Pop() av.Packet {
	if inst.Count == 0 {
		panic("pktque.Buf: Pop() when count == 0")
	}

	i := int(inst.Head) & (len(inst.pkts) - 1)
	pkt := inst.pkts[i]
	inst.pkts[i] = av.Packet{}
	inst.Size -= len(pkt.Data)
	inst.Head++
	inst.Count--

	return pkt
}

func (inst *Buf) grow() {
	newpkts := make([]av.Packet, len(inst.pkts)*2)
	for i := inst.Head; i.LT(inst.Tail); i++ {
		newpkts[int(i)&(len(newpkts)-1)] = inst.pkts[int(i)&(len(inst.pkts)-1)]
	}
	inst.pkts = newpkts
}

// Push func
func (inst *Buf) Push(pkt av.Packet) {
	if inst.Count == len(inst.pkts) {
		inst.grow()
	}
	inst.pkts[int(inst.Tail)&(len(inst.pkts)-1)] = pkt
	inst.Tail++
	inst.Count++
	inst.Size += len(pkt.Data)
}

// Get packet
func (inst *Buf) Get(pos BufPos) av.Packet {
	return inst.pkts[int(pos)&(len(inst.pkts)-1)]
}

// IsValidPos IsValidPos
func (inst *Buf) IsValidPos(pos BufPos) bool {
	return pos.GE(inst.Head) && pos.LT(inst.Tail)
}

// BufPos type
type BufPos int

// LT check
func (inst BufPos) LT(pos BufPos) bool {
	return inst-pos < 0
}

// GE check
func (inst BufPos) GE(pos BufPos) bool {
	return inst-pos >= 0
}

// GT check
func (inst BufPos) GT(pos BufPos) bool {
	return inst-pos > 0
}
