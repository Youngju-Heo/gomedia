package format

import (
	"../av/avutil"
	"../format/aac"
	"../format/flv"
	"../format/mp4"
	"../format/rtmp"
	"../format/rtsp"
	"../format/ts"
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
