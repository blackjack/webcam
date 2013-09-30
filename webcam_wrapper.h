#ifndef WEBCAM_WRAPPER_H
#define WEBCAM_WRAPPER_H

#include "webcam.h"

static buffer_t __frame = { NULL, 0 };
buffer_t go_get_webcam_frame(const char* dev) {

    webcam_t *w = webcam_open(dev);

    webcam_resize(w, 640, 480);
    webcam_stream(w, true);

    while(__frame.length==0) {
        webcam_grab(w, &__frame);
    }

    webcam_stream(w, false);
    webcam_close(w);

    return __frame;
}

#endif
