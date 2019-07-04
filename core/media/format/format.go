package format

import (
	"github.com/Youngju-Heo/gomedia/core/media/av/avutil"
	"github.com/Youngju-Heo/gomedia/core/media/format/aac"
	"github.com/Youngju-Heo/gomedia/core/media/format/flv"
	"github.com/Youngju-Heo/gomedia/core/media/format/mp4"
	"github.com/Youngju-Heo/gomedia/core/media/format/rtmp"
	"github.com/Youngju-Heo/gomedia/core/media/format/rtsp"
	"github.com/Youngju-Heo/gomedia/core/media/format/ts"
)

// RegisterAll func
func RegisterAll() {
	avutil.DefaultHandlers.Add(mp4.Handler)
	avutil.DefaultHandlers.Add(ts.Handler)
	avutil.DefaultHandlers.Add(rtmp.Handler)
	avutil.DefaultHandlers.Add(rtsp.Handler)
	avutil.DefaultHandlers.Add(flv.Handler)
	avutil.DefaultHandlers.Add(aac.Handler)
}
