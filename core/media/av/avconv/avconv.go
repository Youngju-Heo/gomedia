package avconv

import (
	"fmt"
	"io"
	"time"

	"github.com/Youngju-Heo/gomedia/core/media/av"
	"github.com/Youngju-Heo/gomedia/core/media/av/avutil"
	"github.com/Youngju-Heo/gomedia/core/media/av/pktque"
	"github.com/Youngju-Heo/gomedia/core/media/av/transcode"
)

// Debug debug status
var Debug bool

// Option struct
type Option struct {
	Transcode bool
	Args      []string
}

// Options struct
type Options struct {
	OutputCodecTypes []av.CodecType
}

// Demuxer struct
type Demuxer struct {
	transdemux *transcode.Demuxer
	streams    []av.CodecData
	Options
	Demuxer av.Demuxer
}

// Close method
func (mux *Demuxer) Close() (err error) {
	if mux.transdemux != nil {
		return mux.transdemux.Close()
	}
	return
}

// Streams method
func (mux *Demuxer) Streams() (streams []av.CodecData, err error) {
	if err = mux.prepare(); err != nil {
		return
	}
	streams = mux.streams
	return
}

// ReadPacket read data packet
func (mux *Demuxer) ReadPacket() (pkt av.Packet, err error) {
	if err = mux.prepare(); err != nil {
		return
	}
	return mux.transdemux.ReadPacket()
}

func (mux *Demuxer) prepare() (err error) {
	if mux.transdemux != nil {
		return
	}

	/*
		var streams []av.CodecData
		if streams, err = mux.Demuxer.Streams(); err != nil {
			return
		}
	*/

	supports := mux.Options.OutputCodecTypes

	transopts := transcode.Options{}
	transopts.FindAudioDecoderEncoder = func(codec av.AudioCodecData, i int) (ok bool, dec av.AudioDecoder, enc av.AudioEncoder, err error) {
		if len(supports) == 0 {
			return
		}

		support := false
		for _, typ := range supports {
			if typ == codec.Type() {
				support = true
			}
		}

		if support {
			return
		}
		ok = true

		var enctype av.CodecType
		for _, typ := range supports {
			if typ.IsAudio() {
				if enc, _ = avutil.DefaultHandlers.NewAudioEncoder(typ); enc != nil {
					enctype = typ
					break
				}
			}
		}
		if enc == nil {
			err = fmt.Errorf("avconv: convert %s->%s failed", codec.Type(), enctype)
			return
		}

		// TODO: support per stream option
		// enc.SetSampleRate ...

		if dec, err = avutil.DefaultHandlers.NewAudioDecoderParam(codec); err != nil {
			err = fmt.Errorf("avconv: decode %s failed", codec.Type())
			return
		}

		return
	}

	mux.transdemux = &transcode.Demuxer{
		Options: transopts,
		Demuxer: mux.Demuxer,
	}
	if mux.streams, err = mux.transdemux.Streams(); err != nil {
		return
	}

	return
}

// ConvertCmdline cmdline parser
func ConvertCmdline(args []string) (err error) {
	output := ""
	input := ""
	flagi := false
	flagv := false
	flagt := false
	flagre := false
	duration := time.Duration(0)
	options := Options{}

	for _, arg := range args {
		switch arg {
		case "-i":
			flagi = true

		case "-v":
			flagv = true

		case "-t":
			flagt = true

		case "-re":
			flagre = true

		default:
			switch {
			case flagi:
				flagi = false
				input = arg

			case flagt:
				flagt = false
				var f float64
				fmt.Sscanf(arg, "%f", &f)
				duration = time.Duration(f * float64(time.Second))

			default:
				output = arg
			}
		}
	}

	if input == "" {
		err = fmt.Errorf("avconv: input file not specified")
		return
	}

	if output == "" {
		err = fmt.Errorf("avconv: output file not specified")
		return
	}

	var demuxer av.DemuxCloser
	var muxer av.MuxCloser

	if demuxer, err = avutil.Open(input); err != nil {
		return
	}
	defer demuxer.Close()

	var handler avutil.RegisterHandler
	if handler, muxer, err = avutil.DefaultHandlers.FindCreate(output); err != nil {
		return
	}
	defer muxer.Close()

	options.OutputCodecTypes = handler.CodecTypes

	convdemux := &Demuxer{
		Options: options,
		Demuxer: demuxer,
	}
	defer convdemux.Close()

	var streams []av.CodecData
	if streams, err = demuxer.Streams(); err != nil {
		return
	}

	var convstreams []av.CodecData
	if convstreams, err = convdemux.Streams(); err != nil {
		return
	}

	if flagv {
		for _, stream := range streams {
			fmt.Print(stream.Type(), " ")
		}
		fmt.Print("-> ")
		for _, stream := range convstreams {
			fmt.Print(stream.Type(), " ")
		}
		fmt.Println()
	}

	if err = muxer.WriteHeader(convstreams); err != nil {
		return
	}

	filters := pktque.Filters{}
	if flagre {
		filters = append(filters, &pktque.Walltime{})
	}
	filterdemux := &pktque.FilterDemuxer{
		Demuxer: convdemux,
		Filter:  filters,
	}

	for {
		var pkt av.Packet
		if pkt, err = filterdemux.ReadPacket(); err != nil {
			if err == io.EOF {
				err = nil
				break
			}
			return
		}
		if flagv {
			fmt.Println(pkt.Idx, pkt.Time, len(pkt.Data), pkt.IsKeyFrame)
		}
		if duration != 0 && pkt.Time > duration {
			break
		}
		if err = muxer.WritePacket(pkt); err != nil {
			return
		}
	}

	if err = muxer.WriteTrailer(); err != nil {
		return
	}

	return
}
