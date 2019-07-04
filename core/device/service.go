package device

/*

 */
import "C"

import (
	"log"
	"sync"

	"github.com/Youngju-Heo/gomedia/sdl2/sdl"
)

// AudioFormat audio format define
type AudioFormat int

// Flag Define
const (
	InitTimer uint32 = (1 << iota)
	InitVideo
	InitAudio
	InitEvent
)

// Service sdl device context
type Service struct {
	sync.Mutex
	flags        uint32
	isContinue   bool
	audioContext AudioProcessor
}

// Initialize initialize service
func (service *Service) Initialize(capability uint32,
	audioContext AudioProcessor) {
	var capa uint32
	if capability&InitTimer != 0 {
		capa |= sdl.INIT_TIMER
	}
	if capability&InitAudio != 0 {
		capa |= sdl.INIT_AUDIO
	}
	if capability&InitVideo != 0 {
		capa |= sdl.INIT_VIDEO
	}
	if capability&InitEvent != 0 {
		capa |= sdl.INIT_EVENTS
	}

	service.flags = capa
	service.audioContext = audioContext
	service.isContinue = true
}

// Status read service status
func (service *Service) Status() bool {
	service.Lock()
	defer service.Unlock()

	return service.isContinue
}

// Start start service
func (service *Service) Start() (err error) {

	service.Lock()
	service.isContinue = true
	service.Unlock()
	if err = sdl.Init(service.flags); err != nil {
		log.Println("error while initialize device", sdl.GetError())
		return
	}

	defer sdl.Quit()

	// set audio if needed
	if err = service.openAudio(); err != nil {
		return
	}

	if service.flags&sdl.INIT_AUDIO != 0 {
		sdl.PauseAudio(false)
	}

	for service.Status() {
		sdl.Delay(50)
	}

	service.closeAudio()

	return
}

// Stop stop service
func (service *Service) Stop() {
	service.Lock()
	defer service.Unlock()
	service.isContinue = false
}
