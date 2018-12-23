package main

import (
    "flag"
    "fmt"
    "image/png"
    "log"
    "os"
)

var device = flag.String("input", "/dev/video0", "Input video device")
var resolution = flag.String("resolution", "800x600", "Selected resolution of camera")
var format = flag.String("format", "YUYV 4:2:2", "Selected pixel format of camera")

func init() {
    flag.Parse()
}

func main() {
    cam, err := OpenCamera(*device)
    if err != nil {
        log.Fatalf("%s: %v", *device, err)
    }
	defer cam.Close()
    if err := cam.Init(*format, *resolution); err != nil {
		log.Fatalf("Init failed: %v", err)
    }
    frame, err := cam.GetFrame()
    if err != nil {
        log.Fatalf("Getframe: %v", err)
    }
    fname := "test.png"
    of, err := os.Create(fname)
    if err != nil {
		 log.Fatalf("Failed to create %s: %v", fname, err)
    }
    if err := png.Encode(of, frame); err != nil {
        fmt.Printf("Error writing %s: %v\n", fname, err)
    } else {
        fmt.Printf("Wrote %s successfully\n", fname)
    }
    frame.Release()
    of.Close()
}
