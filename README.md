go-webcam
=========

Golang webcam wrapper. It depends on v4l2 framework (works only on Linux)

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
