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

	"github.com/aamcrae/imageserver/snapshot"
	"github.com/aamcrae/webcam"
)

var port = flag.Int("port", 8080, "Web server port number")
var device = flag.String("input", "/dev/video0", "Input video device")
var resolution = flag.String("resolution", "800x600", "Selected resolution of camera")
var format = flag.String("format", "YUYV 4:2:2", "Selected pixel format of camera")
var controls = flag.String("controls", "focus=190,power_line_frequency=1",
	"Control parameters for camera")
var startDelay = flag.Int("delay", 0, "Delay at start (seconds)")
var verbose = flag.Bool("v", false, "Log more information")

var cnames map[string]webcam.ControlID = map[string]webcam.ControlID{
	"focus":                0x009a090a,
	"power_line_frequency": 0x00980918,
	"brightness":           0x00980900,
	"contrast":             0x00980901,
}

func init() {
	flag.Parse()
}

func main() {
	if *startDelay != 0 {
		time.Sleep(time.Duration(*startDelay) * time.Second)
	}
	cm := snapshot.NewSnapper()
	if err := cm.Open(*device, *format, *resolution); err != nil {
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

func readImage(cm *snapshot.Snapper, w http.ResponseWriter, r *http.Request) {
	if *verbose {
		log.Printf("URL request: %v", r.URL)
	}
	frame, err := cm.Snap()
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
