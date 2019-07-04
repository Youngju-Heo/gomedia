package bits

import (
	"io"
)

// Reader struct
type Reader struct {
	R    io.Reader
	n    int
	bits uint64
}

// ReadBits64 func
func (inst *Reader) ReadBits64(n int) (bits uint64, err error) {
	if inst.n < n {
		var b [8]byte
		var got int
		want := (n - inst.n + 7) / 8
		if got, err = inst.R.Read(b[:want]); err != nil {
			return
		}
		if got < want {
			err = io.EOF
			return
		}
		for i := 0; i < got; i++ {
			inst.bits <<= 8
			inst.bits |= uint64(b[i])
		}
		inst.n += got * 8
	}
	bits = inst.bits >> uint(inst.n-n)
	inst.bits ^= bits << uint(inst.n-n)
	inst.n -= n
	return
}

// ReadBits func
func (inst *Reader) ReadBits(n int) (bits uint, err error) {
	var bits64 uint64
	if bits64, err = inst.ReadBits64(n); err != nil {
		return
	}
	bits = uint(bits64)
	return
}

// Read func
func (inst *Reader) Read(p []byte) (n int, err error) {
	for n < len(p) {
		want := 8
		if len(p)-n < want {
			want = len(p) - n
		}
		var bits uint64
		if bits, err = inst.ReadBits64(want * 8); err != nil {
			break
		}
		for i := 0; i < want; i++ {
			p[n+i] = byte(bits >> uint((want-i-1)*8))
		}
		n += want
	}
	return
}

// Writer func
type Writer struct {
	W    io.Writer
	n    int
	bits uint64
}

// WriteBits64 func
func (inst *Writer) WriteBits64(bits uint64, n int) (err error) {
	if inst.n+n > 64 {
		move := uint(64 - inst.n)
		mask := bits >> move
		inst.bits = (inst.bits << move) | mask
		inst.n = 64
		if err = inst.FlushBits(); err != nil {
			return
		}
		n -= int(move)
		bits ^= (mask << move)
	}
	inst.bits = (inst.bits << uint(n)) | bits
	inst.n += n
	return
}

// WriteBits func
func (inst *Writer) WriteBits(bits uint, n int) (err error) {
	return inst.WriteBits64(uint64(bits), n)
}

// Write func
func (inst *Writer) Write(p []byte) (n int, err error) {
	for n < len(p) {
		if err = inst.WriteBits64(uint64(p[n]), 8); err != nil {
			return
		}
		n++
	}
	return
}

// FlushBits func
func (inst *Writer) FlushBits() (err error) {
	if inst.n > 0 {
		var b [8]byte
		bits := inst.bits
		if inst.n%8 != 0 {
			bits <<= uint(8 - (inst.n % 8))
		}
		want := (inst.n + 7) / 8
		for i := 0; i < want; i++ {
			b[i] = byte(bits >> uint((want-i-1)*8))
		}
		if _, err = inst.W.Write(b[:want]); err != nil {
			return
		}
		inst.n = 0
	}
	return
}
