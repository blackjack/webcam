// Example program that uses blakjack/webcam library
// for working with V4L2 devices.
// The application reads frames from device and writes them to stdout
// If your device supports motion formats (e.g. H264 or MJPEG) you can
// use it's output as a video stream.
// Example usage: go run stdout_streamer.go | vlc -
package main

import "github.com/blackjack/webcam"
import "os"
import "fmt"

func readChoice(s string) int {
	var i int
	for true {
		print(s)
		_, err := fmt.Scanf("%d\n", &i)
		if err != nil || i < 1 {
			println("Invalid input. Try again")
		} else {
			break
		}
	}
	return i
}

func main() {
	cam, err := webcam.Open("/dev/video0")
	if err != nil {
		panic(err.Error())
	}
	defer cam.Close()

	format_desc := cam.GetSupportedFormats()
	var formats []webcam.PixelFormat
	for f := range format_desc {
		formats = append(formats, f)
	}

	println("Available formats: ")
	for i, value := range formats {
		fmt.Fprintf(os.Stderr, "[%d] %s\n", i+1, format_desc[value])
	}

	choice := readChoice(fmt.Sprintf("Choose format [1-%d]: ", len(formats)))
	format := formats[choice-1]

	fmt.Fprintf(os.Stderr, "Supported frame sizes for format %s\n", format_desc[format])
	frames := cam.GetSupportedFrameSizes(format)
	for _, value := range frames {
		fmt.Fprintf(os.Stderr, "* %s\n", value.GetString())
	}
	width := readChoice("Enter frame width: ")
	height := readChoice("Enter frame height: ")

	f, w, h, err := cam.SetImageFormat(format, uint32(width), uint32(height))

	if err != nil {
		panic(err.Error())
	} else {
		fmt.Fprintf(os.Stderr, "Resulting image format: %s (%dx%d)\n", format_desc[f], w, h)
	}

	println("Press Enter to start streaming")
	fmt.Scanf("\n")
	err = cam.StartStreaming()
	if err != nil {
		panic(err.Error())
	}

	timeout := uint32(5) //5 seconds
	for cam.WaitForFrame(timeout) == nil {
		print(".")
		frame, err := cam.ReadFrame()
		if len(frame) != 0 {
			os.Stdout.Write(frame)
			os.Stdout.Sync()
		} else if err != nil {
			panic(err.Error())
		}
	}
}
