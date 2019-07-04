package bufio

import (
	"io"
)

// Reader struct
type Reader struct {
	buf [][]byte
	R   io.ReadSeeker
}

// NewReaderSize func
func NewReaderSize(r io.ReadSeeker, size int) *Reader {
	buf := make([]byte, size*2)
	return &Reader{
		R:   r,
		buf: [][]byte{buf[0:size], buf[size:]},
	}
}

// ReadAt func
func (inst *Reader) ReadAt(b []byte, off int64) (n int, err error) {
	return
}
