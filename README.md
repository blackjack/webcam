go-webcam
=========

Golang webcam wrapper. It depends on v4l2 framework (works only on Linux)

    package main

    import "github.com/blackjack/webcam"
    import "os"
    import "fmt"

    func readChoice(s string) int {
      var i int
      for true {
        print(s)
        _, err := fmt.Scanf("%d\n", &i)
        if err != nil {
          println("Invalid input. Try again")
        } else {
          break
        }
      }
      return i
    }

    func getFormatByNumber(m map[webcam.PixelFormat]string, n int) webcam.PixelFormat {
      i := 1
      for key := range m {
        if i == n {
          return key
        }
        i++
      }
      panic(fmt.Sprintf("Index out of range. N:%d, Len:%d", n, len(m)))
    }

    func main() {
      cam, err := webcam.Open("/dev/video0")
      if err != nil {
        println(err.Error())
      }
      defer cam.Close()

      formats := cam.GetSupportedFormats()

      println("Available formats: ")
      i := 0
      for _, value := range formats {
        i++
        fmt.Printf("[%d] %s\n", i, value)
      }

      choice := readChoice(fmt.Sprintf("Choose format [1-%d]: ", len(formats)))
      format := getFormatByNumber(formats, choice)

      fmt.Printf("Supported frame sizes for format %s\n", formats[format])
      frames := cam.GetSupportedFrameSizes(format)
      for _, value := range frames {
        fmt.Printf("* %s\n", value.GetString())
      }
      width := readChoice("Enter frame width: ")
      height := readChoice("Enter frame height: ")

      p, w, h, err := cam.SetImageFormat(format, uint32(width), uint32(height))

      if err != nil {
        println(err.Error())
      } else {
        fmt.Printf("Resulting image format: %s (%dx%d)\n", formats[p], w, h)
      }

      fmt.Println("Press Enter to start streaming")
      fmt.Scanf("\n")
      err = cam.Init()

      if err != nil {
        println(err.Error())
      }

      err = cam.StartStreaming()

      timeout := uint32(5) //5 seconds
      for cam.WaitForFrame(timeout) == nil {
        print(".")
        frame, err := cam.ReadFrame()
        if len(frame) != 0 {
          os.Stdout.Write(frame)
        } else if err != nil {
          break
        }
      }
    }

