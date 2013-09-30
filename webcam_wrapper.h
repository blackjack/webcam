#ifndef WEBCAM_WRAPPER_H
#define WEBCAM_WRAPPER_H

#include "webcam.h"

buffer_t go_get_webcam_frame(const char* dev) {
    buffer_t frame = { NULL, 0 };

    webcam_t *w = webcam_open(dev);

    webcam_resize(w, 640, 480);
    webcam_stream(w, true);

    while(frame.length==0) {
        webcam_grab(w, &frame);
    }

    webcam_stream(w, false);
    webcam_close(w);

    return frame;
}

#endif
