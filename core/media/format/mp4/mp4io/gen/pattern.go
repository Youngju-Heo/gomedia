package main

func moovMovie() {
	atom(Header, MovieHeader)
	atom(MovieExtend, MovieExtend)
	atoms(Tracks, Track)
	_unknowns()
}

func moovMovieHeader() {
	uint8(Version)
	uint24(Flags)
	time32(CreateTime)
	time32(ModifyTime)
	int32(TimeScale)
	int32(Duration)
	fixed32(PreferredRate)
	fixed16(PreferredVolume)
	_skip(10)
	array(Matrix, int32, 9)
	time32(PreviewTime)
	time32(PreviewDuration)
	time32(PosterTime)
	time32(SelectionTime)
	time32(SelectionDuration)
	time32(CurrentTime)
	int32(NextTrackID)
}

func trakTrack() {
	atom(Header, TrackHeader)
	atom(Media, Media)
	_unknowns()
}

func tkhdTrackHeader() {
	uint8(Version)
	uint24(Flags)
	time32(CreateTime)
	time32(ModifyTime)
	int32(TrackID)
	_skip(4)
	int32(Duration)
	_skip(8)
	int16(Layer)
	int16(AlternateGroup)
	fixed16(Volume)
	_skip(2)
	array(Matrix, int32, 9)
	fixed32(TrackWidth)
	fixed32(TrackHeight)
}

func hdlrHandlerRefer() {
	uint8(Version)
	uint24(Flags)
	bytes(Type, 4)
	bytes(SubType, 4)
	bytesleft(Name)
}

func mdiaMedia() {
	atom(Header, MediaHeader)
	atom(Handler, HandlerRefer)
	atom(Info, MediaInfo)
	_unknowns()
}

func mdhdMediaHeader() {
	uint8(Version)
	uint24(Flags)
	time32(CreateTime)
	time32(ModifyTime)
	int32(TimeScale)
	int32(Duration)
	int16(Language)
	int16(Quality)
}

func minfMediaInfo() {
	atom(Sound, SoundMediaInfo)
	atom(Video, VideoMediaInfo)
	atom(Data, DataInfo)
	atom(Sample, SampleTable)
	_unknowns()
}

func dinfDataInfo() {
	atom(Refer, DataRefer)
	_unknowns()
}

func drefDataRefer() {
	uint8(Version)
	uint24(Flags)
	int32(_childrenNR)
	atom(URL, DataReferURL)
}

func urlDataReferURL() {
	uint8(Version)
	uint24(Flags)
}

func smhdSoundMediaInfo() {
	uint8(Version)
	uint24(Flags)
	int16(Balance)
	_skip(2)
}

func vmhdVideoMediaInfo() {
	uint8(Version)
	uint24(Flags)
	int16(GraphicsMode)
	array(Opcolor, int16, 3)
}

func stblSampleTable() {
	atom(SampleDesc, SampleDesc)
	atom(TimeToSample, TimeToSample)
	atom(CompositionOffset, CompositionOffset)
	atom(SampleToChunk, SampleToChunk)
	atom(SyncSample, SyncSample)
	atom(ChunkOffset, ChunkOffset)
	atom(SampleSize, SampleSize)
}

func stsdSampleDesc() {
	uint8(Version)
	_skip(3)
	int32(_childrenNR)
	atom(AVC1Desc, AVC1Desc)
	atom(MP4ADesc, MP4ADesc)
	_unknowns()
}

func mp4aMP4ADesc() {
	_skip(6)
	int16(DataRefIdx)
	int16(Version)
	int16(RevisionLevel)
	int32(Vendor)
	int16(NumberOfChannels)
	int16(SampleSize)
	int16(CompressionID)
	_skip(2)
	fixed32(SampleRate)
	atom(Conf, ElemStreamDesc)
	_unknowns()
}

func avc1AVC1Desc() {
	_skip(6)
	int16(DataRefIdx)
	int16(Version)
	int16(Revision)
	int32(Vendor)
	int32(TemporalQuality)
	int32(SpatialQuality)
	int16(Width)
	int16(Height)
	fixed32(HorizontalResolution)
	fixed32(VorizontalResolution)
	_skip(4)
	int16(FrameCount)
	bytes(CompressorName, 32)
	int16(Depth)
	int16(ColorTableID)
	atom(Conf, AVC1Conf)
	_unknowns()
}

func avcCAVC1Conf() {
	bytesleft(Data)
}

func sttsTimeToSample() {
	uint8(Version)
	uint24(Flags)
	uint32(_len_Entries)
	slice(Entries, TimeToSampleEntry)
}

func TimeToSampleEntry() {
	uint32(Count)
	uint32(Duration)
}

func stscSampleToChunk() {
	uint8(Version)
	uint24(Flags)
	uint32(_len_Entries)
	slice(Entries, SampleToChunkEntry)
}

func SampleToChunkEntry() {
	uint32(FirstChunk)
	uint32(SamplesPerChunk)
	uint32(SampleDescID)
}

func cttsCompositionOffset() {
	uint8(Version)
	uint24(Flags)
	uint32(_len_Entries)
	slice(Entries, CompositionOffsetEntry)
}

func CompositionOffsetEntry() {
	uint32(Count)
	uint32(Offset)
}

func stssSyncSample() {
	uint8(Version)
	uint24(Flags)
	uint32(_len_Entries)
	slice(Entries, uint32)
}

func stcoChunkOffset() {
	uint8(Version)
	uint24(Flags)
	uint32(_len_Entries)
	slice(Entries, uint32)
}

func moofMovieFrag() {
	atom(Header, MovieFragHeader)
	atoms(Tracks, TrackFrag)
	_unknowns()
}

func mfhdMovieFragHeader() {
	uint8(Version)
	uint24(Flags)
	uint32(Seqnum)
}

func trafTrackFrag() {
	atom(Header, TrackFragHeader)
	atom(DecodeTime, TrackFragDecodeTime)
	atom(Run, TrackFragRun)
	_unknowns()
}

func mvexMovieExtend() {
	atoms(Tracks, TrackExtend)
	_unknowns()
}

func trexTrackExtend() {
	uint8(Version)
	uint24(Flags)
	uint32(TrackID)
	uint32(DefaultSampleDescIdx)
	uint32(DefaultSampleDuration)
	uint32(DefaultSampleSize)
	uint32(DefaultSampleFlags)
}

func stszSampleSize() {
	uint8(Version)
	uint24(Flags)
	uint32(SampleSize)
	_code(func() {
		if self.SampleSize != 0 {
			return
		}
	})
	uint32(_len_Entries)
	slice(Entries, uint32)
}

func trunTrackFragRun() {
	uint8(Version)
	uint24(Flags)
	uint32(_len_Entries)

	uint32(DataOffset, _code(func() {
		if self.Flags&TRUN_DATA_OFFSET != 0 {
			doit()
		}
	}))

	uint32(FirstSampleFlags, _code(func() {
		if self.Flags&TRUN_FIRST_SAMPLE_FLAGS != 0 {
			doit()
		}
	}))

	slice(Entries, TrackFragRunEntry, _code(func() {
		for i, entry := range self.Entries {
			var flags uint32
			if i > 0 {
				flags = self.Flags
			} else {
				flags = self.FirstSampleFlags
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
	}, func() {
		for i := range self.Entries {
			var flags uint32
			if i > 0 {
				flags = self.Flags
			} else {
				flags = self.FirstSampleFlags
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
	}, func() {
		for i := 0; i < int(_len_Entries); i++ {
			var flags uint32
			if i > 0 {
				flags = self.Flags
			} else {
				flags = self.FirstSampleFlags
			}
			entry := &self.Entries[i]
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
	}))
}

func TrackFragRunEntry() {
	uint32(Duration)
	uint32(Size)
	uint32(Flags)
	uint32(Cts)
}

func tfhdTrackFragHeader() {
	uint8(Version)
	uint24(Flags)

	uint64(BaseDataOffset, _code(func() {
		if self.Flags&TFHD_BASE_DATA_OFFSET != 0 {
			doit()
		}
	}))

	uint32(StsdID, _code(func() {
		if self.Flags&TFHD_STSD_ID != 0 {
			doit()
		}
	}))

	uint32(DefaultDuration, _code(func() {
		if self.Flags&TFHD_DEFAULT_DURATION != 0 {
			doit()
		}
	}))

	uint32(DefaultSize, _code(func() {
		if self.Flags&TFHD_DEFAULT_SIZE != 0 {
			doit()
		}
	}))

	uint32(DefaultFlags, _code(func() {
		if self.Flags&TFHD_DEFAULT_FLAGS != 0 {
			doit()
		}
	}))
}

func tfdtTrackFragDecodeTime() {
	uint8(Version)
	uint24(Flags)
	time64(Time, _code(func() {
		if self.Version != 0 {
			PutTime64(b[n:], self.Time)
			n += 8
		} else {
			PutTime32(b[n:], self.Time)
			n += 4
		}
	}, func() {
		if self.Version != 0 {
			n += 8
		} else {
			n += 4
		}
	}, func() {
		if self.Version != 0 {
			self.Time = GetTime64(b[n:])
			n += 8
		} else {
			self.Time = GetTime32(b[n:])
			n += 4
		}
	}))
}
