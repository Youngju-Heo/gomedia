# SDL2 and ffmpeg media related library package for go language

# dependent
* https://github.com/golang/freetype/raster

# code reference(forked from)
* https://github.com/veandco/go-sdl2
* https://github.com/nareix/joy4

# ffmpeg compile
* compile
    ```
    ./configure --prefix=/d/work/ffmpeg/deps \
    --enable-static \
    --disable-shared \
    --disable-dxva2 \
    --disable-everything \
    --enable-decoder=aac,g723_1,g729,gif,h263,h264,hevc,mjpeg,mp3,mpeg4,mpegvideo,opus,pcm_alaw,pcm_mulaw,pcm_s16be,pcm_s16be_planar,pcm_s16le,pcm_s16le_planar,srt,movtext,subrip,text,vp8,vp9 \
    --enable-encoder=aac,gif,h263,mjpeg,mp3,mpeg4,opus,pcm_alaw,pcm_mulaw,pcm_s16be,pcm_s16be_planar,pcm_s16le,pcm_s16le_planar,srt,movtext,subbrip,text \
    --enable-parser=aac,h264,mpegaudio,mpegvideo,vp9,h263,mpeg4video \
    --enable-demuxer=aac,h264,m4v,mp4,mjpeg,mp3,gif,hevc,matroska,mov,mpegvideo \
    --enable-muxer=mpegvideo,mp3,mp4,gif,mov,matroska,webm,h264,mjpeg,mulaw,alaw,opus,aac \
    --enable-protocol=file \
    --enable-filter=scale,fps,copy,palettegen,vflip,paletteuse,crop
    
    make -j 8 install
    ```
