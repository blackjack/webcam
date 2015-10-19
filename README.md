# go-webcam

[![Build Status](https://travis-ci.org/blackjack/webcam.png?branch=master)](https://travis-ci.org/blackjack/webcam) [![GoDoc](https://godoc.org/github.com/google/go-github/github?status.svg)](https://godoc.org/github.com/blackjack/webcam)

This is a **go** library for working with webcams and other video capturing devices.
It depends entirely on [V4L2](http://linuxtv.org/downloads/v4l-dvb-apis/) framework, thus will compile and work only on **Linux** machine.

## Usage

```go
import "github.com/blackjack/webcam"

cam, err := webcam.Open("/dev/video0") // Open webcam
if err != nil { panic(err.Error()) }
defer cam.Close()
// ...
// Setup webcam image format and frame size here (see examples or documentation)
// ...
err = webcam.StartStreaming()
if err != nil { panic(err.Error()) }
for cam.WaitForFrame(5) == nil { // Wait with 5 seconds timeout
  print(".")
  frame, err := cam.ReadFrame()
  if len(frame) != 0 {
    // Process received frame
  } else if err != nil {
      panic(err.Error())
  }
}
```
For more detailed example see [examples folder](https://github.com/blackjack/webcam/tree/master/examples)

## Roadmap

The library is under development so possible API changes can still happen. Currently library supports streaming
using only MMAP method, which should be sufficient for most of devices available on the market.
Other streaming methods can be added in future (please create issue if you need this).

Also currently image format is defined by 4-byte code received from V4L2, which is good in terms of
compatibility with different versions of Linux kernel, but not very handy if you want to do some image manipulations.
Plans are to aligh V4L2 image format codes with [Image](https://golang.org/pkg/image/) package from Go library.

## License

See LICENSE file
