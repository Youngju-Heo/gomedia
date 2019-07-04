package avutil

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/Youngju-Heo/gomedia/core/media/av"
)

// HandlerDemuxer struct
type HandlerDemuxer struct {
	av.Demuxer
	r io.ReadCloser
}

// Close method
func (hndl *HandlerDemuxer) Close() error {
	return hndl.r.Close()
}

// HandlerMuxer struct
type HandlerMuxer struct {
	av.Muxer
	w     io.WriteCloser
	stage int
}

// WriteHeader func
func (hndl *HandlerMuxer) WriteHeader(streams []av.CodecData) (err error) {
	if hndl.stage == 0 {
		if err = hndl.Muxer.WriteHeader(streams); err != nil {
			return
		}
		hndl.stage++
	}
	return
}

// WriteTrailer func
func (hndl *HandlerMuxer) WriteTrailer() (err error) {
	if hndl.stage == 1 {
		hndl.stage++
		if err = hndl.Muxer.WriteTrailer(); err != nil {
			return
		}
	}
	return
}

// Close func
func (hndl *HandlerMuxer) Close() (err error) {
	if err = hndl.WriteTrailer(); err != nil {
		return
	}
	return hndl.w.Close()
}

// RegisterHandler struct
type RegisterHandler struct {
	Ext           string
	ReaderDemuxer func(io.Reader) av.Demuxer
	WriterMuxer   func(io.Writer) av.Muxer
	URLMuxer      func(string) (bool, av.MuxCloser, error)
	URLDemuxer    func(string) (bool, av.DemuxCloser, error)
	URLReader     func(string) (bool, io.ReadCloser, error)
	Probe         func([]byte) bool
	AudioEncoder  func(av.CodecType) (av.AudioEncoder, error)
	AudioDecoder  func(av.AudioCodecData) (av.AudioDecoder, error)
	ServerDemuxer func(string) (bool, av.DemuxCloser, error)
	ServerMuxer   func(string) (bool, av.MuxCloser, error)
	CodecTypes    []av.CodecType
}

// Handlers struct
type Handlers struct {
	handlers []RegisterHandler
}

// Add func
func (hndl *Handlers) Add(fn func(*RegisterHandler)) {
	handler := &RegisterHandler{}
	fn(handler)
	hndl.handlers = append(hndl.handlers, *handler)
}

func (hndl *Handlers) openURL(u *url.URL, uri string) (r io.ReadCloser, err error) {
	if u != nil && (u.Scheme != "" && u.Scheme != "file") {
		for _, handler := range hndl.handlers {
			if handler.URLReader != nil {
				var ok bool
				if ok, r, err = handler.URLReader(uri); ok {
					return
				}
			}
		}
		err = fmt.Errorf("avutil: openURL %s failed", uri)
	} else {
		if strings.HasPrefix(uri, "file:///") {
			uri = uri[7:]
			if match, _ := regexp.MatchString("^/[a-zA-z]:.*", uri); match {
				uri = uri[1:]
			}
		}
		r, err = os.Open(uri)
	}
	return
}

func (hndl *Handlers) createURL(u *url.URL, uri string) (w io.WriteCloser, err error) {
	w, err = os.Create(uri)
	return
}

// NewAudioEncoder NewAudioEncoder
func (hndl *Handlers) NewAudioEncoder(typ av.CodecType) (enc av.AudioEncoder, err error) {
	for _, handler := range hndl.handlers {
		if handler.AudioEncoder != nil {
			if enc, _ = handler.AudioEncoder(typ); enc != nil {
				return
			}
		}
	}
	err = fmt.Errorf("avutil: encoder %s %s", typ, "not found")
	return
}

// NewAudioDecoderParam NewAudioDecoderParam
func (hndl *Handlers) NewAudioDecoderParam(codec av.AudioCodecData) (dec av.AudioDecoder, err error) {
	for _, handler := range hndl.handlers {
		if handler.AudioDecoder != nil {
			if dec, _ = handler.AudioDecoder(codec); dec != nil {
				return
			}
		}
	}
	err = fmt.Errorf("avutil: decoder %d not found", codec.Type())
	return
}

// Open Open
func (hndl *Handlers) Open(uri string) (demuxer av.DemuxCloser, err error) {
	listen := false
	if strings.HasPrefix(uri, "listen:") {
		uri = uri[len("listen:"):]
		listen = true
	}

	for _, handler := range hndl.handlers {
		if listen {
			if handler.ServerDemuxer != nil {
				var ok bool
				if ok, demuxer, err = handler.ServerDemuxer(uri); ok {
					return
				}
			}
		} else {
			if handler.URLDemuxer != nil {
				var ok bool
				if ok, demuxer, err = handler.URLDemuxer(uri); ok {
					return
				}
			}
		}
	}

	var r io.ReadCloser
	var ext string
	var u *url.URL
	if u, _ = url.Parse(uri); u != nil && u.Scheme != "" {
		ext = path.Ext(u.Path)
	} else {
		ext = path.Ext(uri)
	}

	if ext != "" {
		for _, handler := range hndl.handlers {
			if handler.Ext == ext {
				if handler.ReaderDemuxer != nil {
					if r, err = hndl.openURL(u, uri); err != nil {
						return
					}
					demuxer = &HandlerDemuxer{
						Demuxer: handler.ReaderDemuxer(r),
						r:       r,
					}
					return
				}
			}
		}
	}

	var probebuf [1024]byte
	if r, err = hndl.openURL(u, uri); err != nil {
		return
	}
	if _, err = io.ReadFull(r, probebuf[:]); err != nil {
		return
	}

	for _, handler := range hndl.handlers {
		if handler.Probe != nil && handler.Probe(probebuf[:]) && handler.ReaderDemuxer != nil {
			var _r io.Reader
			if rs, ok := r.(io.ReadSeeker); ok {
				if _, err = rs.Seek(0, 0); err != nil {
					return
				}
				_r = rs
			} else {
				_r = io.MultiReader(bytes.NewReader(probebuf[:]), r)
			}
			demuxer = &HandlerDemuxer{
				Demuxer: handler.ReaderDemuxer(_r),
				r:       r,
			}
			return
		}
	}

	r.Close()
	err = fmt.Errorf("avutil: open %s failed", uri)
	return
}

// Create Create
func (hndl *Handlers) Create(uri string) (muxer av.MuxCloser, err error) {
	_, muxer, err = hndl.FindCreate(uri)
	return
}

// FindCreate FindCreate
func (hndl *Handlers) FindCreate(uri string) (handler RegisterHandler, muxer av.MuxCloser, err error) {
	listen := false
	if strings.HasPrefix(uri, "listen:") {
		uri = uri[len("listen:"):]
		listen = true
	}

	for _, handler = range hndl.handlers {
		if listen {
			if handler.ServerMuxer != nil {
				var ok bool
				if ok, muxer, err = handler.ServerMuxer(uri); ok {
					return
				}
			}
		} else {
			if handler.URLMuxer != nil {
				var ok bool
				if ok, muxer, err = handler.URLMuxer(uri); ok {
					return
				}
			}
		}
	}

	var ext string
	var u *url.URL
	if u, _ = url.Parse(uri); u != nil && u.Scheme != "" {
		ext = path.Ext(u.Path)
	} else {
		ext = path.Ext(uri)
	}

	if ext != "" {
		for _, handler = range hndl.handlers {
			if handler.Ext == ext && handler.WriterMuxer != nil {
				var w io.WriteCloser
				if w, err = hndl.createURL(u, uri); err != nil {
					return
				}
				muxer = &HandlerMuxer{
					Muxer: handler.WriterMuxer(w),
					w:     w,
				}
				return
			}
		}
	}

	err = fmt.Errorf("avutil: create muxer %s failed", uri)
	return
}

// DefaultHandlers default
var DefaultHandlers = &Handlers{}

// Open Open
func Open(url string) (demuxer av.DemuxCloser, err error) {
	return DefaultHandlers.Open(url)
}

// Create Create
func Create(url string) (muxer av.MuxCloser, err error) {
	return DefaultHandlers.Create(url)
}

// CopyPackets CopyPackets
func CopyPackets(dst av.PacketWriter, src av.PacketReader) (err error) {
	for {
		var pkt av.Packet
		if pkt, err = src.ReadPacket(); err != nil {
			if err == io.EOF {
				break
			}
			return
		}
		if err = dst.WritePacket(pkt); err != nil {
			return
		}
	}
	return
}

// CopyFile CopyFile
func CopyFile(dst av.Muxer, src av.Demuxer) (err error) {
	var streams []av.CodecData
	if streams, err = src.Streams(); err != nil {
		return
	}
	if err = dst.WriteHeader(streams); err != nil {
		return
	}
	if err = CopyPackets(dst, src); err != nil {
		if err != io.EOF {
			return
		}
	}
	if err = dst.WriteTrailer(); err != nil {
		return
	}
	return
}
