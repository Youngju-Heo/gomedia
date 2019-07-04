package device

/*
#cgo CFLAGS: -I../../deps/include
#cgo LDFLAGS: -L../../deps/lib
#cgo linux,386 LDFLAGS: -lSDL2_linux_386 -Wl,--no-undefined -lm -ldl -lasound -lm -ldl -lpthread -lX11 -lXext -lXcursor -lXinerama -lXi -lXrandr -lXss -lXxf86vm -lpthread -lrt
#cgo linux,amd64 LDFLAGS: -lSDL2_linux_amd64 -Wl,--no-undefined -lm -ldl -lasound -lm -ldl -lpthread -lX11 -lXext -lXcursor -lXinerama -lXi -lXrandr -lXss -lXxf86vm -lpthread -lrt
#cgo windows,386 LDFLAGS: -lSDL2_windows_386 -lSDL2main_windows_386 -mwindows -Wl,--no-undefined -lm -ldinput8 -ldxguid -ldxerr8 -luser32 -lgdi32 -lwinmm -limm32 -lole32 -loleaut32 -lshell32 -lsetupapi -lversion -luuid -static-libgcc
#cgo windows,amd64 LDFLAGS: -lSDL2_windows_amd64 -lSDL2main_windows_amd64 -mwindows -Wl,--no-undefined -lm -ldinput8 -ldxguid -ldxerr8 -luser32 -lgdi32 -lwinmm -limm32 -lole32 -loleaut32 -lshell32 -lversion -luuid -lsetupapi -static-libgcc
#cgo darwin,amd64 LDFLAGS: -lSDL2_darwin_amd64 -lm -liconv -Wl,-framework,CoreAudio -Wl,-framework,AudioToolbox -Wl,-framework,ForceFeedback -lobjc -Wl,-framework,CoreVideo -Wl,-framework,Cocoa -Wl,-framework,Carbon -Wl,-framework,IOKit -Wl,-framework,Metal
#cgo android,arm LDFLAGS: -lSDL2_android_arm -Wl,--no-undefined -lm -ldl -llog -landroid -lGLESv2
#cgo linux,arm,!android LDFLAGS: -L/opt/vc/lib -L/opt/vc/lib64 -lSDL2_linux_arm -Wl,--no-undefined -lm -ldl -liconv -lbcm_host -lvcos -lvchiq_arm -pthread

#include <stdint.h>
#include <stdio.h>
#include <SDL2/SDL.h>

extern void goAudioCallback(int param, uint8_t* stream, int len);

static void inlineAudioCallback(void* param, uint8_t* buffer, int len) {
  uint64_t intParam = (uint64_t)param;
  goAudioCallback((int32_t)intParam, buffer, len);
}

static int initializeAudio(int param, int freq, int format, int channels, int samples) {
  uint64_t paramInternal = param;
  SDL_AudioSpec spec;

  printf("initializeAudio(param=%d,freq=%d,format=%d,channels=%d,samples=%d)\n",
  param,
  freq,
  format,
  channels,
  samples);

  spec.freq     = freq;
  spec.format   = format;
  spec.channels = channels;
  spec.silence  = 0;
  spec.samples  = samples;
  spec.padding  = 0;
  spec.size     = 0;
  spec.callback = inlineAudioCallback;
  spec.userdata = (void*)paramInternal;

  return SDL_OpenAudio(&spec, NULL);
}


*/
import "C"
import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/Youngju-Heo/gomedia/core/common"
	"github.com/Youngju-Heo/gomedia/core/media/av"

	"github.com/Youngju-Heo/gomedia/sdl2/sdl"
)

//export goAudioCallback
func goAudioCallback(param C.int, stream *C.uint8_t, len C.int) {
	v := common.RestorePointer(int(param))
	if v != nil {
		n := int(len)
		hdr := reflect.SliceHeader{Data: uintptr(unsafe.Pointer(stream)), Len: n, Cap: n}
		buf := *(*[]C.uint8_t)(unsafe.Pointer(&hdr))
		processor := v.(AudioProcessor)
		rslts := processor.Process(int(len))
		for idx, val := range rslts {
			buf[idx] = C.uint8_t(val)
		}

	}
}

func sampleAudioFormatToSdlFormat(format av.SampleFormat) (fmt int) {
	switch format {
	case av.S16:
		fmt = sdl.AUDIO_S16
	}

	return
}

// AudioProcessor interface
type AudioProcessor interface {
	Process(len int) []byte
	Frequency() int
	Format() av.SampleFormat
	Channels() av.ChannelLayout
	RecommendSampleCount() int
}

func (service *Service) openAudio() (err error) {
	if service.flags&sdl.INIT_AUDIO == 0 {
		return
	}

	idx := common.SavePointer(service.audioContext)
	sdlFormat := sampleAudioFormatToSdlFormat(service.audioContext.Format())
	if ret := C.initializeAudio(C.int(idx),
		C.int(service.audioContext.Frequency()),
		C.int(sdlFormat),
		C.int(service.audioContext.Channels().Count()),
		C.int(service.audioContext.RecommendSampleCount())); ret != C.int(0) {
		return fmt.Errorf("init audio failed:%v", ret)
	}
	return
}

func (service *Service) closeAudio() {
	if service.flags&sdl.INIT_AUDIO != 0 {
		sdl.PauseAudio(true)
		sdl.CloseAudio()
	}
}
