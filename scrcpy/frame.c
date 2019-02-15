#include <libavformat/avformat.h>
#include <libavcodec/avcodec.h>
#include <libavcodec/videotoolbox.h>
#include <libavutil/avutil.h>
#include <stdio.h>

#define BUFSIZE 0x10000

int goReadPacket(void* p, void *buf, int size);
void goPushFrame();
void goNotifyStopped();
AVFrame* goGetHardwareFrame();
int goAvHwframeTransferData();

static int read_packet(void *opaque, uint8_t *buf, int buf_size) {
    return goReadPacket(opaque, (void *) buf, buf_size);
}

static enum AVPixelFormat get_hw_format(AVCodecContext *ctx,
                                        const enum AVPixelFormat *pix_fmts)
{
    const enum AVPixelFormat *p;

    for (p = pix_fmts; *p != -1; p++) {
        if (*p == AV_PIX_FMT_VIDEOTOOLBOX)
            return *p;
    }

    fprintf(stderr, "Failed to get HW surface format\n");
    return AV_PIX_FMT_NONE;
}

static int hw_decoder_init(AVCodecContext *ctx,
        AVBufferRef **hw_device_ctx,
        const enum AVHWDeviceType type)
{
    int err = 0;
    if ((err = av_hwdevice_ctx_create(hw_device_ctx, type,
                                      NULL, NULL, 0)) < 0) {
        return err;
    }
    ctx->hw_device_ctx = av_buffer_ref(*hw_device_ctx);

    return err;
}

int run_decoder() {
    AVCodec *codec = avcodec_find_decoder(AV_CODEC_ID_H264);
    if (!codec) {
        fprintf(stderr, "H.264 decoder not found\n");
        goto run_end;
    }

    AVCodecContext *codec_ctx = avcodec_alloc_context3(codec);
    if (!codec_ctx) {
        fprintf(stderr, "Could not allocate decoder context\n");
        goto run_end;
    }

    enum AVHWDeviceType type = AV_HWDEVICE_TYPE_VIDEOTOOLBOX;
    AVBufferRef *hw_device_ctx = NULL;
    if (hw_decoder_init(codec_ctx, &hw_device_ctx, type) < 0) {
        fprintf(stderr, "Failed to create specified HW device\n");
        goto run_end;
    }
    codec_ctx->hw_device_ctx = hw_device_ctx;
    codec_ctx->get_format = get_hw_format;

    if (avcodec_open2(codec_ctx, codec, NULL) < 0) {
        fprintf(stderr, "Could not open H.264 codec\n");
        goto run_finally_free_codec_ctx;
    }

    AVFormatContext *format_ctx = avformat_alloc_context();
    if (!format_ctx) {
        fprintf(stderr, "Could not allocate format context\n");
        goto run_finally_close_codec;
    }

    unsigned char *buffer = av_malloc(BUFSIZE);
    if (!buffer) {
        fprintf(stderr, "Could not allocate buffer\n");
        goto run_finally_free_format_ctx;
    }

    AVIOContext *avio_ctx = avio_alloc_context(buffer, BUFSIZE, 0, NULL, read_packet, NULL, NULL);
    if (!avio_ctx) {
        fprintf(stderr, "Could not allocate avio context\n");
        // avformat_open_input takes ownership of 'buffer'
        // so only free the buffer before avformat_open_input()
        av_free(buffer);
        goto run_finally_free_format_ctx;
    }

    format_ctx->pb = avio_ctx;

    if (avformat_open_input(&format_ctx, NULL, NULL, NULL) < 0) {
        fprintf(stderr, "Could not open video stream\n");
        goto run_finally_free_avio_ctx;
    }

    AVPacket packet;
    av_init_packet(&packet);
    packet.data = NULL;
    packet.size = 0;

    while (!av_read_frame(format_ctx, &packet)) {
// the new decoding/encoding API has been introduced by:
// <http://git.videolan.org/?p=ffmpeg.git;a=commitdiff;h=7fc329e2dd6226dfecaa4a1d7adf353bf2773726>
#if LIBAVCODEC_VERSION_INT >= AV_VERSION_INT(57, 37, 0)
        int ret;
        if ((ret = avcodec_send_packet(codec_ctx, &packet)) < 0) {
            fprintf(stderr, "Could not send video packet: %d\n", ret);
            goto run_quit;
        }
        ret = avcodec_receive_frame(codec_ctx, goGetHardwareFrame());
        if (!ret) {
            ret = goAvHwframeTransferData();
            if (!ret) {
                // a frame was received
                goPushFrame();
            } else {
                fprintf(stderr, "Could not receive video frame(1): %d\n", ret);
                goto run_quit;
            }
        } else if (ret != AVERROR(EAGAIN)) {
            fprintf(stderr, "Could not receive video frame(2): %d\n", ret);
            av_packet_unref(&packet);
            goto run_quit;
        }
#else
        while (packet.size > 0) {
            int got_picture;
            int len = avcodec_decode_video2(codec_ctx, goGetDecodingFrame(), &got_picture, &packet);
            if (len < 0) {
                fprintf(stderr, "Could not decode video packet: %d\n", len);
                goto run_quit;
            }
            if (got_picture) {
                goPushFrame();
            }
            packet.size -= len;
            packet.data += len;
        }
#endif
        av_packet_unref(&packet);

        if (avio_ctx->eof_reached) {
            break;
        }
    }

    fprintf(stderr, "End of frames\n");

run_quit:
    avformat_close_input(&format_ctx);
run_finally_free_avio_ctx:
    av_freep(&avio_ctx);
run_finally_free_format_ctx:
    avformat_free_context(format_ctx);
run_finally_close_codec:
    avcodec_close(codec_ctx);
run_finally_free_codec_ctx:
    avcodec_free_context(&codec_ctx);
    goNotifyStopped();
run_end:
    return 0;
}