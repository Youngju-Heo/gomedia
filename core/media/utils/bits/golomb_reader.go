package bits

import (
	"io"
)

// GolombBitReader type
type GolombBitReader struct {
	R    io.Reader
	buf  [1]byte
	left byte
}

// ReadBit func
func (inst *GolombBitReader) ReadBit() (res uint, err error) {
	if inst.left == 0 {
		if _, err = inst.R.Read(inst.buf[:]); err != nil {
			return
		}
		inst.left = 8
	}
	inst.left--
	res = uint(inst.buf[0]>>inst.left) & 1
	return
}

// ReadBits func
func (inst *GolombBitReader) ReadBits(n int) (res uint, err error) {
	for i := 0; i < n; i++ {
		var bit uint
		if bit, err = inst.ReadBit(); err != nil {
			return
		}
		res |= bit << uint(n-i-1)
	}
	return
}

// ReadExponentialGolombCode func
func (inst *GolombBitReader) ReadExponentialGolombCode() (res uint, err error) {
	i := 0
	for {
		var bit uint
		if bit, err = inst.ReadBit(); err != nil {
			return
		}
		if !(bit == 0 && i < 32) {
			break
		}
		i++
	}
	if res, err = inst.ReadBits(i); err != nil {
		return
	}
	res += (1 << uint(i)) - 1
	return
}

// ReadSE func
func (inst *GolombBitReader) ReadSE() (res uint, err error) {
	if res, err = inst.ReadExponentialGolombCode(); err != nil {
		return
	}
	if res&0x01 != 0 {
		res = (res + 1) / 2
	} else {
		res = -res / 2
	}
	return
}
