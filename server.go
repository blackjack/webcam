package main

import (
	"flag"
	"fmt"
	"image/jpeg"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aamcrae/imageserver/camera"
)

var port = flag.Int("port", 8080, "Web server port number")
var device = flag.String("input", "/dev/video0", "Input video device")
var resolution = flag.String("resolution", "800x600", "Selected resolution of camera")
var format = flag.String("format", "YUYV 4:2:2", "Selected pixel format of camera")
var controls = flag.String("controls", "focus=190,power_line_frequency=1",
	"Control parameters for camera")
var startDelay = flag.Int("delay", 0, "Delay at start (seconds)")
var verbose = flag.Bool("v", false, "Log more information")

func init() {
	flag.Parse()
}

func main() {
	if *startDelay != 0 {
		time.Sleep(time.Duration(*startDelay) * time.Second)
	}
	cm, err := camera.Open(*device)
	if err != nil {
		log.Fatalf("%s: %v", *device, err)
	}
	defer cm.Close()
	if err := cm.Init(*format, *resolution); err != nil {
		log.Fatalf("Init failed: %v", err)
	}
	// Initialise camera controls.
	if len(*controls) != 0 {
		for _, control := range strings.Split(*controls, ",") {
			// If no parameter, assume bool and set to true.
			s := strings.Split(control, "=")
			if len(s) == 1 {
				s = append(s, "true")
			}
			if len(s) != 2 {
				log.Fatalf("Bad control option: %s", control)
			}
			val, err := strconv.Atoi(s[1])
			if err != nil {
				log.Fatalf("Bad control value: %s (%v)", control, err)
			}
			if err = cm.SetControl(s[0], int32(val)); err != nil {
				log.Fatalf("SetControl error: %s (%v)", control, err)
			}
		}
	}
	http.Handle("/image", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		readImage(cm, w, r)
	}))
	url := fmt.Sprintf(":%d", *port)
	if *verbose {
		log.Printf("Starting server on %s", url)
	}
	s := &http.Server{Addr: url}
	log.Fatal(s.ListenAndServe())
}

func readImage(cm *camera.Camera, w http.ResponseWriter, r *http.Request) {
	if *verbose {
		log.Printf("URL request: %v", r.URL)
	}
	frame, err := cm.GetFrame()
	if err != nil {
		log.Fatalf("Getframe: %v", err)
	}
	defer frame.Release()
	w.Header().Set("Content-Type", "image/jpeg")
	if err := jpeg.Encode(w, frame, nil); err != nil {
		log.Printf("Error writing image: %v\n", err)
	} else if *verbose {
		log.Printf("Wrote image successfully\n")
	}
}
