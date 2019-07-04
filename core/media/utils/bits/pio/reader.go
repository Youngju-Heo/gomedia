package pio

// U8 func
func U8(b []byte) (i uint8) {
	return b[0]
}

// U16BE func
func U16BE(b []byte) (i uint16) {
	i = uint16(b[0])
	i <<= 8
	i |= uint16(b[1])
	return
}

// I16BE func
func I16BE(b []byte) (i int16) {
	i = int16(b[0])
	i <<= 8
	i |= int16(b[1])
	return
}

// I24BE func
func I24BE(b []byte) (i int32) {
	i = int32(int8(b[0]))
	i <<= 8
	i |= int32(b[1])
	i <<= 8
	i |= int32(b[2])
	return
}

// U24BE func
func U24BE(b []byte) (i uint32) {
	i = uint32(b[0])
	i <<= 8
	i |= uint32(b[1])
	i <<= 8
	i |= uint32(b[2])
	return
}

// I32BE func
func I32BE(b []byte) (i int32) {
	i = int32(int8(b[0]))
	i <<= 8
	i |= int32(b[1])
	i <<= 8
	i |= int32(b[2])
	i <<= 8
	i |= int32(b[3])
	return
}

// U32LE func
func U32LE(b []byte) (i uint32) {
	i = uint32(b[3])
	i <<= 8
	i |= uint32(b[2])
	i <<= 8
	i |= uint32(b[1])
	i <<= 8
	i |= uint32(b[0])
	return
}

// U32BE func
func U32BE(b []byte) (i uint32) {
	i = uint32(b[0])
	i <<= 8
	i |= uint32(b[1])
	i <<= 8
	i |= uint32(b[2])
	i <<= 8
	i |= uint32(b[3])
	return
}

// U40BE func
func U40BE(b []byte) (i uint64) {
	i = uint64(b[0])
	i <<= 8
	i |= uint64(b[1])
	i <<= 8
	i |= uint64(b[2])
	i <<= 8
	i |= uint64(b[3])
	i <<= 8
	i |= uint64(b[4])
	return
}

// U64BE func
func U64BE(b []byte) (i uint64) {
	i = uint64(b[0])
	i <<= 8
	i |= uint64(b[1])
	i <<= 8
	i |= uint64(b[2])
	i <<= 8
	i |= uint64(b[3])
	i <<= 8
	i |= uint64(b[4])
	i <<= 8
	i |= uint64(b[5])
	i <<= 8
	i |= uint64(b[6])
	i <<= 8
	i |= uint64(b[7])
	return
}

// I64BE func
func I64BE(b []byte) (i int64) {
	i = int64(int8(b[0]))
	i <<= 8
	i |= int64(b[1])
	i <<= 8
	i |= int64(b[2])
	i <<= 8
	i |= int64(b[3])
	i <<= 8
	i |= int64(b[4])
	i <<= 8
	i |= int64(b[5])
	i <<= 8
	i |= int64(b[6])
	i <<= 8
	i |= int64(b[7])
	return
}
