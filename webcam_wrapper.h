#ifndef WEBCAM_WRAPPER_H
#define WEBCAM_WRAPPER_H

#include "webcam.h"

static buffer_t __frame = { NULL, 0 };
uint8_t* go_get_webcam_frame(const char* dev, int* length) {

    webcam_t *w = webcam_open(dev);

    webcam_resize(w, 640, 480);
    webcam_stream(w, true);

    while(__frame.length==0) {
        webcam_grab(w, &__frame);
    }

    fprintf(stderr, "length: %d\n", *length);
    if (__frame.length > 0) {
        *length = __frame.length;
        return __frame.start;
    } else
        return NULL;

    webcam_stream(w, false);
    webcam_close(w);

    if (__frame.start != NULL) free(__frame.start);
}

#endif
