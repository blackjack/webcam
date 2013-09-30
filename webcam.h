#include <sys/types.h>
#include <sys/stat.h>
#include <sys/mman.h>
#include <fcntl.h>

#include <assert.h>
#include <pthread.h>
#include <stdlib.h>
#include <stdio.h>
#include <stdbool.h>
#include <stdint.h>
#include <errno.h>
#include <string.h>
#include <unistd.h>

#include <linux/videodev2.h>

#define CLEAR(x) memset(&(x), 0, sizeof(x))

/**
 * Buffer structure
 */
typedef struct buffer {
    uint8_t *start;
    size_t  length;
} buffer_t;

/**
 * Webcam structure
 */
typedef struct webcam {
    char            *name;
    int             fd;
    buffer_t        *buffers;
    uint8_t         nbuffers;

    buffer_t        frame;
    pthread_t       thread;
    pthread_mutex_t mtx_frame;

    uint16_t        width;
    uint16_t        height;
    uint8_t         colorspace;

    char            formats[16][5];
    bool            streaming;
} webcam_t;

webcam_t *webcam_open(const char *dev);
void webcam_close(webcam_t *w);
void webcam_resize(webcam_t *w, uint16_t width, uint16_t height);
void webcam_stream(webcam_t *w, bool flag);
void webcam_grab(webcam_t *w, buffer_t *frame);
