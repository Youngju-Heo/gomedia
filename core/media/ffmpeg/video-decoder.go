package ffmpeg

/*
#cgo CFLAGS: -I../../../deps/include
#include "ffmpeg.h"

int decode_video(AVCodecContext *ctx, AVFrame *frame, void *data, int size, int *got) {
  struct AVPacket pkt = {.data = data, .size = size};
  int result;
  *got = 0;

  if ((result = avcodec_send_packet(ctx, &pkt)) == 0 && (result = avcodec_receive_frame(ctx, frame)) == 0) {
    *got = 1;
  }

	return result;
}
*/
import "C"
import (
	"fmt"
	"image"
	"reflect"
	"runtime"
	"unsafe"

	"github.com/Youngju-Heo/gomedia/core/media/av"
)

// VideoFrame decoded frame
type VideoFrame struct {
	Image image.YCbCr
	frame *C.AVFrame
}

// Free VideoFrame
func (instance *VideoFrame) Free() {
	instance.Image = image.YCbCr{}
	C.av_frame_free(&instance.frame)
}

func freeVideoFrame(instance *VideoFrame) {
	instance.Free()
}

// VideoDecoder instance
type VideoDecoder struct {
	ff        *ffctx
	Extradata []byte
}

// NewDecoder initialize new decoder
func NewDecoder(codecType av.CodecType) (*VideoDecoder, error) {
	decoder := &VideoDecoder{}

	var id uint32
	var err error

	switch codecType {
	case av.H264:
		id = C.AV_CODEC_ID_H264
	case av.JPEG:
		id = C.AV_CODEC_ID_MJPEG
	case av.HEVC:
		id = C.AV_CODEC_ID_HEVC
	default:
		return nil, fmt.Errorf("invalid codec type: %v", codecType)
	}

	c := C.avcodec_find_decoder(id)
	if c == nil || C.avcodec_get_type(id) != C.AVMEDIA_TYPE_VIDEO {
		return nil, fmt.Errorf("cannot find decoder codecID=%d", id)
	}

	if decoder.ff, err = newFFCtxByCodec(c); err != nil {
		return nil, err
	}

	if err = decoder.Setup(); err != nil {
		return nil, err
	}

	return decoder, nil
}

// Setup initialize VideoDecoder
func (decoder *VideoDecoder) Setup() error {
	ff := &decoder.ff.ff
	if len(decoder.Extradata) > 0 {
		ff.codecCtx.extradata = (*C.uint8_t)(unsafe.Pointer(&decoder.Extradata[0]))
		ff.codecCtx.extradata_size = C.int(len(decoder.Extradata))
	}

	if C.avcodec_open2(ff.codecCtx, ff.codec, nil) != 0 {
		return fmt.Errorf("decoder codec open failed")
	}

	return nil
}

func fromCPtr(buf unsafe.Pointer, size int) (ret []uint8) {
	hdr := (*reflect.SliceHeader)((unsafe.Pointer(&ret)))
	hdr.Cap = size
	hdr.Len = size
	hdr.Data = uintptr(buf)
	return
}

// Decode decode video frame
func (decoder *VideoDecoder) Decode(pkt []byte) (bool, *VideoFrame, error) {
	ff := &decoder.ff.ff
	cgotimg := C.int(0)

	var img *VideoFrame

	frame := C.av_frame_alloc()
	cerr := C.decode_video(ff.codecCtx, frame, unsafe.Pointer(&pkt[0]), C.int(len(pkt)), &cgotimg)
	if cerr < C.int(0) {
		return false, nil, fmt.Errorf("video decode failed: %d", cerr)
	}

	if cgotimg != C.int(0) {
		w := int(frame.width)
		h := int(frame.height)
		ys := int(frame.linesize[0])
		cs := int(frame.linesize[1])

		img = &VideoFrame{Image: image.YCbCr{
			Y:              fromCPtr(unsafe.Pointer(frame.data[0]), ys*h),
			Cb:             fromCPtr(unsafe.Pointer(frame.data[1]), cs*h/2),
			Cr:             fromCPtr(unsafe.Pointer(frame.data[2]), cs*h/2),
			YStride:        ys,
			CStride:        cs,
			SubsampleRatio: image.YCbCrSubsampleRatio420,
			Rect:           image.Rect(0, 0, w, h),
		}, frame: frame}
		runtime.SetFinalizer(img, freeVideoFrame)
	}

	return cgotimg != C.int(0), img, nil
}

// Close close VideoDecoder
func (decoder *VideoDecoder) Close() {

}
