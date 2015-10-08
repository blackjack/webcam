#include "webcam.h"

#include <linux/videodev2.h>
#include <sys/ioctl.h>
#include <sys/mman.h>
#include <string.h>
#include <stdio.h>
#include <fcntl.h>
#include <errno.h>
#include <unistd.h>

#define CLEAR(x) memset(&(x), 0, sizeof(x));

int _ioctl( int fh, int request, void* arg )
{
  int r;

  do {
    r = ioctl( fh, request, arg );
  } while ( -1 == r && EINTR == errno );

  return r;
}

int openWebcam( char* device )
{
  return open( device, O_RDWR | O_NONBLOCK );
}

int closeWebcam( int fd )
{
  return close( fd );
}

int checkCapabilities( int fd, int* is_video_device, int* can_stream )
{
  struct v4l2_capability cap;
  CLEAR( cap );
  int res = _ioctl( fd, VIDIOC_QUERYCAP, &cap );

  if ( res < 0 ) {
    return res;
  }

  *is_video_device = cap.capabilities & V4L2_CAP_VIDEO_CAPTURE;
  *can_stream = cap.capabilities & V4L2_CAP_STREAMING;

  return res;
}

int getPixelFormat( int fd, int index, uint32_t* code, char description[32] )
{
  struct v4l2_fmtdesc vfd;
  CLEAR( vfd );

  vfd.index = index;
  vfd.type = V4L2_BUF_TYPE_VIDEO_CAPTURE;

  int res = _ioctl( fd, VIDIOC_ENUM_FMT, &vfd );

  if ( res != 0 ) { return res; }

  *code = vfd.pixelformat;
  memcpy( description, vfd.description, 32 );

  return res;
}


int getFrameSize( int fd, int index, uint32_t code, uint32_t frameSize[6] )
{
  struct v4l2_frmsizeenum vfse;
  CLEAR( vfse );
  vfse.index = index;
  vfse.pixel_format = code;

  int res = _ioctl( fd, VIDIOC_ENUM_FRAMESIZES, &vfse );

  if ( res < 0 ) { return res; }

  switch ( vfse.type ) {
  case V4L2_FRMSIZE_TYPE_DISCRETE:
    frameSize[0] = vfse.discrete.width;
    frameSize[1] = vfse.discrete.width;
    frameSize[2] = 0;
    frameSize[3] = vfse.discrete.height;
    frameSize[4] = vfse.discrete.height;
    frameSize[5] = 0;
    return res;

  case V4L2_FRMSIZE_TYPE_CONTINUOUS:
  case V4L2_FRMSIZE_TYPE_STEPWISE:
    frameSize[0] = vfse.stepwise.min_width;
    frameSize[1] = vfse.stepwise.max_width;
    frameSize[2] = vfse.stepwise.step_width;
    frameSize[3] = vfse.stepwise.min_height;
    frameSize[4] = vfse.stepwise.max_height;
    frameSize[5] = vfse.stepwise.step_height;
    return res;
  }

  return res;
}


int setImageFormat( int fd, uint32_t* formatcode, uint32_t* width, uint32_t* height )
{
  struct v4l2_format fmt;
  CLEAR( fmt );
  fmt.type = V4L2_BUF_TYPE_VIDEO_CAPTURE;
  fmt.fmt.pix.width = *width;
  fmt.fmt.pix.height = *height;
  fmt.fmt.pix.pixelformat = *formatcode;
  fmt.fmt.pix.field = V4L2_FIELD_ANY;

  int res = _ioctl( fd, VIDIOC_S_FMT, &fmt );

  if ( res < 0 ) { return res; }

  *width = fmt.fmt.pix.width;
  *height = fmt.fmt.pix.height;
  *formatcode = fmt.fmt.pix.pixelformat;

  return res;
}


int mmapRequestBuffers( int fd, uint32_t* buf_count )
{
  struct v4l2_requestbuffers req;
  CLEAR( req );

  req.count = *buf_count;
  req.type = V4L2_BUF_TYPE_VIDEO_CAPTURE;
  req.memory = V4L2_MEMORY_MMAP;

  int res = _ioctl( fd, VIDIOC_REQBUFS, &req );

  if ( res < 0 ) { return res; }

  *buf_count = req.count;
  return res;;
}

int mmapQueryBuffer( int fd, uint32_t index, uint32_t* length, void** start )
{
  struct v4l2_buffer buf;
  CLEAR( buf );
  buf.type        = V4L2_BUF_TYPE_VIDEO_CAPTURE;
  buf.memory      = V4L2_MEMORY_MMAP;
  buf.index       = index;

  int res = _ioctl( fd, VIDIOC_QUERYBUF, &buf );

  if ( res < 0 ) { return res; }

  *length = buf.length;
  *start = mmap( NULL /* start anywhere */,
                 buf.length,
                 PROT_READ | PROT_WRITE /* required */,
                 MAP_SHARED /* recommended */,
                 fd, buf.m.offset );

  if ( *start == MAP_FAILED ) {
    return -1;
  } else {
    return 0;
  }
}

int mmapDequeueBuffer( int fd, uint32_t* index, uint32_t* length )
{
  struct v4l2_buffer buf;
  CLEAR( buf );
  buf.type = V4L2_BUF_TYPE_VIDEO_CAPTURE;
  buf.memory = V4L2_MEMORY_MMAP;

  int res = _ioctl( fd, VIDIOC_DQBUF, &buf );
  *index = buf.index;
  *length = buf.bytesused;

  if ( res < 0 && errno == EAGAIN ) {
    return 1;
  }

  return res;
}

int mmapEnqueueBuffer( int fd, uint32_t index )
{
  struct v4l2_buffer buf;
  CLEAR( buf );

  buf.type = V4L2_BUF_TYPE_VIDEO_CAPTURE;
  buf.memory = V4L2_MEMORY_MMAP;
  buf.index = index;

  int res = _ioctl( fd, VIDIOC_QBUF, &buf );
  return res;
}

int mmapReleaseBuffer( void* start, uint32_t length )
{
  return munmap( start, length );
}

int startStreaming( int fd )
{
  enum v4l2_buf_type type = V4L2_BUF_TYPE_VIDEO_CAPTURE;
  int res = _ioctl( fd, VIDIOC_STREAMON, &type );
  return res;
}

int waitForFrame( int fd, uint32_t timeout )
{
  for ( ;; ) {
    fd_set fds;
    struct timeval tv;
    FD_ZERO( &fds );
    FD_SET( fd, &fds );

    tv.tv_sec = timeout;
    tv.tv_usec = 0;

    int res = select( fd + 1, &fds, NULL, NULL, &tv );

    if ( res < 0 ) {
      if ( errno == EINTR ) {
        continue;
      }

      return res;
    }

    return res;
  }
}




