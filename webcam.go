package main

/*
#cgo CFLAGS: -std=gnu99
#include <stdint.h>
#include <stdlib.h>
#include "webcam.h"
*/
import "C"
import "unsafe"
import "errors"
import "os"

type PixelFormat uint32

type FrameSize struct {
  min_width uint32;
  max_width uint32;
  step_width uint32;

  min_height uint32;
  max_height uint32;
  step_height uint32;
}

type buffer struct {
  start unsafe.Pointer;
  length uint32;
}

type Webcam struct {
  fd int;
  buffers []buffer;
}

func OpenWebcam(path string) (*Webcam,error) {
  cpath := C.CString(path)
  fd,err := C.openWebcam(cpath)
  C.free(unsafe.Pointer(cpath))
  
  if (fd<0) {
    return nil, err;
  }
  
  var is_video_device, can_stream C.int
  res, err := C.checkCapabilities(fd, &is_video_device, &can_stream);
  
  if (res<0) {
    return nil, err;
  }
  
  if int(is_video_device)==0 {
    return nil, errors.New("Not a video capture device")
  }
  
  if int(can_stream)==0 {
    return nil, errors.New("Device does not support the streaming I/O method")
  }
  
  w := new(Webcam);
  w.fd = int(fd);
  return w, nil;
}

func (w *Webcam) getSupportedFormats() map[PixelFormat]string {
  result := make(map[PixelFormat]string)
  
  var desc [32]C.char;
  var code C.uint32_t;
  
  for index := 0; C.getPixelFormat( C.int(w.fd), C.int(index), &code, &desc[0]) == 0; index++ {
    result[PixelFormat(code)] = C.GoString(&desc[0])
  }
  
  return result;
}

func (w *Webcam) getSupportedFrameSizes(f PixelFormat) []FrameSize {
  result := make([]FrameSize,0)
  
  var sizes [6]C.uint32_t
  
  for index := 0; C.getFrameSize( C.int(w.fd), C.int(index), C.uint32_t(f), &sizes[0]) == 0; index++ {
    var s FrameSize
    s.min_width = uint32(sizes[0]);
    s.max_width = uint32(sizes[1]);
    s.step_width = uint32(sizes[2]);
    s.min_height = uint32(sizes[3]);
    s.max_height = uint32(sizes[4]);
    s.step_height = uint32(sizes[5]);
    result = append(result,s)
  }
  
  return result;
}


func (w *Webcam) setImageFormat(f PixelFormat, width, height uint32) (PixelFormat,uint32,uint32,error) {
  
  code := C.uint32_t(f);
  cw := C.uint32_t(width);
  ch := C.uint32_t(height);
  
  
  res, err := C.setImageFormat(C.int(w.fd), &code, &cw, &ch);
  if (res<0) {
    return 0,0,0,err
  } else {
    return PixelFormat(code),uint32(cw),uint32(ch),nil
  }
}

func (w *Webcam) init() error {
  
  buf_count := C.uint32_t(256);
  
  res, err := C.mmapRequestBuffers(C.int(w.fd), &buf_count);
  if (res<0) {
    return err;
  }
  
  if (uint32(buf_count)<2) {
    return errors.New("Insufficient buffer memory")
  }
  
  w.buffers = make([]buffer,uint32(buf_count))
  
  for index,buf := range w.buffers {
    var length C.uint32_t
    var start unsafe.Pointer
    
    res, err := C.mmapQueryBuffer(C.int(w.fd), C.uint32_t(index), &length, &start)
    
    if (res<0) {
      if (err!=nil) {
        return err;
      } else {
        return errors.New("Failed to map memory");
      }
    }
    
    buf.start = start
    buf.length = uint32(length)
    w.buffers[index]=buf
  }
  
  for index,_ := range w.buffers {
    res, err = C.mmapEnqueueBuffer(C.int(w.fd), C.uint32_t(index))
    if (res<0) {
      return errors.New("Failed to enqueue buffer");
    }
  }
  
  return nil
}

func (w *Webcam) startStreaming() error {
  res, err := C.startStreaming(C.int(w.fd));
  if (res<0) {
    return err;
  } else {
    return nil;
  }
}

func (w *Webcam) readFrame() ([]byte,error) {
  var result []byte
  
  var index C.uint32_t
  var length C.uint32_t
  res, err := C.mmapDequeueBuffer(C.int(w.fd), &index, &length)
  
  if (res<0) {
    return nil,err;
  } else if (res>0) {
    return result,nil;
  }
  
  buffer := w.buffers[int(index)]
  result = C.GoBytes(buffer.start,C.int(length))
  
  res, err = C.mmapEnqueueBuffer(C.int(w.fd), index)
  
  if (res<0) {
    return nil, err;
  } else {
    return result,nil;
  }
}


func main() {
  cam,err := OpenWebcam("/dev/video0")
  if err!=nil {
    println(err.Error())
  }
  l := cam.getSupportedFormats()
  
  var p PixelFormat
  i := 0
  for p = range l {
    if i == 1 { break; }
    i++;
  }
  
  s := cam.getSupportedFrameSizes(p)

  println(l[p],len(s))
  
  p,w,h,err := cam.setImageFormat(p,1280,720)
  
  if (err!=nil) {
    println(err.Error())
  } else {
    println(l[p],w,h)
  }
  
  err = cam.init()
  
  if (err!=nil) {
    println(err.Error())
  } else {
    println(len(cam.buffers))
  }
  
  err = cam.startStreaming()
  
  for true {
    frame, err := cam.readFrame()
    if len(frame) != 0{
      os.Stdout.Write(frame)
    } else if (err!=nil) {
      break;
    }
  }
}