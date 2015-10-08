#ifndef WEBCAM_H
#define WEBCAM_H

#include <stdint.h>

//wrappers to return some v4l2 C struct pointers CGO is not capable to understand
void* newFrmSizeEnum();

//
int openWebcam(char* device);
int checkCapabilities(int fd, int* is_video_device, int* can_stream);
int getPixelFormat(int fd, int index, uint32_t* code, char description[32]);
int getFrameSize(int fd, int index, uint32_t code, uint32_t frameSize[6]);
int setImageFormat(int fd, uint32_t* formatcode, uint32_t* width, uint32_t* height);

int mmapRequestBuffers(int fd,uint32_t* buf_count);
int mmapQueryBuffer(int fd,uint32_t index, uint32_t* length, void** start);
int mmapDequeueBuffer(int fd, uint32_t* index, uint32_t* length);
int mmapEnqueueBuffer(int fd,uint32_t index);

int startStreaming(int fd);

#endif //WEBCAM_H