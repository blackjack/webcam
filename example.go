package main

import "github.com/blackjack/webcam"
import "os"

func main() {
  cam,err := Open("/dev/video0")
  defer cam.Close()
  if err!=nil {
    println(err.Error())
  }
  formats := cam.GetSupportedFormats()
  
  var p PixelFormat
  i := 0
  for p = range formats {
    if i == 1 { break; }
    i++
  }
  
  s := cam.GetSupportedFrameSizes(p)

  println(l[p],len(s))
  
  p,w,h,err := cam.SetImageFormat(p,1280,720)
  
  if (err!=nil) {
    println(err.Error())
  } else {
    println(l[p],w,h)
  }
  
  err = cam.Init()
  
  if (err!=nil) {
    println(err.Error())
  } else {
    println(len(cam.buffers))
  }
  
  err = cam.StartStreaming()
  
  for cam.WaitForFrame(5)==nil {
    print(".")
    frame, err := cam.ReadFrame()
    if len(frame) != 0{
      os.Stdout.Write(frame)
    } else if (err!=nil) {
      break
    }
  }
}
