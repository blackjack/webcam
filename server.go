// Program that serves jpeg images taken from a webcam.
package main

import (
	"flag"
	"fmt"
	"image/gif"
	"image/png"
	"image/jpeg"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aamcrae/imageserver/frame"
	"github.com/aamcrae/imageserver/snapshot"
	"github.com/aamcrae/webcam"
)

var port = flag.Int("port", 8080, "Web server port number")
var path = flag.String("path", "image.jpg", "Image filename path")
var device = flag.String("input", "/dev/video0", "Input video device")
var resolution = flag.String("resolution", "800x600", "Camera resolution")
var format = flag.String("format", "YUYV", "Pixel format of camera")
var controls = flag.String("controls", "",
	"Control parameters for camera")
var startDelay = flag.Int("delay", 0, "Delay at start (seconds)")
var verbose = flag.Bool("v", false, "Log more information")

var cnames map[string]webcam.ControlID = map[string]webcam.ControlID{
	"focus":                0x009a090a,
	"power_line_frequency": 0x00980918,
	"brightness":           0x00980900,
	"contrast":             0x00980901,
}

func main() {
	flag.Parse()
	s := strings.Split(*resolution, "x")
	if len(s) != 2 {
		log.Fatalf("%s: Illegal resolution", *resolution)
	}
	x, err := strconv.Atoi(s[0])
	if err != nil {
		log.Fatalf("%s: illegal width: %v", s[0], err)
	}
	y, err := strconv.Atoi(s[1])
	if err != nil {
		log.Fatalf("%s: illegal height: %v", s[1], err)
	}
	if *startDelay != 0 {
		time.Sleep(time.Duration(*startDelay) * time.Second)
	}
	cm := snapshot.NewSnapper()
	if err := cm.Open(*device, frame.FourCC(*format), x, y); err != nil {
		log.Fatalf("%s: %v", *device, err)
	}
	defer cm.Close()
	// Set camera controls.
	if len(*controls) != 0 {
		for _, control := range strings.Split(*controls, ",") {
			s := strings.Split(control, "=")
			if len(s) != 2 {
				log.Fatalf("Bad control option: %s", control)
			}
			id, ok := cnames[s[0]]
			if !ok {
				log.Fatalf("%s: Unknown control", s[0])
			}
			val, err := strconv.Atoi(s[1])
			if err != nil {
				log.Fatalf("Bad control value: %s (%v)", control, err)
			}
			if err = cm.SetControl(id, int32(val)); err != nil {
				log.Fatalf("SetControl error: %s (%v)", control, err)
			}
		}
	}
	var encode func(http.ResponseWriter, frame.Frame) error
	if strings.HasSuffix(*path, "jpg") || strings.HasSuffix(*path, "jpeg") {
		encode = encodeJpeg
	} else if strings.HasSuffix(*path, "png") {
		encode = encodePNG
	} else if strings.HasSuffix(*path, "gif") {
		encode = encodeGIF
	}
	http.Handle(fmt.Sprintf("/%s", *path), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if *verbose {
			log.Printf("URL request: %v", r.URL)
		}
		f, err := cm.Snap()
		if err != nil {
			log.Fatalf("Getframe: %v", err)
		}
		if err := encode(w, f); err != nil {
			log.Printf("Error writing image: %v\n", err)
		} else if *verbose {
			log.Printf("Wrote image successfully\n")
		}
	}))
	url := fmt.Sprintf(":%d", *port)
	if *verbose {
		log.Printf("Starting server on %s", url)
	}
	server := &http.Server{Addr: url}
	log.Fatal(server.ListenAndServe())
}

func encodeJpeg(w http.ResponseWriter, f frame.Frame) error {
	defer f.Release()
	w.Header().Set("Content-Type", "image/jpeg")
	return jpeg.Encode(w, f, nil)
}

func encodePNG(w http.ResponseWriter, f frame.Frame) error {
	defer f.Release()
	w.Header().Set("Content-Type", "image/png")
	return png.Encode(w, f)
}

func encodeGIF(w http.ResponseWriter, f frame.Frame) error {
	defer f.Release()
	w.Header().Set("Content-Type", "image/gif")
	return gif.Encode(w, f, nil)
}
