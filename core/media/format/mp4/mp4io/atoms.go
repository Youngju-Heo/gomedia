package mp4io

import "github.com/Youngju-Heo/gomedia/core/media/utils/bits/pio"
import "time"

// MOOF const
const MOOF = Tag(0x6d6f6f66)

// Tag const
func (inst MovieFrag) Tag() Tag {
	return MOOF
}

// HDLR const
const HDLR = Tag(0x68646c72)

// Tag const
func (inst HandlerRefer) Tag() Tag {
	return HDLR
}

// AVC1 const
const AVC1 = Tag(0x61766331)

// Tag func
func (inst AVC1Desc) Tag() Tag {
	return AVC1
}

// URL const
const URL = Tag(0x75726c20)

// Tag func
func (inst DataReferURL) Tag() Tag {
	return URL
}

// TREX const
const TREX = Tag(0x74726578)

// Tag func
func (inst TrackExtend) Tag() Tag {
	return TREX
}

// ESDS const
const ESDS = Tag(0x65736473)

// Tag func
func (inst ElemStreamDesc) Tag() Tag {
	return ESDS
}

// MDHD const
const MDHD = Tag(0x6d646864)

// Tag func
func (inst MediaHeader) Tag() Tag {
	return MDHD
}

// STTS const
const STTS = Tag(0x73747473)

// Tag func
func (inst TimeToSample) Tag() Tag {
	return STTS
}

// STSS const
const STSS = Tag(0x73747373)

// Tag func
func (inst SyncSample) Tag() Tag {
	return STSS
}

// MFHD const
const MFHD = Tag(0x6d666864)

// Tag func
func (inst MovieFragHeader) Tag() Tag {
	return MFHD
}

// MVHD const
const MVHD = Tag(0x6d766864)

// Tag func
func (inst MovieHeader) Tag() Tag {
	return MVHD
}

// MINF const
const MINF = Tag(0x6d696e66)

// Tag func
func (inst MediaInfo) Tag() Tag {
	return MINF
}

// MOOV const
const MOOV = Tag(0x6d6f6f76)

//Tag func
func (inst Movie) Tag() Tag {
	return MOOV
}

//MVEX const
const MVEX = Tag(0x6d766578)

//Tag func
func (inst MovieExtend) Tag() Tag {
	return MVEX
}

// STSD const
const STSD = Tag(0x73747364)

//Tag func
func (inst SampleDesc) Tag() Tag {
	return STSD
}

// MP4A const
const MP4A = Tag(0x6d703461)

//Tag func
func (inst MP4ADesc) Tag() Tag {
	return MP4A
}

// CTTS const
const CTTS = Tag(0x63747473)

//Tag func
func (inst CompositionOffset) Tag() Tag {
	return CTTS
}

// STCO const
const STCO = Tag(0x7374636f)

//Tag func
func (inst ChunkOffset) Tag() Tag {
	return STCO
}

// TRUN const
const TRUN = Tag(0x7472756e)

//Tag func
func (inst TrackFragRun) Tag() Tag {
	return TRUN
}

// TRAK const
const TRAK = Tag(0x7472616b)

//Tag func
func (inst Track) Tag() Tag {
	return TRAK
}

// MDIA const
const MDIA = Tag(0x6d646961)

//Tag func
func (inst Media) Tag() Tag {
	return MDIA
}

// STSC const
const STSC = Tag(0x73747363)

//Tag func
func (inst SampleToChunk) Tag() Tag {
	return STSC
}

// VMHD const
const VMHD = Tag(0x766d6864)

//Tag func
func (inst VideoMediaInfo) Tag() Tag {
	return VMHD
}

// STBL const
const STBL = Tag(0x7374626c)

//Tag func
func (inst SampleTable) Tag() Tag {
	return STBL
}

// AVCC const
const AVCC = Tag(0x61766343)

//Tag func
func (inst AVC1Conf) Tag() Tag {
	return AVCC
}

// TFDT const
const TFDT = Tag(0x74666474)

//Tag func
func (inst TrackFragDecodeTime) Tag() Tag {
	return TFDT
}

// DINF const
const DINF = Tag(0x64696e66)

//Tag func
func (inst DataInfo) Tag() Tag {
	return DINF
}

// DREF const
const DREF = Tag(0x64726566)

//Tag func
func (inst DataRefer) Tag() Tag {
	return DREF
}

// TRAF const
const TRAF = Tag(0x74726166)

//Tag func
func (inst TrackFrag) Tag() Tag {
	return TRAF
}

// STSZ const
const STSZ = Tag(0x7374737a)

//Tag func
func (inst SampleSize) Tag() Tag {
	return STSZ
}

// TFHD const
const TFHD = Tag(0x74666864)

//Tag func
func (inst TrackFragHeader) Tag() Tag {
	return TFHD
}

// TKHD const
const TKHD = Tag(0x746b6864)

//Tag func
func (inst TrackHeader) Tag() Tag {
	return TKHD
}

// SMHD const
const SMHD = Tag(0x736d6864)

//Tag func
func (inst SoundMediaInfo) Tag() Tag {
	return SMHD
}

// MDAT const
const MDAT = Tag(0x6d646174)

// Movie struct
type Movie struct {
	Header      *MovieHeader
	MovieExtend *MovieExtend
	Tracks      []*Track
	Unknowns    []Atom
	AtomPos
}

// Marshal func
func (inst Movie) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(MOOV))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst Movie) marshal(b []byte) (n int) {
	if inst.Header != nil {
		n += inst.Header.Marshal(b[n:])
	}
	if inst.MovieExtend != nil {
		n += inst.MovieExtend.Marshal(b[n:])
	}
	for _, atom := range inst.Tracks {
		n += atom.Marshal(b[n:])
	}
	for _, atom := range inst.Unknowns {
		n += atom.Marshal(b[n:])
	}
	return
}

// Len func
func (inst Movie) Len() (n int) {
	n += 8
	if inst.Header != nil {
		n += inst.Header.Len()
	}
	if inst.MovieExtend != nil {
		n += inst.MovieExtend.Len()
	}
	for _, atom := range inst.Tracks {
		n += atom.Len()
	}
	for _, atom := range inst.Unknowns {
		n += atom.Len()
	}
	return
}

// Unmarshal func
func (inst *Movie) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	for n+8 < len(b) {
		tag := Tag(pio.U32BE(b[n+4:]))
		size := int(pio.U32BE(b[n:]))
		if len(b) < n+size {
			err = parseErr("TagSizeInvalid", n+offset, err)
			return
		}
		switch tag {
		case MVHD:
			{
				atom := &MovieHeader{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("mvhd", n+offset, err)
					return
				}
				inst.Header = atom
			}
		case MVEX:
			{
				atom := &MovieExtend{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("mvex", n+offset, err)
					return
				}
				inst.MovieExtend = atom
			}
		case TRAK:
			{
				atom := &Track{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("trak", n+offset, err)
					return
				}
				inst.Tracks = append(inst.Tracks, atom)
			}
		default:
			{
				atom := &Dummy{TagItem: tag, Data: b[n : n+size]}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("", n+offset, err)
					return
				}
				inst.Unknowns = append(inst.Unknowns, atom)
			}
		}
		n += size
	}
	return
}

// Children func
func (inst Movie) Children() (r []Atom) {
	if inst.Header != nil {
		r = append(r, inst.Header)
	}
	if inst.MovieExtend != nil {
		r = append(r, inst.MovieExtend)
	}
	for _, atom := range inst.Tracks {
		r = append(r, atom)
	}
	r = append(r, inst.Unknowns...)
	return
}

// MovieHeader struct
type MovieHeader struct {
	Version           uint8
	Flags             uint32
	CreateTime        time.Time
	ModifyTime        time.Time
	TimeScale         int32
	Duration          int32
	PreferredRate     float64
	PreferredVolume   float64
	Matrix            [9]int32
	PreviewTime       time.Time
	PreviewDuration   time.Time
	PosterTime        time.Time
	SelectionTime     time.Time
	SelectionDuration time.Time
	CurrentTime       time.Time
	NextTrackID       int32
	AtomPos
}

// Marshal func
func (inst MovieHeader) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(MVHD))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst MovieHeader) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], inst.Version)
	n++
	pio.PutU24BE(b[n:], inst.Flags)
	n += 3
	PutTime32(b[n:], inst.CreateTime)
	n += 4
	PutTime32(b[n:], inst.ModifyTime)
	n += 4
	pio.PutI32BE(b[n:], inst.TimeScale)
	n += 4
	pio.PutI32BE(b[n:], inst.Duration)
	n += 4
	PutFixed32(b[n:], inst.PreferredRate)
	n += 4
	PutFixed16(b[n:], inst.PreferredVolume)
	n += 2
	n += 10
	for _, entry := range inst.Matrix {
		pio.PutI32BE(b[n:], entry)
		n += 4
	}
	PutTime32(b[n:], inst.PreviewTime)
	n += 4
	PutTime32(b[n:], inst.PreviewDuration)
	n += 4
	PutTime32(b[n:], inst.PosterTime)
	n += 4
	PutTime32(b[n:], inst.SelectionTime)
	n += 4
	PutTime32(b[n:], inst.SelectionDuration)
	n += 4
	PutTime32(b[n:], inst.CurrentTime)
	n += 4
	pio.PutI32BE(b[n:], inst.NextTrackID)
	n += 4
	return
}

// Len func
func (inst MovieHeader) Len() (n int) {
	n += 8
	n++
	n += 3
	n += 4
	n += 4
	n += 4
	n += 4
	n += 4
	n += 2
	n += 10
	n += 4 * len(inst.Matrix[:])
	n += 4
	n += 4
	n += 4
	n += 4
	n += 4
	n += 4
	n += 4
	return
}

// Unmarshal func
func (inst *MovieHeader) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	if len(b) < n+1 {
		err = parseErr("Version", n+offset, err)
		return
	}
	inst.Version = pio.U8(b[n:])
	n++
	if len(b) < n+3 {
		err = parseErr("Flags", n+offset, err)
		return
	}
	inst.Flags = pio.U24BE(b[n:])
	n += 3
	if len(b) < n+4 {
		err = parseErr("CreateTime", n+offset, err)
		return
	}
	inst.CreateTime = GetTime32(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("ModifyTime", n+offset, err)
		return
	}
	inst.ModifyTime = GetTime32(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("TimeScale", n+offset, err)
		return
	}
	inst.TimeScale = pio.I32BE(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("Duration", n+offset, err)
		return
	}
	inst.Duration = pio.I32BE(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("PreferredRate", n+offset, err)
		return
	}
	inst.PreferredRate = GetFixed32(b[n:])
	n += 4
	if len(b) < n+2 {
		err = parseErr("PreferredVolume", n+offset, err)
		return
	}
	inst.PreferredVolume = GetFixed16(b[n:])
	n += 2
	n += 10
	if len(b) < n+4*len(inst.Matrix) {
		err = parseErr("Matrix", n+offset, err)
		return
	}
	for i := range inst.Matrix {
		inst.Matrix[i] = pio.I32BE(b[n:])
		n += 4
	}
	if len(b) < n+4 {
		err = parseErr("PreviewTime", n+offset, err)
		return
	}
	inst.PreviewTime = GetTime32(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("PreviewDuration", n+offset, err)
		return
	}
	inst.PreviewDuration = GetTime32(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("PosterTime", n+offset, err)
		return
	}
	inst.PosterTime = GetTime32(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("SelectionTime", n+offset, err)
		return
	}
	inst.SelectionTime = GetTime32(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("SelectionDuration", n+offset, err)
		return
	}
	inst.SelectionDuration = GetTime32(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("CurrentTime", n+offset, err)
		return
	}
	inst.CurrentTime = GetTime32(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("NextTrackID", n+offset, err)
		return
	}
	inst.NextTrackID = pio.I32BE(b[n:])
	n += 4
	return
}

// Children func
func (inst MovieHeader) Children() (r []Atom) {
	return
}

// Track struct
type Track struct {
	Header   *TrackHeader
	Media    *Media
	Unknowns []Atom
	AtomPos
}

// Marshal func
func (inst Track) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(TRAK))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst Track) marshal(b []byte) (n int) {
	if inst.Header != nil {
		n += inst.Header.Marshal(b[n:])
	}
	if inst.Media != nil {
		n += inst.Media.Marshal(b[n:])
	}
	for _, atom := range inst.Unknowns {
		n += atom.Marshal(b[n:])
	}
	return
}

// Len func
func (inst Track) Len() (n int) {
	n += 8
	if inst.Header != nil {
		n += inst.Header.Len()
	}
	if inst.Media != nil {
		n += inst.Media.Len()
	}
	for _, atom := range inst.Unknowns {
		n += atom.Len()
	}
	return
}

// Unmarshal func
func (inst *Track) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	for n+8 < len(b) {
		tag := Tag(pio.U32BE(b[n+4:]))
		size := int(pio.U32BE(b[n:]))
		if len(b) < n+size {
			err = parseErr("TagSizeInvalid", n+offset, err)
			return
		}
		switch tag {
		case TKHD:
			{
				atom := &TrackHeader{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("tkhd", n+offset, err)
					return
				}
				inst.Header = atom
			}
		case MDIA:
			{
				atom := &Media{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("mdia", n+offset, err)
					return
				}
				inst.Media = atom
			}
		default:
			{
				atom := &Dummy{TagItem: tag, Data: b[n : n+size]}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("", n+offset, err)
					return
				}
				inst.Unknowns = append(inst.Unknowns, atom)
			}
		}
		n += size
	}
	return
}

// Children func
func (inst Track) Children() (r []Atom) {
	if inst.Header != nil {
		r = append(r, inst.Header)
	}
	if inst.Media != nil {
		r = append(r, inst.Media)
	}
	r = append(r, inst.Unknowns...)
	return
}

// TrackHeader struct
type TrackHeader struct {
	Version        uint8
	Flags          uint32
	CreateTime     time.Time
	ModifyTime     time.Time
	TrackID        int32
	Duration       int32
	Layer          int16
	AlternateGroup int16
	Volume         float64
	Matrix         [9]int32
	TrackWidth     float64
	TrackHeight    float64
	AtomPos
}

// Marshal func
func (inst TrackHeader) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(TKHD))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst TrackHeader) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], inst.Version)
	n++
	pio.PutU24BE(b[n:], inst.Flags)
	n += 3
	PutTime32(b[n:], inst.CreateTime)
	n += 4
	PutTime32(b[n:], inst.ModifyTime)
	n += 4
	pio.PutI32BE(b[n:], inst.TrackID)
	n += 4
	n += 4
	pio.PutI32BE(b[n:], inst.Duration)
	n += 4
	n += 8
	pio.PutI16BE(b[n:], inst.Layer)
	n += 2
	pio.PutI16BE(b[n:], inst.AlternateGroup)
	n += 2
	PutFixed16(b[n:], inst.Volume)
	n += 2
	n += 2
	for _, entry := range inst.Matrix {
		pio.PutI32BE(b[n:], entry)
		n += 4
	}
	PutFixed32(b[n:], inst.TrackWidth)
	n += 4
	PutFixed32(b[n:], inst.TrackHeight)
	n += 4
	return
}

// Len func
func (inst TrackHeader) Len() (n int) {
	n += 8
	n++
	n += 3
	n += 4
	n += 4
	n += 4
	n += 4
	n += 4
	n += 8
	n += 2
	n += 2
	n += 2
	n += 2
	n += 4 * len(inst.Matrix[:])
	n += 4
	n += 4
	return
}

// Unmarshal func
func (inst *TrackHeader) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	if len(b) < n+1 {
		err = parseErr("Version", n+offset, err)
		return
	}
	inst.Version = pio.U8(b[n:])
	n++
	if len(b) < n+3 {
		err = parseErr("Flags", n+offset, err)
		return
	}
	inst.Flags = pio.U24BE(b[n:])
	n += 3
	if len(b) < n+4 {
		err = parseErr("CreateTime", n+offset, err)
		return
	}
	inst.CreateTime = GetTime32(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("ModifyTime", n+offset, err)
		return
	}
	inst.ModifyTime = GetTime32(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("TrackID", n+offset, err)
		return
	}
	inst.TrackID = pio.I32BE(b[n:])
	n += 4
	n += 4
	if len(b) < n+4 {
		err = parseErr("Duration", n+offset, err)
		return
	}
	inst.Duration = pio.I32BE(b[n:])
	n += 4
	n += 8
	if len(b) < n+2 {
		err = parseErr("Layer", n+offset, err)
		return
	}
	inst.Layer = pio.I16BE(b[n:])
	n += 2
	if len(b) < n+2 {
		err = parseErr("AlternateGroup", n+offset, err)
		return
	}
	inst.AlternateGroup = pio.I16BE(b[n:])
	n += 2
	if len(b) < n+2 {
		err = parseErr("Volume", n+offset, err)
		return
	}
	inst.Volume = GetFixed16(b[n:])
	n += 2
	n += 2
	if len(b) < n+4*len(inst.Matrix) {
		err = parseErr("Matrix", n+offset, err)
		return
	}
	for i := range inst.Matrix {
		inst.Matrix[i] = pio.I32BE(b[n:])
		n += 4
	}
	if len(b) < n+4 {
		err = parseErr("TrackWidth", n+offset, err)
		return
	}
	inst.TrackWidth = GetFixed32(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("TrackHeight", n+offset, err)
		return
	}
	inst.TrackHeight = GetFixed32(b[n:])
	n += 4
	return
}

// Children func
func (inst TrackHeader) Children() (r []Atom) {
	return
}

// HandlerRefer struct
type HandlerRefer struct {
	Version uint8
	Flags   uint32
	Type    [4]byte
	SubType [4]byte
	Name    []byte
	AtomPos
}

// Marshal func
func (inst HandlerRefer) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(HDLR))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst HandlerRefer) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], inst.Version)
	n++
	pio.PutU24BE(b[n:], inst.Flags)
	n += 3
	copy(b[n:], inst.Type[:])
	n += len(inst.Type[:])
	copy(b[n:], inst.SubType[:])
	n += len(inst.SubType[:])
	copy(b[n:], inst.Name[:])
	n += len(inst.Name[:])
	return
}

// Len func
func (inst HandlerRefer) Len() (n int) {
	n += 8
	n++
	n += 3
	n += len(inst.Type[:])
	n += len(inst.SubType[:])
	n += len(inst.Name[:])
	return
}

// Unmarshal func
func (inst *HandlerRefer) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	if len(b) < n+1 {
		err = parseErr("Version", n+offset, err)
		return
	}
	inst.Version = pio.U8(b[n:])
	n++
	if len(b) < n+3 {
		err = parseErr("Flags", n+offset, err)
		return
	}
	inst.Flags = pio.U24BE(b[n:])
	n += 3
	if len(b) < n+len(inst.Type) {
		err = parseErr("Type", n+offset, err)
		return
	}
	copy(inst.Type[:], b[n:])
	n += len(inst.Type)
	if len(b) < n+len(inst.SubType) {
		err = parseErr("SubType", n+offset, err)
		return
	}
	copy(inst.SubType[:], b[n:])
	n += len(inst.SubType)
	inst.Name = b[n:]
	n += len(b[n:])
	return
}

// Children func
func (inst HandlerRefer) Children() (r []Atom) {
	return
}

// Media struct
type Media struct {
	Header   *MediaHeader
	Handler  *HandlerRefer
	Info     *MediaInfo
	Unknowns []Atom
	AtomPos
}

// Marshal func
func (inst Media) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(MDIA))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}

func (inst Media) marshal(b []byte) (n int) {
	if inst.Header != nil {
		n += inst.Header.Marshal(b[n:])
	}
	if inst.Handler != nil {
		n += inst.Handler.Marshal(b[n:])
	}
	if inst.Info != nil {
		n += inst.Info.Marshal(b[n:])
	}
	for _, atom := range inst.Unknowns {
		n += atom.Marshal(b[n:])
	}
	return
}

// Len func
func (inst Media) Len() (n int) {
	n += 8
	if inst.Header != nil {
		n += inst.Header.Len()
	}
	if inst.Handler != nil {
		n += inst.Handler.Len()
	}
	if inst.Info != nil {
		n += inst.Info.Len()
	}
	for _, atom := range inst.Unknowns {
		n += atom.Len()
	}
	return
}

// Unmarshal func
func (inst *Media) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	for n+8 < len(b) {
		tag := Tag(pio.U32BE(b[n+4:]))
		size := int(pio.U32BE(b[n:]))
		if len(b) < n+size {
			err = parseErr("TagSizeInvalid", n+offset, err)
			return
		}
		switch tag {
		case MDHD:
			{
				atom := &MediaHeader{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("mdhd", n+offset, err)
					return
				}
				inst.Header = atom
			}
		case HDLR:
			{
				atom := &HandlerRefer{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("hdlr", n+offset, err)
					return
				}
				inst.Handler = atom
			}
		case MINF:
			{
				atom := &MediaInfo{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("minf", n+offset, err)
					return
				}
				inst.Info = atom
			}
		default:
			{
				atom := &Dummy{TagItem: tag, Data: b[n : n+size]}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("", n+offset, err)
					return
				}
				inst.Unknowns = append(inst.Unknowns, atom)
			}
		}
		n += size
	}
	return
}

// Children func
func (inst Media) Children() (r []Atom) {
	if inst.Header != nil {
		r = append(r, inst.Header)
	}
	if inst.Handler != nil {
		r = append(r, inst.Handler)
	}
	if inst.Info != nil {
		r = append(r, inst.Info)
	}
	r = append(r, inst.Unknowns...)
	return
}

// MediaHeader struct
type MediaHeader struct {
	Version    uint8
	Flags      uint32
	CreateTime time.Time
	ModifyTime time.Time
	TimeScale  int32
	Duration   int32
	Language   int16
	Quality    int16
	AtomPos
}

// Marshal func
func (inst MediaHeader) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(MDHD))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst MediaHeader) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], inst.Version)
	n++
	pio.PutU24BE(b[n:], inst.Flags)
	n += 3
	PutTime32(b[n:], inst.CreateTime)
	n += 4
	PutTime32(b[n:], inst.ModifyTime)
	n += 4
	pio.PutI32BE(b[n:], inst.TimeScale)
	n += 4
	pio.PutI32BE(b[n:], inst.Duration)
	n += 4
	pio.PutI16BE(b[n:], inst.Language)
	n += 2
	pio.PutI16BE(b[n:], inst.Quality)
	n += 2
	return
}

// Len func
func (inst MediaHeader) Len() (n int) {
	n += 8
	n++
	n += 3
	n += 4
	n += 4
	n += 4
	n += 4
	n += 2
	n += 2
	return
}

// Unmarshal func
func (inst *MediaHeader) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	if len(b) < n+1 {
		err = parseErr("Version", n+offset, err)
		return
	}
	inst.Version = pio.U8(b[n:])
	n++
	if len(b) < n+3 {
		err = parseErr("Flags", n+offset, err)
		return
	}
	inst.Flags = pio.U24BE(b[n:])
	n += 3
	if len(b) < n+4 {
		err = parseErr("CreateTime", n+offset, err)
		return
	}
	inst.CreateTime = GetTime32(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("ModifyTime", n+offset, err)
		return
	}
	inst.ModifyTime = GetTime32(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("TimeScale", n+offset, err)
		return
	}
	inst.TimeScale = pio.I32BE(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("Duration", n+offset, err)
		return
	}
	inst.Duration = pio.I32BE(b[n:])
	n += 4
	if len(b) < n+2 {
		err = parseErr("Language", n+offset, err)
		return
	}
	inst.Language = pio.I16BE(b[n:])
	n += 2
	if len(b) < n+2 {
		err = parseErr("Quality", n+offset, err)
		return
	}
	inst.Quality = pio.I16BE(b[n:])
	n += 2
	return
}

// Children func
func (inst MediaHeader) Children() (r []Atom) {
	return
}

// MediaInfo struct
type MediaInfo struct {
	Sound    *SoundMediaInfo
	Video    *VideoMediaInfo
	Data     *DataInfo
	Sample   *SampleTable
	Unknowns []Atom
	AtomPos
}

// Marshal func
func (inst MediaInfo) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(MINF))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst MediaInfo) marshal(b []byte) (n int) {
	if inst.Sound != nil {
		n += inst.Sound.Marshal(b[n:])
	}
	if inst.Video != nil {
		n += inst.Video.Marshal(b[n:])
	}
	if inst.Data != nil {
		n += inst.Data.Marshal(b[n:])
	}
	if inst.Sample != nil {
		n += inst.Sample.Marshal(b[n:])
	}
	for _, atom := range inst.Unknowns {
		n += atom.Marshal(b[n:])
	}
	return
}

// Len func
func (inst MediaInfo) Len() (n int) {
	n += 8
	if inst.Sound != nil {
		n += inst.Sound.Len()
	}
	if inst.Video != nil {
		n += inst.Video.Len()
	}
	if inst.Data != nil {
		n += inst.Data.Len()
	}
	if inst.Sample != nil {
		n += inst.Sample.Len()
	}
	for _, atom := range inst.Unknowns {
		n += atom.Len()
	}
	return
}

// Unmarshal func
func (inst *MediaInfo) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	for n+8 < len(b) {
		tag := Tag(pio.U32BE(b[n+4:]))
		size := int(pio.U32BE(b[n:]))
		if len(b) < n+size {
			err = parseErr("TagSizeInvalid", n+offset, err)
			return
		}
		switch tag {
		case SMHD:
			{
				atom := &SoundMediaInfo{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("smhd", n+offset, err)
					return
				}
				inst.Sound = atom
			}
		case VMHD:
			{
				atom := &VideoMediaInfo{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("vmhd", n+offset, err)
					return
				}
				inst.Video = atom
			}
		case DINF:
			{
				atom := &DataInfo{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("dinf", n+offset, err)
					return
				}
				inst.Data = atom
			}
		case STBL:
			{
				atom := &SampleTable{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("stbl", n+offset, err)
					return
				}
				inst.Sample = atom
			}
		default:
			{
				atom := &Dummy{TagItem: tag, Data: b[n : n+size]}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("", n+offset, err)
					return
				}
				inst.Unknowns = append(inst.Unknowns, atom)
			}
		}
		n += size
	}
	return
}

// Children func
func (inst MediaInfo) Children() (r []Atom) {
	if inst.Sound != nil {
		r = append(r, inst.Sound)
	}
	if inst.Video != nil {
		r = append(r, inst.Video)
	}
	if inst.Data != nil {
		r = append(r, inst.Data)
	}
	if inst.Sample != nil {
		r = append(r, inst.Sample)
	}
	r = append(r, inst.Unknowns...)
	return
}

// DataInfo struct
type DataInfo struct {
	Refer    *DataRefer
	Unknowns []Atom
	AtomPos
}

// Marshal func
func (inst DataInfo) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(DINF))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst DataInfo) marshal(b []byte) (n int) {
	if inst.Refer != nil {
		n += inst.Refer.Marshal(b[n:])
	}
	for _, atom := range inst.Unknowns {
		n += atom.Marshal(b[n:])
	}
	return
}

// Len func
func (inst DataInfo) Len() (n int) {
	n += 8
	if inst.Refer != nil {
		n += inst.Refer.Len()
	}
	for _, atom := range inst.Unknowns {
		n += atom.Len()
	}
	return
}

// Unmarshal func
func (inst *DataInfo) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	for n+8 < len(b) {
		tag := Tag(pio.U32BE(b[n+4:]))
		size := int(pio.U32BE(b[n:]))
		if len(b) < n+size {
			err = parseErr("TagSizeInvalid", n+offset, err)
			return
		}
		switch tag {
		case DREF:
			{
				atom := &DataRefer{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("dref", n+offset, err)
					return
				}
				inst.Refer = atom
			}
		default:
			{
				atom := &Dummy{TagItem: tag, Data: b[n : n+size]}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("", n+offset, err)
					return
				}
				inst.Unknowns = append(inst.Unknowns, atom)
			}
		}
		n += size
	}
	return
}

// Children func
func (inst DataInfo) Children() (r []Atom) {
	if inst.Refer != nil {
		r = append(r, inst.Refer)
	}
	r = append(r, inst.Unknowns...)
	return
}

// DataRefer struct
type DataRefer struct {
	Version uint8
	Flags   uint32
	URL     *DataReferURL
	AtomPos
}

// Marshal func
func (inst DataRefer) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(DREF))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst DataRefer) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], inst.Version)
	n++
	pio.PutU24BE(b[n:], inst.Flags)
	n += 3
	_childrenNR := 0
	if inst.URL != nil {
		_childrenNR++
	}
	pio.PutI32BE(b[n:], int32(_childrenNR))
	n += 4
	if inst.URL != nil {
		n += inst.URL.Marshal(b[n:])
	}
	return
}

// Len func
func (inst DataRefer) Len() (n int) {
	n += 8
	n++
	n += 3
	n += 4
	if inst.URL != nil {
		n += inst.URL.Len()
	}
	return
}

// Unmarshal func
func (inst *DataRefer) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	if len(b) < n+1 {
		err = parseErr("Version", n+offset, err)
		return
	}
	inst.Version = pio.U8(b[n:])
	n++
	if len(b) < n+3 {
		err = parseErr("Flags", n+offset, err)
		return
	}
	inst.Flags = pio.U24BE(b[n:])
	n += 3
	n += 4
	for n+8 < len(b) {
		tag := Tag(pio.U32BE(b[n+4:]))
		size := int(pio.U32BE(b[n:]))
		if len(b) < n+size {
			err = parseErr("TagSizeInvalid", n+offset, err)
			return
		}
		switch tag {
		case URL:
			{
				atom := &DataReferURL{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("url ", n+offset, err)
					return
				}
				inst.URL = atom
			}
		}
		n += size
	}
	return
}

// Children func
func (inst DataRefer) Children() (r []Atom) {
	if inst.URL != nil {
		r = append(r, inst.URL)
	}
	return
}

// DataReferURL struct
type DataReferURL struct {
	Version uint8
	Flags   uint32
	AtomPos
}

// Marshal func
func (inst DataReferURL) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(URL))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst DataReferURL) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], inst.Version)
	n++
	pio.PutU24BE(b[n:], inst.Flags)
	n += 3
	return
}

// Len func
func (inst DataReferURL) Len() (n int) {
	n += 8
	n++
	n += 3
	return
}

// Unmarshal func
func (inst *DataReferURL) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	if len(b) < n+1 {
		err = parseErr("Version", n+offset, err)
		return
	}
	inst.Version = pio.U8(b[n:])
	n++
	if len(b) < n+3 {
		err = parseErr("Flags", n+offset, err)
		return
	}
	inst.Flags = pio.U24BE(b[n:])
	n += 3
	return
}

// Children func
func (inst DataReferURL) Children() (r []Atom) {
	return
}

// SoundMediaInfo struct
type SoundMediaInfo struct {
	Version uint8
	Flags   uint32
	Balance int16
	AtomPos
}

// Marshal func
func (inst SoundMediaInfo) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(SMHD))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst SoundMediaInfo) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], inst.Version)
	n++
	pio.PutU24BE(b[n:], inst.Flags)
	n += 3
	pio.PutI16BE(b[n:], inst.Balance)
	n += 2
	n += 2
	return
}

// Len func
func (inst SoundMediaInfo) Len() (n int) {
	n += 8
	n++
	n += 3
	n += 2
	n += 2
	return
}

// Unmarshal func
func (inst *SoundMediaInfo) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	if len(b) < n+1 {
		err = parseErr("Version", n+offset, err)
		return
	}
	inst.Version = pio.U8(b[n:])
	n++
	if len(b) < n+3 {
		err = parseErr("Flags", n+offset, err)
		return
	}
	inst.Flags = pio.U24BE(b[n:])
	n += 3
	if len(b) < n+2 {
		err = parseErr("Balance", n+offset, err)
		return
	}
	inst.Balance = pio.I16BE(b[n:])
	n += 2
	n += 2
	return
}

// Children func
func (inst SoundMediaInfo) Children() (r []Atom) {
	return
}

// VideoMediaInfo struct
type VideoMediaInfo struct {
	Version      uint8
	Flags        uint32
	GraphicsMode int16
	Opcolor      [3]int16
	AtomPos
}

// Marshal func
func (inst VideoMediaInfo) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(VMHD))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst VideoMediaInfo) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], inst.Version)
	n++
	pio.PutU24BE(b[n:], inst.Flags)
	n += 3
	pio.PutI16BE(b[n:], inst.GraphicsMode)
	n += 2
	for _, entry := range inst.Opcolor {
		pio.PutI16BE(b[n:], entry)
		n += 2
	}
	return
}

// Len func
func (inst VideoMediaInfo) Len() (n int) {
	n += 8
	n++
	n += 3
	n += 2
	n += 2 * len(inst.Opcolor[:])
	return
}

// Unmarshal func
func (inst *VideoMediaInfo) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	if len(b) < n+1 {
		err = parseErr("Version", n+offset, err)
		return
	}
	inst.Version = pio.U8(b[n:])
	n++
	if len(b) < n+3 {
		err = parseErr("Flags", n+offset, err)
		return
	}
	inst.Flags = pio.U24BE(b[n:])
	n += 3
	if len(b) < n+2 {
		err = parseErr("GraphicsMode", n+offset, err)
		return
	}
	inst.GraphicsMode = pio.I16BE(b[n:])
	n += 2
	if len(b) < n+2*len(inst.Opcolor) {
		err = parseErr("Opcolor", n+offset, err)
		return
	}
	for i := range inst.Opcolor {
		inst.Opcolor[i] = pio.I16BE(b[n:])
		n += 2
	}
	return
}

// Children func
func (inst VideoMediaInfo) Children() (r []Atom) {
	return
}

// SampleTable struct
type SampleTable struct {
	SampleDesc        *SampleDesc
	TimeToSample      *TimeToSample
	CompositionOffset *CompositionOffset
	SampleToChunk     *SampleToChunk
	SyncSample        *SyncSample
	ChunkOffset       *ChunkOffset
	SampleSize        *SampleSize
	AtomPos
}

// Marshal func
func (inst SampleTable) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(STBL))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst SampleTable) marshal(b []byte) (n int) {
	if inst.SampleDesc != nil {
		n += inst.SampleDesc.Marshal(b[n:])
	}
	if inst.TimeToSample != nil {
		n += inst.TimeToSample.Marshal(b[n:])
	}
	if inst.CompositionOffset != nil {
		n += inst.CompositionOffset.Marshal(b[n:])
	}
	if inst.SampleToChunk != nil {
		n += inst.SampleToChunk.Marshal(b[n:])
	}
	if inst.SyncSample != nil {
		n += inst.SyncSample.Marshal(b[n:])
	}
	if inst.ChunkOffset != nil {
		n += inst.ChunkOffset.Marshal(b[n:])
	}
	if inst.SampleSize != nil {
		n += inst.SampleSize.Marshal(b[n:])
	}
	return
}

// Len func
func (inst SampleTable) Len() (n int) {
	n += 8
	if inst.SampleDesc != nil {
		n += inst.SampleDesc.Len()
	}
	if inst.TimeToSample != nil {
		n += inst.TimeToSample.Len()
	}
	if inst.CompositionOffset != nil {
		n += inst.CompositionOffset.Len()
	}
	if inst.SampleToChunk != nil {
		n += inst.SampleToChunk.Len()
	}
	if inst.SyncSample != nil {
		n += inst.SyncSample.Len()
	}
	if inst.ChunkOffset != nil {
		n += inst.ChunkOffset.Len()
	}
	if inst.SampleSize != nil {
		n += inst.SampleSize.Len()
	}
	return
}

// Unmarshal func
func (inst *SampleTable) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	for n+8 < len(b) {
		tag := Tag(pio.U32BE(b[n+4:]))
		size := int(pio.U32BE(b[n:]))
		if len(b) < n+size {
			err = parseErr("TagSizeInvalid", n+offset, err)
			return
		}
		switch tag {
		case STSD:
			{
				atom := &SampleDesc{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("stsd", n+offset, err)
					return
				}
				inst.SampleDesc = atom
			}
		case STTS:
			{
				atom := &TimeToSample{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("stts", n+offset, err)
					return
				}
				inst.TimeToSample = atom
			}
		case CTTS:
			{
				atom := &CompositionOffset{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("ctts", n+offset, err)
					return
				}
				inst.CompositionOffset = atom
			}
		case STSC:
			{
				atom := &SampleToChunk{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("stsc", n+offset, err)
					return
				}
				inst.SampleToChunk = atom
			}
		case STSS:
			{
				atom := &SyncSample{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("stss", n+offset, err)
					return
				}
				inst.SyncSample = atom
			}
		case STCO:
			{
				atom := &ChunkOffset{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("stco", n+offset, err)
					return
				}
				inst.ChunkOffset = atom
			}
		case STSZ:
			{
				atom := &SampleSize{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("stsz", n+offset, err)
					return
				}
				inst.SampleSize = atom
			}
		}
		n += size
	}
	return
}

// Children func
func (inst SampleTable) Children() (r []Atom) {
	if inst.SampleDesc != nil {
		r = append(r, inst.SampleDesc)
	}
	if inst.TimeToSample != nil {
		r = append(r, inst.TimeToSample)
	}
	if inst.CompositionOffset != nil {
		r = append(r, inst.CompositionOffset)
	}
	if inst.SampleToChunk != nil {
		r = append(r, inst.SampleToChunk)
	}
	if inst.SyncSample != nil {
		r = append(r, inst.SyncSample)
	}
	if inst.ChunkOffset != nil {
		r = append(r, inst.ChunkOffset)
	}
	if inst.SampleSize != nil {
		r = append(r, inst.SampleSize)
	}
	return
}

// SampleDesc struct
type SampleDesc struct {
	Version  uint8
	AVC1Desc *AVC1Desc
	MP4ADesc *MP4ADesc
	Unknowns []Atom
	AtomPos
}

// Marshal func
func (inst SampleDesc) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(STSD))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst SampleDesc) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], inst.Version)
	n++
	n += 3
	_childrenNR := 0
	if inst.AVC1Desc != nil {
		_childrenNR++
	}
	if inst.MP4ADesc != nil {
		_childrenNR++
	}
	_childrenNR += len(inst.Unknowns)
	pio.PutI32BE(b[n:], int32(_childrenNR))
	n += 4
	if inst.AVC1Desc != nil {
		n += inst.AVC1Desc.Marshal(b[n:])
	}
	if inst.MP4ADesc != nil {
		n += inst.MP4ADesc.Marshal(b[n:])
	}
	for _, atom := range inst.Unknowns {
		n += atom.Marshal(b[n:])
	}
	return
}

// Len func
func (inst SampleDesc) Len() (n int) {
	n += 8
	n++
	n += 3
	n += 4
	if inst.AVC1Desc != nil {
		n += inst.AVC1Desc.Len()
	}
	if inst.MP4ADesc != nil {
		n += inst.MP4ADesc.Len()
	}
	for _, atom := range inst.Unknowns {
		n += atom.Len()
	}
	return
}

// Unmarshal func
func (inst *SampleDesc) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	if len(b) < n+1 {
		err = parseErr("Version", n+offset, err)
		return
	}
	inst.Version = pio.U8(b[n:])
	n++
	n += 3
	n += 4
	for n+8 < len(b) {
		tag := Tag(pio.U32BE(b[n+4:]))
		size := int(pio.U32BE(b[n:]))
		if len(b) < n+size {
			err = parseErr("TagSizeInvalid", n+offset, err)
			return
		}
		switch tag {
		case AVC1:
			{
				atom := &AVC1Desc{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("avc1", n+offset, err)
					return
				}
				inst.AVC1Desc = atom
			}
		case MP4A:
			{
				atom := &MP4ADesc{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("mp4a", n+offset, err)
					return
				}
				inst.MP4ADesc = atom
			}
		default:
			{
				atom := &Dummy{TagItem: tag, Data: b[n : n+size]}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("", n+offset, err)
					return
				}
				inst.Unknowns = append(inst.Unknowns, atom)
			}
		}
		n += size
	}
	return
}

// Children func
func (inst SampleDesc) Children() (r []Atom) {
	if inst.AVC1Desc != nil {
		r = append(r, inst.AVC1Desc)
	}
	if inst.MP4ADesc != nil {
		r = append(r, inst.MP4ADesc)
	}
	r = append(r, inst.Unknowns...)
	return
}

// MP4ADesc struct
type MP4ADesc struct {
	DataRefIdx       int16
	Version          int16
	RevisionLevel    int16
	Vendor           int32
	NumberOfChannels int16
	SampleSize       int16
	CompressionID    int16
	SampleRate       float64
	Conf             *ElemStreamDesc
	Unknowns         []Atom
	AtomPos
}

// Marshal func
func (inst MP4ADesc) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(MP4A))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst MP4ADesc) marshal(b []byte) (n int) {
	n += 6
	pio.PutI16BE(b[n:], inst.DataRefIdx)
	n += 2
	pio.PutI16BE(b[n:], inst.Version)
	n += 2
	pio.PutI16BE(b[n:], inst.RevisionLevel)
	n += 2
	pio.PutI32BE(b[n:], inst.Vendor)
	n += 4
	pio.PutI16BE(b[n:], inst.NumberOfChannels)
	n += 2
	pio.PutI16BE(b[n:], inst.SampleSize)
	n += 2
	pio.PutI16BE(b[n:], inst.CompressionID)
	n += 2
	n += 2
	PutFixed32(b[n:], inst.SampleRate)
	n += 4
	if inst.Conf != nil {
		n += inst.Conf.Marshal(b[n:])
	}
	for _, atom := range inst.Unknowns {
		n += atom.Marshal(b[n:])
	}
	return
}

// Len func
func (inst MP4ADesc) Len() (n int) {
	n += 8
	n += 6
	n += 2
	n += 2
	n += 2
	n += 4
	n += 2
	n += 2
	n += 2
	n += 2
	n += 4
	if inst.Conf != nil {
		n += inst.Conf.Len()
	}
	for _, atom := range inst.Unknowns {
		n += atom.Len()
	}
	return
}

// Unmarshal func
func (inst *MP4ADesc) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	n += 6
	if len(b) < n+2 {
		err = parseErr("DataRefIdx", n+offset, err)
		return
	}
	inst.DataRefIdx = pio.I16BE(b[n:])
	n += 2
	if len(b) < n+2 {
		err = parseErr("Version", n+offset, err)
		return
	}
	inst.Version = pio.I16BE(b[n:])
	n += 2
	if len(b) < n+2 {
		err = parseErr("RevisionLevel", n+offset, err)
		return
	}
	inst.RevisionLevel = pio.I16BE(b[n:])
	n += 2
	if len(b) < n+4 {
		err = parseErr("Vendor", n+offset, err)
		return
	}
	inst.Vendor = pio.I32BE(b[n:])
	n += 4
	if len(b) < n+2 {
		err = parseErr("NumberOfChannels", n+offset, err)
		return
	}
	inst.NumberOfChannels = pio.I16BE(b[n:])
	n += 2
	if len(b) < n+2 {
		err = parseErr("SampleSize", n+offset, err)
		return
	}
	inst.SampleSize = pio.I16BE(b[n:])
	n += 2
	if len(b) < n+2 {
		err = parseErr("CompressionID", n+offset, err)
		return
	}
	inst.CompressionID = pio.I16BE(b[n:])
	n += 2
	n += 2
	if len(b) < n+4 {
		err = parseErr("SampleRate", n+offset, err)
		return
	}
	inst.SampleRate = GetFixed32(b[n:])
	n += 4
	for n+8 < len(b) {
		tag := Tag(pio.U32BE(b[n+4:]))
		size := int(pio.U32BE(b[n:]))
		if len(b) < n+size {
			err = parseErr("TagSizeInvalid", n+offset, err)
			return
		}
		switch tag {
		case ESDS:
			{
				atom := &ElemStreamDesc{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("esds", n+offset, err)
					return
				}
				inst.Conf = atom
			}
		default:
			{
				atom := &Dummy{TagItem: tag, Data: b[n : n+size]}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("", n+offset, err)
					return
				}
				inst.Unknowns = append(inst.Unknowns, atom)
			}
		}
		n += size
	}
	return
}

// Children func
func (inst MP4ADesc) Children() (r []Atom) {
	if inst.Conf != nil {
		r = append(r, inst.Conf)
	}
	r = append(r, inst.Unknowns...)
	return
}

// AVC1Desc struct
type AVC1Desc struct {
	DataRefIdx           int16
	Version              int16
	Revision             int16
	Vendor               int32
	TemporalQuality      int32
	SpatialQuality       int32
	Width                int16
	Height               int16
	HorizontalResolution float64
	VorizontalResolution float64
	FrameCount           int16
	CompressorName       [32]byte
	Depth                int16
	ColorTableID         int16
	Conf                 *AVC1Conf
	Unknowns             []Atom
	AtomPos
}

// Marshal func
func (inst AVC1Desc) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(AVC1))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst AVC1Desc) marshal(b []byte) (n int) {
	n += 6
	pio.PutI16BE(b[n:], inst.DataRefIdx)
	n += 2
	pio.PutI16BE(b[n:], inst.Version)
	n += 2
	pio.PutI16BE(b[n:], inst.Revision)
	n += 2
	pio.PutI32BE(b[n:], inst.Vendor)
	n += 4
	pio.PutI32BE(b[n:], inst.TemporalQuality)
	n += 4
	pio.PutI32BE(b[n:], inst.SpatialQuality)
	n += 4
	pio.PutI16BE(b[n:], inst.Width)
	n += 2
	pio.PutI16BE(b[n:], inst.Height)
	n += 2
	PutFixed32(b[n:], inst.HorizontalResolution)
	n += 4
	PutFixed32(b[n:], inst.VorizontalResolution)
	n += 4
	n += 4
	pio.PutI16BE(b[n:], inst.FrameCount)
	n += 2
	copy(b[n:], inst.CompressorName[:])
	n += len(inst.CompressorName[:])
	pio.PutI16BE(b[n:], inst.Depth)
	n += 2
	pio.PutI16BE(b[n:], inst.ColorTableID)
	n += 2
	if inst.Conf != nil {
		n += inst.Conf.Marshal(b[n:])
	}
	for _, atom := range inst.Unknowns {
		n += atom.Marshal(b[n:])
	}
	return
}

// Len func
func (inst AVC1Desc) Len() (n int) {
	n += 8
	n += 6
	n += 2
	n += 2
	n += 2
	n += 4
	n += 4
	n += 4
	n += 2
	n += 2
	n += 4
	n += 4
	n += 4
	n += 2
	n += len(inst.CompressorName[:])
	n += 2
	n += 2
	if inst.Conf != nil {
		n += inst.Conf.Len()
	}
	for _, atom := range inst.Unknowns {
		n += atom.Len()
	}
	return
}

// Unmarshal func
func (inst *AVC1Desc) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	n += 6
	if len(b) < n+2 {
		err = parseErr("DataRefIdx", n+offset, err)
		return
	}
	inst.DataRefIdx = pio.I16BE(b[n:])
	n += 2
	if len(b) < n+2 {
		err = parseErr("Version", n+offset, err)
		return
	}
	inst.Version = pio.I16BE(b[n:])
	n += 2
	if len(b) < n+2 {
		err = parseErr("Revision", n+offset, err)
		return
	}
	inst.Revision = pio.I16BE(b[n:])
	n += 2
	if len(b) < n+4 {
		err = parseErr("Vendor", n+offset, err)
		return
	}
	inst.Vendor = pio.I32BE(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("TemporalQuality", n+offset, err)
		return
	}
	inst.TemporalQuality = pio.I32BE(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("SpatialQuality", n+offset, err)
		return
	}
	inst.SpatialQuality = pio.I32BE(b[n:])
	n += 4
	if len(b) < n+2 {
		err = parseErr("Width", n+offset, err)
		return
	}
	inst.Width = pio.I16BE(b[n:])
	n += 2
	if len(b) < n+2 {
		err = parseErr("Height", n+offset, err)
		return
	}
	inst.Height = pio.I16BE(b[n:])
	n += 2
	if len(b) < n+4 {
		err = parseErr("HorizontalResolution", n+offset, err)
		return
	}
	inst.HorizontalResolution = GetFixed32(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("VorizontalResolution", n+offset, err)
		return
	}
	inst.VorizontalResolution = GetFixed32(b[n:])
	n += 4
	n += 4
	if len(b) < n+2 {
		err = parseErr("FrameCount", n+offset, err)
		return
	}
	inst.FrameCount = pio.I16BE(b[n:])
	n += 2
	if len(b) < n+len(inst.CompressorName) {
		err = parseErr("CompressorName", n+offset, err)
		return
	}
	copy(inst.CompressorName[:], b[n:])
	n += len(inst.CompressorName)
	if len(b) < n+2 {
		err = parseErr("Depth", n+offset, err)
		return
	}
	inst.Depth = pio.I16BE(b[n:])
	n += 2
	if len(b) < n+2 {
		err = parseErr("ColorTableID", n+offset, err)
		return
	}
	inst.ColorTableID = pio.I16BE(b[n:])
	n += 2
	for n+8 < len(b) {
		tag := Tag(pio.U32BE(b[n+4:]))
		size := int(pio.U32BE(b[n:]))
		if len(b) < n+size {
			err = parseErr("TagSizeInvalid", n+offset, err)
			return
		}
		switch tag {
		case AVCC:
			{
				atom := &AVC1Conf{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("avcC", n+offset, err)
					return
				}
				inst.Conf = atom
			}
		default:
			{
				atom := &Dummy{TagItem: tag, Data: b[n : n+size]}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("", n+offset, err)
					return
				}
				inst.Unknowns = append(inst.Unknowns, atom)
			}
		}
		n += size
	}
	return
}

// Children func
func (inst AVC1Desc) Children() (r []Atom) {
	if inst.Conf != nil {
		r = append(r, inst.Conf)
	}
	r = append(r, inst.Unknowns...)
	return
}

// AVC1Conf struct
type AVC1Conf struct {
	Data []byte
	AtomPos
}

// Marshal func
func (inst AVC1Conf) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(AVCC))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst AVC1Conf) marshal(b []byte) (n int) {
	copy(b[n:], inst.Data[:])
	n += len(inst.Data[:])
	return
}

// Len func
func (inst AVC1Conf) Len() (n int) {
	n += 8
	n += len(inst.Data[:])
	return
}

// Unmarshal func
func (inst *AVC1Conf) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	inst.Data = b[n:]
	n += len(b[n:])
	return
}

// Children func
func (inst AVC1Conf) Children() (r []Atom) {
	return
}

// TimeToSample struct
type TimeToSample struct {
	Version uint8
	Flags   uint32
	Entries []TimeToSampleEntry
	AtomPos
}

// Marshal func
func (inst TimeToSample) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(STTS))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst TimeToSample) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], inst.Version)
	n++
	pio.PutU24BE(b[n:], inst.Flags)
	n += 3
	pio.PutU32BE(b[n:], uint32(len(inst.Entries)))
	n += 4
	for _, entry := range inst.Entries {
		PutTimeToSampleEntry(b[n:], entry)
		n += LenTimeToSampleEntry
	}
	return
}

// Len func
func (inst TimeToSample) Len() (n int) {
	n += 8
	n++
	n += 3
	n += 4
	n += LenTimeToSampleEntry * len(inst.Entries)
	return
}

// Unmarshal func
func (inst *TimeToSample) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	if len(b) < n+1 {
		err = parseErr("Version", n+offset, err)
		return
	}
	inst.Version = pio.U8(b[n:])
	n++
	if len(b) < n+3 {
		err = parseErr("Flags", n+offset, err)
		return
	}
	inst.Flags = pio.U24BE(b[n:])
	n += 3
	var lenEntries uint32
	lenEntries = pio.U32BE(b[n:])
	n += 4
	inst.Entries = make([]TimeToSampleEntry, lenEntries)
	if len(b) < n+LenTimeToSampleEntry*len(inst.Entries) {
		err = parseErr("TimeToSampleEntry", n+offset, err)
		return
	}
	for i := range inst.Entries {
		inst.Entries[i] = GetTimeToSampleEntry(b[n:])
		n += LenTimeToSampleEntry
	}
	return
}

// Children func
func (inst TimeToSample) Children() (r []Atom) {
	return
}

// TimeToSampleEntry struct
type TimeToSampleEntry struct {
	Count    uint32
	Duration uint32
}

// GetTimeToSampleEntry func
func GetTimeToSampleEntry(b []byte) (inst TimeToSampleEntry) {
	inst.Count = pio.U32BE(b[0:])
	inst.Duration = pio.U32BE(b[4:])
	return
}

// PutTimeToSampleEntry func
func PutTimeToSampleEntry(b []byte, inst TimeToSampleEntry) {
	pio.PutU32BE(b[0:], inst.Count)
	pio.PutU32BE(b[4:], inst.Duration)
}

// LenTimeToSampleEntry const
const LenTimeToSampleEntry = 8

// SampleToChunk struct
type SampleToChunk struct {
	Version uint8
	Flags   uint32
	Entries []SampleToChunkEntry
	AtomPos
}

// Marshal func
func (inst SampleToChunk) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(STSC))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst SampleToChunk) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], inst.Version)
	n++
	pio.PutU24BE(b[n:], inst.Flags)
	n += 3
	pio.PutU32BE(b[n:], uint32(len(inst.Entries)))
	n += 4
	for _, entry := range inst.Entries {
		PutSampleToChunkEntry(b[n:], entry)
		n += LenSampleToChunkEntry
	}
	return
}

// Len func
func (inst SampleToChunk) Len() (n int) {
	n += 8
	n++
	n += 3
	n += 4
	n += LenSampleToChunkEntry * len(inst.Entries)
	return
}

// Unmarshal func
func (inst *SampleToChunk) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	if len(b) < n+1 {
		err = parseErr("Version", n+offset, err)
		return
	}
	inst.Version = pio.U8(b[n:])
	n++
	if len(b) < n+3 {
		err = parseErr("Flags", n+offset, err)
		return
	}
	inst.Flags = pio.U24BE(b[n:])
	n += 3
	var lenEntries uint32
	lenEntries = pio.U32BE(b[n:])
	n += 4
	inst.Entries = make([]SampleToChunkEntry, lenEntries)
	if len(b) < n+LenSampleToChunkEntry*len(inst.Entries) {
		err = parseErr("SampleToChunkEntry", n+offset, err)
		return
	}
	for i := range inst.Entries {
		inst.Entries[i] = GetSampleToChunkEntry(b[n:])
		n += LenSampleToChunkEntry
	}
	return
}

// Children func
func (inst SampleToChunk) Children() (r []Atom) {
	return
}

// SampleToChunkEntry struct
type SampleToChunkEntry struct {
	FirstChunk      uint32
	SamplesPerChunk uint32
	SampleDescID    uint32
}

// GetSampleToChunkEntry func
func GetSampleToChunkEntry(b []byte) (inst SampleToChunkEntry) {
	inst.FirstChunk = pio.U32BE(b[0:])
	inst.SamplesPerChunk = pio.U32BE(b[4:])
	inst.SampleDescID = pio.U32BE(b[8:])
	return
}

// PutSampleToChunkEntry func
func PutSampleToChunkEntry(b []byte, inst SampleToChunkEntry) {
	pio.PutU32BE(b[0:], inst.FirstChunk)
	pio.PutU32BE(b[4:], inst.SamplesPerChunk)
	pio.PutU32BE(b[8:], inst.SampleDescID)
}

// LenSampleToChunkEntry const
const LenSampleToChunkEntry = 12

// CompositionOffset struct
type CompositionOffset struct {
	Version uint8
	Flags   uint32
	Entries []CompositionOffsetEntry
	AtomPos
}

// Marshal func
func (inst CompositionOffset) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(CTTS))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst CompositionOffset) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], inst.Version)
	n++
	pio.PutU24BE(b[n:], inst.Flags)
	n += 3
	pio.PutU32BE(b[n:], uint32(len(inst.Entries)))
	n += 4
	for _, entry := range inst.Entries {
		PutCompositionOffsetEntry(b[n:], entry)
		n += LenCompositionOffsetEntry
	}
	return
}

// Len func
func (inst CompositionOffset) Len() (n int) {
	n += 8
	n++
	n += 3
	n += 4
	n += LenCompositionOffsetEntry * len(inst.Entries)
	return
}

// Unmarshal func
func (inst *CompositionOffset) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	if len(b) < n+1 {
		err = parseErr("Version", n+offset, err)
		return
	}
	inst.Version = pio.U8(b[n:])
	n++
	if len(b) < n+3 {
		err = parseErr("Flags", n+offset, err)
		return
	}
	inst.Flags = pio.U24BE(b[n:])
	n += 3
	var lenEntries uint32
	lenEntries = pio.U32BE(b[n:])
	n += 4
	inst.Entries = make([]CompositionOffsetEntry, lenEntries)
	if len(b) < n+LenCompositionOffsetEntry*len(inst.Entries) {
		err = parseErr("CompositionOffsetEntry", n+offset, err)
		return
	}
	for i := range inst.Entries {
		inst.Entries[i] = GetCompositionOffsetEntry(b[n:])
		n += LenCompositionOffsetEntry
	}
	return
}

// Children func
func (inst CompositionOffset) Children() (r []Atom) {
	return
}

// CompositionOffsetEntry struct
type CompositionOffsetEntry struct {
	Count  uint32
	Offset uint32
}

// GetCompositionOffsetEntry func
func GetCompositionOffsetEntry(b []byte) (inst CompositionOffsetEntry) {
	inst.Count = pio.U32BE(b[0:])
	inst.Offset = pio.U32BE(b[4:])
	return
}

// PutCompositionOffsetEntry func
func PutCompositionOffsetEntry(b []byte, inst CompositionOffsetEntry) {
	pio.PutU32BE(b[0:], inst.Count)
	pio.PutU32BE(b[4:], inst.Offset)
}

// LenCompositionOffsetEntry func
const LenCompositionOffsetEntry = 8

// SyncSample struct
type SyncSample struct {
	Version uint8
	Flags   uint32
	Entries []uint32
	AtomPos
}

// Marshal func
func (inst SyncSample) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(STSS))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst SyncSample) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], inst.Version)
	n++
	pio.PutU24BE(b[n:], inst.Flags)
	n += 3
	pio.PutU32BE(b[n:], uint32(len(inst.Entries)))
	n += 4
	for _, entry := range inst.Entries {
		pio.PutU32BE(b[n:], entry)
		n += 4
	}
	return
}

// Len func
func (inst SyncSample) Len() (n int) {
	n += 8
	n++
	n += 3
	n += 4
	n += 4 * len(inst.Entries)
	return
}

// Unmarshal func
func (inst *SyncSample) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	if len(b) < n+1 {
		err = parseErr("Version", n+offset, err)
		return
	}
	inst.Version = pio.U8(b[n:])
	n++
	if len(b) < n+3 {
		err = parseErr("Flags", n+offset, err)
		return
	}
	inst.Flags = pio.U24BE(b[n:])
	n += 3
	var lenEntries uint32
	lenEntries = pio.U32BE(b[n:])
	n += 4
	inst.Entries = make([]uint32, lenEntries)
	if len(b) < n+4*len(inst.Entries) {
		err = parseErr("uint32", n+offset, err)
		return
	}
	for i := range inst.Entries {
		inst.Entries[i] = pio.U32BE(b[n:])
		n += 4
	}
	return
}

// Children func
func (inst SyncSample) Children() (r []Atom) {
	return
}

// ChunkOffset struct
type ChunkOffset struct {
	Version uint8
	Flags   uint32
	Entries []uint32
	AtomPos
}

// Marshal func
func (inst ChunkOffset) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(STCO))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst ChunkOffset) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], inst.Version)
	n++
	pio.PutU24BE(b[n:], inst.Flags)
	n += 3
	pio.PutU32BE(b[n:], uint32(len(inst.Entries)))
	n += 4
	for _, entry := range inst.Entries {
		pio.PutU32BE(b[n:], entry)
		n += 4
	}
	return
}

// Len func
func (inst ChunkOffset) Len() (n int) {
	n += 8
	n++
	n += 3
	n += 4
	n += 4 * len(inst.Entries)
	return
}

// Unmarshal func
func (inst *ChunkOffset) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	if len(b) < n+1 {
		err = parseErr("Version", n+offset, err)
		return
	}
	inst.Version = pio.U8(b[n:])
	n++
	if len(b) < n+3 {
		err = parseErr("Flags", n+offset, err)
		return
	}
	inst.Flags = pio.U24BE(b[n:])
	n += 3
	var lenEntries uint32
	lenEntries = pio.U32BE(b[n:])
	n += 4
	inst.Entries = make([]uint32, lenEntries)
	if len(b) < n+4*len(inst.Entries) {
		err = parseErr("uint32", n+offset, err)
		return
	}
	for i := range inst.Entries {
		inst.Entries[i] = pio.U32BE(b[n:])
		n += 4
	}
	return
}

// Children func
func (inst ChunkOffset) Children() (r []Atom) {
	return
}

// MovieFrag struct
type MovieFrag struct {
	Header   *MovieFragHeader
	Tracks   []*TrackFrag
	Unknowns []Atom
	AtomPos
}

// Marshal func
func (inst MovieFrag) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(MOOF))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst MovieFrag) marshal(b []byte) (n int) {
	if inst.Header != nil {
		n += inst.Header.Marshal(b[n:])
	}
	for _, atom := range inst.Tracks {
		n += atom.Marshal(b[n:])
	}
	for _, atom := range inst.Unknowns {
		n += atom.Marshal(b[n:])
	}
	return
}

// Len func
func (inst MovieFrag) Len() (n int) {
	n += 8
	if inst.Header != nil {
		n += inst.Header.Len()
	}
	for _, atom := range inst.Tracks {
		n += atom.Len()
	}
	for _, atom := range inst.Unknowns {
		n += atom.Len()
	}
	return
}

// Unmarshal func
func (inst *MovieFrag) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	for n+8 < len(b) {
		tag := Tag(pio.U32BE(b[n+4:]))
		size := int(pio.U32BE(b[n:]))
		if len(b) < n+size {
			err = parseErr("TagSizeInvalid", n+offset, err)
			return
		}
		switch tag {
		case MFHD:
			{
				atom := &MovieFragHeader{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("mfhd", n+offset, err)
					return
				}
				inst.Header = atom
			}
		case TRAF:
			{
				atom := &TrackFrag{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("traf", n+offset, err)
					return
				}
				inst.Tracks = append(inst.Tracks, atom)
			}
		default:
			{
				atom := &Dummy{TagItem: tag, Data: b[n : n+size]}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("", n+offset, err)
					return
				}
				inst.Unknowns = append(inst.Unknowns, atom)
			}
		}
		n += size
	}
	return
}

// Children func
func (inst MovieFrag) Children() (r []Atom) {
	if inst.Header != nil {
		r = append(r, inst.Header)
	}
	for _, atom := range inst.Tracks {
		r = append(r, atom)
	}
	r = append(r, inst.Unknowns...)
	return
}

// MovieFragHeader struct
type MovieFragHeader struct {
	Version uint8
	Flags   uint32
	Seqnum  uint32
	AtomPos
}

// Marshal func
func (inst MovieFragHeader) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(MFHD))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst MovieFragHeader) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], inst.Version)
	n++
	pio.PutU24BE(b[n:], inst.Flags)
	n += 3
	pio.PutU32BE(b[n:], inst.Seqnum)
	n += 4
	return
}

// Len func
func (inst MovieFragHeader) Len() (n int) {
	n += 8
	n++
	n += 3
	n += 4
	return
}

// Unmarshal func
func (inst *MovieFragHeader) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	if len(b) < n+1 {
		err = parseErr("Version", n+offset, err)
		return
	}
	inst.Version = pio.U8(b[n:])
	n++
	if len(b) < n+3 {
		err = parseErr("Flags", n+offset, err)
		return
	}
	inst.Flags = pio.U24BE(b[n:])
	n += 3
	if len(b) < n+4 {
		err = parseErr("Seqnum", n+offset, err)
		return
	}
	inst.Seqnum = pio.U32BE(b[n:])
	n += 4
	return
}

// Children func
func (inst MovieFragHeader) Children() (r []Atom) {
	return
}

// TrackFrag struct
type TrackFrag struct {
	Header     *TrackFragHeader
	DecodeTime *TrackFragDecodeTime
	Run        *TrackFragRun
	Unknowns   []Atom
	AtomPos
}

// Marshal func
func (inst TrackFrag) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(TRAF))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst TrackFrag) marshal(b []byte) (n int) {
	if inst.Header != nil {
		n += inst.Header.Marshal(b[n:])
	}
	if inst.DecodeTime != nil {
		n += inst.DecodeTime.Marshal(b[n:])
	}
	if inst.Run != nil {
		n += inst.Run.Marshal(b[n:])
	}
	for _, atom := range inst.Unknowns {
		n += atom.Marshal(b[n:])
	}
	return
}

// Len func
func (inst TrackFrag) Len() (n int) {
	n += 8
	if inst.Header != nil {
		n += inst.Header.Len()
	}
	if inst.DecodeTime != nil {
		n += inst.DecodeTime.Len()
	}
	if inst.Run != nil {
		n += inst.Run.Len()
	}
	for _, atom := range inst.Unknowns {
		n += atom.Len()
	}
	return
}

// Unmarshal func
func (inst *TrackFrag) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	for n+8 < len(b) {
		tag := Tag(pio.U32BE(b[n+4:]))
		size := int(pio.U32BE(b[n:]))
		if len(b) < n+size {
			err = parseErr("TagSizeInvalid", n+offset, err)
			return
		}
		switch tag {
		case TFHD:
			{
				atom := &TrackFragHeader{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("tfhd", n+offset, err)
					return
				}
				inst.Header = atom
			}
		case TFDT:
			{
				atom := &TrackFragDecodeTime{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("tfdt", n+offset, err)
					return
				}
				inst.DecodeTime = atom
			}
		case TRUN:
			{
				atom := &TrackFragRun{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("trun", n+offset, err)
					return
				}
				inst.Run = atom
			}
		default:
			{
				atom := &Dummy{TagItem: tag, Data: b[n : n+size]}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("", n+offset, err)
					return
				}
				inst.Unknowns = append(inst.Unknowns, atom)
			}
		}
		n += size
	}
	return
}

// Children func
func (inst TrackFrag) Children() (r []Atom) {
	if inst.Header != nil {
		r = append(r, inst.Header)
	}
	if inst.DecodeTime != nil {
		r = append(r, inst.DecodeTime)
	}
	if inst.Run != nil {
		r = append(r, inst.Run)
	}
	r = append(r, inst.Unknowns...)
	return
}

// MovieExtend struct
type MovieExtend struct {
	Tracks   []*TrackExtend
	Unknowns []Atom
	AtomPos
}

// Marshal func
func (inst MovieExtend) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(MVEX))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst MovieExtend) marshal(b []byte) (n int) {
	for _, atom := range inst.Tracks {
		n += atom.Marshal(b[n:])
	}
	for _, atom := range inst.Unknowns {
		n += atom.Marshal(b[n:])
	}
	return
}

// Len func
func (inst MovieExtend) Len() (n int) {
	n += 8
	for _, atom := range inst.Tracks {
		n += atom.Len()
	}
	for _, atom := range inst.Unknowns {
		n += atom.Len()
	}
	return
}

// Unmarshal func
func (inst *MovieExtend) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	for n+8 < len(b) {
		tag := Tag(pio.U32BE(b[n+4:]))
		size := int(pio.U32BE(b[n:]))
		if len(b) < n+size {
			err = parseErr("TagSizeInvalid", n+offset, err)
			return
		}
		switch tag {
		case TREX:
			{
				atom := &TrackExtend{}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("trex", n+offset, err)
					return
				}
				inst.Tracks = append(inst.Tracks, atom)
			}
		default:
			{
				atom := &Dummy{TagItem: tag, Data: b[n : n+size]}
				if _, err = atom.Unmarshal(b[n:n+size], offset+n); err != nil {
					err = parseErr("", n+offset, err)
					return
				}
				inst.Unknowns = append(inst.Unknowns, atom)
			}
		}
		n += size
	}
	return
}

// Children func
func (inst MovieExtend) Children() (r []Atom) {
	for _, atom := range inst.Tracks {
		r = append(r, atom)
	}
	r = append(r, inst.Unknowns...)
	return
}

// TrackExtend struct
type TrackExtend struct {
	Version               uint8
	Flags                 uint32
	TrackID               uint32
	DefaultSampleDescIdx  uint32
	DefaultSampleDuration uint32
	DefaultSampleSize     uint32
	DefaultSampleFlags    uint32
	AtomPos
}

// Marshal func
func (inst TrackExtend) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(TREX))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst TrackExtend) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], inst.Version)
	n++
	pio.PutU24BE(b[n:], inst.Flags)
	n += 3
	pio.PutU32BE(b[n:], inst.TrackID)
	n += 4
	pio.PutU32BE(b[n:], inst.DefaultSampleDescIdx)
	n += 4
	pio.PutU32BE(b[n:], inst.DefaultSampleDuration)
	n += 4
	pio.PutU32BE(b[n:], inst.DefaultSampleSize)
	n += 4
	pio.PutU32BE(b[n:], inst.DefaultSampleFlags)
	n += 4
	return
}

// Len func
func (inst TrackExtend) Len() (n int) {
	n += 8
	n++
	n += 3
	n += 4
	n += 4
	n += 4
	n += 4
	n += 4
	return
}

// Unmarshal func
func (inst *TrackExtend) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	if len(b) < n+1 {
		err = parseErr("Version", n+offset, err)
		return
	}
	inst.Version = pio.U8(b[n:])
	n++
	if len(b) < n+3 {
		err = parseErr("Flags", n+offset, err)
		return
	}
	inst.Flags = pio.U24BE(b[n:])
	n += 3
	if len(b) < n+4 {
		err = parseErr("TrackID", n+offset, err)
		return
	}
	inst.TrackID = pio.U32BE(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("DefaultSampleDescIdx", n+offset, err)
		return
	}
	inst.DefaultSampleDescIdx = pio.U32BE(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("DefaultSampleDuration", n+offset, err)
		return
	}
	inst.DefaultSampleDuration = pio.U32BE(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("DefaultSampleSize", n+offset, err)
		return
	}
	inst.DefaultSampleSize = pio.U32BE(b[n:])
	n += 4
	if len(b) < n+4 {
		err = parseErr("DefaultSampleFlags", n+offset, err)
		return
	}
	inst.DefaultSampleFlags = pio.U32BE(b[n:])
	n += 4
	return
}

// Children func
func (inst TrackExtend) Children() (r []Atom) {
	return
}

// SampleSize struct
type SampleSize struct {
	Version    uint8
	Flags      uint32
	SampleSize uint32
	Entries    []uint32
	AtomPos
}

// Marshal func
func (inst SampleSize) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(STSZ))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst SampleSize) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], inst.Version)
	n++
	pio.PutU24BE(b[n:], inst.Flags)
	n += 3
	pio.PutU32BE(b[n:], inst.SampleSize)
	n += 4
	if inst.SampleSize != 0 {
		return
	}
	pio.PutU32BE(b[n:], uint32(len(inst.Entries)))
	n += 4
	for _, entry := range inst.Entries {
		pio.PutU32BE(b[n:], entry)
		n += 4
	}
	return
}

// Len func
func (inst SampleSize) Len() (n int) {
	n += 8
	n++
	n += 3
	n += 4
	if inst.SampleSize != 0 {
		return
	}
	n += 4
	n += 4 * len(inst.Entries)
	return
}

// Unmarshal func
func (inst *SampleSize) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	if len(b) < n+1 {
		err = parseErr("Version", n+offset, err)
		return
	}
	inst.Version = pio.U8(b[n:])
	n++
	if len(b) < n+3 {
		err = parseErr("Flags", n+offset, err)
		return
	}
	inst.Flags = pio.U24BE(b[n:])
	n += 3
	if len(b) < n+4 {
		err = parseErr("SampleSize", n+offset, err)
		return
	}
	inst.SampleSize = pio.U32BE(b[n:])
	n += 4
	if inst.SampleSize != 0 {
		return
	}
	var lenEntries uint32
	lenEntries = pio.U32BE(b[n:])
	n += 4
	inst.Entries = make([]uint32, lenEntries)
	if len(b) < n+4*len(inst.Entries) {
		err = parseErr("uint32", n+offset, err)
		return
	}
	for i := range inst.Entries {
		inst.Entries[i] = pio.U32BE(b[n:])
		n += 4
	}
	return
}

// Children func
func (inst SampleSize) Children() (r []Atom) {
	return
}

// TrackFragRun struct
type TrackFragRun struct {
	Version          uint8
	Flags            uint32
	DataOffset       uint32
	FirstSampleFlags uint32
	Entries          []TrackFragRunEntry
	AtomPos
}

// Marshal func
func (inst TrackFragRun) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(TRUN))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst TrackFragRun) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], inst.Version)
	n++
	pio.PutU24BE(b[n:], inst.Flags)
	n += 3
	pio.PutU32BE(b[n:], uint32(len(inst.Entries)))
	n += 4
	if inst.Flags&TRUN_DATA_OFFSET != 0 {
		{
			pio.PutU32BE(b[n:], inst.DataOffset)
			n += 4
		}
	}
	if inst.Flags&TRUN_FIRST_SAMPLE_FLAGS != 0 {
		{
			pio.PutU32BE(b[n:], inst.FirstSampleFlags)
			n += 4
		}
	}

	for i, entry := range inst.Entries {
		var flags uint32
		if i > 0 {
			flags = inst.Flags
		} else {
			flags = inst.FirstSampleFlags
		}
		if flags&TRUN_SAMPLE_DURATION != 0 {
			pio.PutU32BE(b[n:], entry.Duration)
			n += 4
		}
		if flags&TRUN_SAMPLE_SIZE != 0 {
			pio.PutU32BE(b[n:], entry.Size)
			n += 4
		}
		if flags&TRUN_SAMPLE_FLAGS != 0 {
			pio.PutU32BE(b[n:], entry.Flags)
			n += 4
		}
		if flags&TRUN_SAMPLE_CTS != 0 {
			pio.PutU32BE(b[n:], entry.Cts)
			n += 4
		}
	}
	return
}

// Len func
func (inst TrackFragRun) Len() (n int) {
	n += 8
	n++
	n += 3
	n += 4
	if inst.Flags&TRUN_DATA_OFFSET != 0 {
		{
			n += 4
		}
	}
	if inst.Flags&TRUN_FIRST_SAMPLE_FLAGS != 0 {
		{
			n += 4
		}
	}

	for i := range inst.Entries {
		var flags uint32
		if i > 0 {
			flags = inst.Flags
		} else {
			flags = inst.FirstSampleFlags
		}
		if flags&TRUN_SAMPLE_DURATION != 0 {
			n += 4
		}
		if flags&TRUN_SAMPLE_SIZE != 0 {
			n += 4
		}
		if flags&TRUN_SAMPLE_FLAGS != 0 {
			n += 4
		}
		if flags&TRUN_SAMPLE_CTS != 0 {
			n += 4
		}
	}
	return
}

// Unmarshal func
func (inst *TrackFragRun) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	if len(b) < n+1 {
		err = parseErr("Version", n+offset, err)
		return
	}
	inst.Version = pio.U8(b[n:])
	n++
	if len(b) < n+3 {
		err = parseErr("Flags", n+offset, err)
		return
	}
	inst.Flags = pio.U24BE(b[n:])
	n += 3
	var lenEntries uint32
	lenEntries = pio.U32BE(b[n:])
	n += 4
	inst.Entries = make([]TrackFragRunEntry, lenEntries)
	if inst.Flags&TRUN_DATA_OFFSET != 0 {
		{
			if len(b) < n+4 {
				err = parseErr("DataOffset", n+offset, err)
				return
			}
			inst.DataOffset = pio.U32BE(b[n:])
			n += 4
		}
	}
	if inst.Flags&TRUN_FIRST_SAMPLE_FLAGS != 0 {
		{
			if len(b) < n+4 {
				err = parseErr("FirstSampleFlags", n+offset, err)
				return
			}
			inst.FirstSampleFlags = pio.U32BE(b[n:])
			n += 4
		}
	}

	for i := 0; i < int(lenEntries); i++ {
		var flags uint32
		if i > 0 {
			flags = inst.Flags
		} else {
			flags = inst.FirstSampleFlags
		}
		entry := &inst.Entries[i]
		if flags&TRUN_SAMPLE_DURATION != 0 {
			entry.Duration = pio.U32BE(b[n:])
			n += 4
		}
		if flags&TRUN_SAMPLE_SIZE != 0 {
			entry.Size = pio.U32BE(b[n:])
			n += 4
		}
		if flags&TRUN_SAMPLE_FLAGS != 0 {
			entry.Flags = pio.U32BE(b[n:])
			n += 4
		}
		if flags&TRUN_SAMPLE_CTS != 0 {
			entry.Cts = pio.U32BE(b[n:])
			n += 4
		}
	}
	return
}

// Children func
func (inst TrackFragRun) Children() (r []Atom) {
	return
}

// TrackFragRunEntry struct
type TrackFragRunEntry struct {
	Duration uint32
	Size     uint32
	Flags    uint32
	Cts      uint32
}

// GetTrackFragRunEntry func
func GetTrackFragRunEntry(b []byte) (inst TrackFragRunEntry) {
	inst.Duration = pio.U32BE(b[0:])
	inst.Size = pio.U32BE(b[4:])
	inst.Flags = pio.U32BE(b[8:])
	inst.Cts = pio.U32BE(b[12:])
	return
}

// PutTrackFragRunEntry func
func PutTrackFragRunEntry(b []byte, inst TrackFragRunEntry) {
	pio.PutU32BE(b[0:], inst.Duration)
	pio.PutU32BE(b[4:], inst.Size)
	pio.PutU32BE(b[8:], inst.Flags)
	pio.PutU32BE(b[12:], inst.Cts)
}

// LenTrackFragRunEntry const
const LenTrackFragRunEntry = 16

// TrackFragHeader struct
type TrackFragHeader struct {
	Version         uint8
	Flags           uint32
	BaseDataOffset  uint64
	StsdID          uint32
	DefaultDuration uint32
	DefaultSize     uint32
	DefaultFlags    uint32
	AtomPos
}

// Marshal func
func (inst TrackFragHeader) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(TFHD))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst TrackFragHeader) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], inst.Version)
	n++
	pio.PutU24BE(b[n:], inst.Flags)
	n += 3
	if inst.Flags&TFHD_BASE_DATA_OFFSET != 0 {
		{
			pio.PutU64BE(b[n:], inst.BaseDataOffset)
			n += 8
		}
	}
	if inst.Flags&TFHD_STSD_ID != 0 {
		{
			pio.PutU32BE(b[n:], inst.StsdID)
			n += 4
		}
	}
	if inst.Flags&TFHD_DEFAULT_DURATION != 0 {
		{
			pio.PutU32BE(b[n:], inst.DefaultDuration)
			n += 4
		}
	}
	if inst.Flags&TFHD_DEFAULT_SIZE != 0 {
		{
			pio.PutU32BE(b[n:], inst.DefaultSize)
			n += 4
		}
	}
	if inst.Flags&TFHD_DEFAULT_FLAGS != 0 {
		{
			pio.PutU32BE(b[n:], inst.DefaultFlags)
			n += 4
		}
	}
	return
}

// Len func
func (inst TrackFragHeader) Len() (n int) {
	n += 8
	n++
	n += 3
	if inst.Flags&TFHD_BASE_DATA_OFFSET != 0 {
		{
			n += 8
		}
	}
	if inst.Flags&TFHD_STSD_ID != 0 {
		{
			n += 4
		}
	}
	if inst.Flags&TFHD_DEFAULT_DURATION != 0 {
		{
			n += 4
		}
	}
	if inst.Flags&TFHD_DEFAULT_SIZE != 0 {
		{
			n += 4
		}
	}
	if inst.Flags&TFHD_DEFAULT_FLAGS != 0 {
		{
			n += 4
		}
	}
	return
}

// Unmarshal func
func (inst *TrackFragHeader) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	if len(b) < n+1 {
		err = parseErr("Version", n+offset, err)
		return
	}
	inst.Version = pio.U8(b[n:])
	n++
	if len(b) < n+3 {
		err = parseErr("Flags", n+offset, err)
		return
	}
	inst.Flags = pio.U24BE(b[n:])
	n += 3
	if inst.Flags&TFHD_BASE_DATA_OFFSET != 0 {
		{
			if len(b) < n+8 {
				err = parseErr("BaseDataOffset", n+offset, err)
				return
			}
			inst.BaseDataOffset = pio.U64BE(b[n:])
			n += 8
		}
	}
	if inst.Flags&TFHD_STSD_ID != 0 {
		{
			if len(b) < n+4 {
				err = parseErr("StsdID", n+offset, err)
				return
			}
			inst.StsdID = pio.U32BE(b[n:])
			n += 4
		}
	}
	if inst.Flags&TFHD_DEFAULT_DURATION != 0 {
		{
			if len(b) < n+4 {
				err = parseErr("DefaultDuration", n+offset, err)
				return
			}
			inst.DefaultDuration = pio.U32BE(b[n:])
			n += 4
		}
	}
	if inst.Flags&TFHD_DEFAULT_SIZE != 0 {
		{
			if len(b) < n+4 {
				err = parseErr("DefaultSize", n+offset, err)
				return
			}
			inst.DefaultSize = pio.U32BE(b[n:])
			n += 4
		}
	}
	if inst.Flags&TFHD_DEFAULT_FLAGS != 0 {
		{
			if len(b) < n+4 {
				err = parseErr("DefaultFlags", n+offset, err)
				return
			}
			inst.DefaultFlags = pio.U32BE(b[n:])
			n += 4
		}
	}
	return
}

// Children func
func (inst TrackFragHeader) Children() (r []Atom) {
	return
}

// TrackFragDecodeTime struct
type TrackFragDecodeTime struct {
	Version uint8
	Flags   uint32
	Time    time.Time
	AtomPos
}

// Marshal func
func (inst TrackFragDecodeTime) Marshal(b []byte) (n int) {
	pio.PutU32BE(b[4:], uint32(TFDT))
	n += inst.marshal(b[8:]) + 8
	pio.PutU32BE(b[0:], uint32(n))
	return
}
func (inst TrackFragDecodeTime) marshal(b []byte) (n int) {
	pio.PutU8(b[n:], inst.Version)
	n++
	pio.PutU24BE(b[n:], inst.Flags)
	n += 3
	if inst.Version != 0 {
		PutTime64(b[n:], inst.Time)
		n += 8
	} else {

		PutTime32(b[n:], inst.Time)
		n += 4
	}
	return
}

// Len func
func (inst TrackFragDecodeTime) Len() (n int) {
	n += 8
	n++
	n += 3
	if inst.Version != 0 {
		n += 8
	} else {

		n += 4
	}
	return
}

// Unmarshal func
func (inst *TrackFragDecodeTime) Unmarshal(b []byte, offset int) (n int, err error) {
	(&inst.AtomPos).setPos(offset, len(b))
	n += 8
	if len(b) < n+1 {
		err = parseErr("Version", n+offset, err)
		return
	}
	inst.Version = pio.U8(b[n:])
	n++
	if len(b) < n+3 {
		err = parseErr("Flags", n+offset, err)
		return
	}
	inst.Flags = pio.U24BE(b[n:])
	n += 3
	if inst.Version != 0 {
		inst.Time = GetTime64(b[n:])
		n += 8
	} else {

		inst.Time = GetTime32(b[n:])
		n += 4
	}
	return
}

// Children func
func (inst TrackFragDecodeTime) Children() (r []Atom) {
	return
}
