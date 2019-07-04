package ffmpeg

/*
#cgo CFLAGS: -I../../../deps/include
#cgo darwin,amd64 LDFLAGS: -L../../../deps/lib -lavcodec_darwin_amd64 -lavformat_darwin_amd64 -lswresample_darwin_amd64 -lswscale_darwin_amd64 -lavutil_darwin_amd64 -lmp3lame_darwin_amd64 -lSDL2_darwin_amd64 -liconv -framework CoreVideo
#cgo linux,amd64 LDFLAGS: -L../../../deps/lib -lavcodec -lavformat -lswresample -lswscale -lavutil -lmp3lame -lSDL2_linux_amd64
#cgo windows,amd64 LDFLAGS: -L../../../deps/lib -lavcodec_windows_amd64 -lavformat_windows_amd64 -lswresample_windows_amd64 -lswscale_windows_amd64 -lavutil_windows_amd64 -lmp3lame_windows_amd64 -lSDL2_windows_amd64 -liconv -lbcrypt
#include "ffmpeg.h"

static char msgBuffer[128];
static const char* get_error_msg(int code) {
  av_make_error_string(msgBuffer, 128, code);
  return msgBuffer;
}

*/
import "C"
import (
	"runtime"
	"sync"
	"unsafe"
)

const (
	// QUIET mode
	QUIET = int(C.AV_LOG_QUIET)
	// PANIC mode
	PANIC = int(C.AV_LOG_PANIC)
	// FATAL mode
	FATAL = int(C.AV_LOG_FATAL)
	// ERROR mode
	ERROR = int(C.AV_LOG_ERROR)
	// WARNING mode
	WARNING = int(C.AV_LOG_WARNING)
	// INFO mode
	INFO = int(C.AV_LOG_INFO)
	// VERBOSE mode
	VERBOSE = int(C.AV_LOG_VERBOSE)
	// DEBUG mode
	DEBUG = int(C.AV_LOG_DEBUG)
	// TRACE mode
	TRACE = int(C.AV_LOG_TRACE)
)

// HasEncoder func
func HasEncoder(name string) bool {

	return C.avcodec_find_encoder_by_name(C.CString(name)) != nil
}

// HasDecoder func
func HasDecoder(name string) bool {
	return C.avcodec_find_decoder_by_name(C.CString(name)) != nil
}

//func EncodersList() []string
//func DecodersList() []string

// SetLogLevel func
func SetLogLevel(level int) {
	C.av_log_set_level(C.int(level))
}

type ffctx struct {
	ff C.FFCtx
}

func newFFCtxByCodec(codec *C.AVCodec) (ff *ffctx, err error) {
	ff = &ffctx{}
	ff.ff.codec = codec
	ff.ff.codecCtx = C.avcodec_alloc_context3(codec)
	ff.ff.profile = C.FF_PROFILE_UNKNOWN
	runtime.SetFinalizer(ff, freeFFCtx)
	return
}

func freeFFCtx(self *ffctx) {
	ff := &self.ff
	if ff.frame != nil {
		C.av_frame_free(&ff.frame)
		ff.frame = nil
	}
	if ff.codecCtx != nil {
		C.avcodec_close(ff.codecCtx)
		C.av_free(unsafe.Pointer(ff.codecCtx))
		ff.codecCtx = nil
	}
	if ff.options != nil {
		C.av_dict_free(&ff.options)
		ff.options = nil
	}
}

var logMutex sync.Mutex

// GetFFErrorMessage get ffmpeg error message by code
func GetFFErrorMessage(code int) (msg string) {
	logMutex.Lock()
	defer logMutex.Unlock()
	msg = C.GoString(C.get_error_msg(C.int(code)))
	return
}
