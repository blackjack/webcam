#ifndef WEBCAM_H
#define WEBCAM_H

#include <stdint.h>

int openWebcam(char* device);
int closeWebcam(int fd);
int checkCapabilities(int fd, int* is_video_device, int* can_stream);
int getPixelFormat(int fd, int index, uint32_t* code, char description[32]);
int getFrameSize(int fd, int index, uint32_t code, uint32_t frameSize[6]);
int setImageFormat(int fd, uint32_t* formatcode, uint32_t* width, uint32_t* height);

int mmapRequestBuffers(int fd,uint32_t* buf_count);
int mmapQueryBuffer(int fd,uint32_t index, uint32_t* length, void** start);
int mmapDequeueBuffer(int fd, uint32_t* index, uint32_t* length);
int mmapEnqueueBuffer(int fd,uint32_t index);
int mmapReleaseBuffer(void* start, uint32_t length);

int startStreaming(int fd);

int waitForFrame(int fd, uint32_t timeout);


#endif //WEBCAM_H