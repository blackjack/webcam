// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Program that serves images taken from a webcam.
package main

import (
	"flag"
	"fmt"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/blackjack/webcam"
	"github.com/blackjack/webcam/frame"
	"github.com/blackjack/webcam/snapshot"
)

var port = flag.Int("port", 8080, "Web server port number")
var path = flag.String("path", "image", "Image base filename")
var device = flag.String("input", "/dev/video0", "Input video device")
var resolution = flag.String("resolution", "800x600", "Camera resolution")
var format = flag.String("format", "YUYV", "Pixel format of camera")
var controls = flag.String("controls", "",
	"Control parameters for camera (use --controls=list to list controls)")
var startDelay = flag.Int("delay", 2, "Delay at start (seconds)")
var verbose = flag.Bool("v", false, "Log more information")

var cnames map[string]webcam.ControlID = map[string]webcam.ControlID{
	"focus":                0x009a090a,
	"power_line_frequency": 0x00980918,
	"brightness":           0x00980900,
	"contrast":             0x00980901,
	"autoiso":              0x009a0918,
	"autoexp":              0x009a0901,
	"saturation":           0x00980902,
	"sharpness":            0x0098091b,
	"rotate":               0x00980922,
	"stabilization":        0x009a0916,
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
	if *controls == "list" {
		fmt.Printf("Control list (not all cameras may support all options):\n")
		for c, _ := range cnames {
			fmt.Printf("    %s\n", c)
		}
		return
	}
	if *startDelay != 0 {
		time.Sleep(time.Duration(*startDelay) * time.Second)
	}
	cm := snapshot.NewSnapper()
	if err := cm.Open(*device, frame.FourCC(*format), x, y); err != nil {
		log.Fatalf("%s: %v", *device, err)
	}
	if *startDelay != 0 {
		time.Sleep(time.Duration(*startDelay) * time.Second)
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
			if *verbose {
				fmt.Printf("Setting control '%s' to %d\n", s[0], val)
			}
			if err = cm.SetControl(id, int32(val)); err != nil {
				log.Fatalf("SetControl error: %s (%v)", control, err)
			}
		}
	}
	encodeJpeg := func(w http.ResponseWriter, f frame.Frame) error {
		w.Header().Set("Content-Type", "image/jpeg")
		return jpeg.Encode(w, f, nil)
	}
	encodePNG := func(w http.ResponseWriter, f frame.Frame) error {
		w.Header().Set("Content-Type", "image/png")
		return png.Encode(w, f)
	}
	encodeGIF := func(w http.ResponseWriter, f frame.Frame) error {
		w.Header().Set("Content-Type", "image/gif")
		return gif.Encode(w, f, nil)
	}
	http.Handle(fmt.Sprintf("/%s.jpg", *path), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		publishImage(cm, w, r, encodeJpeg)
	}))
	http.Handle(fmt.Sprintf("/%s.jpeg", *path), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		publishImage(cm, w, r, encodeJpeg)
	}))
	http.Handle(fmt.Sprintf("/%s.png", *path), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		publishImage(cm, w, r, encodePNG)
	}))
	http.Handle(fmt.Sprintf("/%s.gif", *path), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		publishImage(cm, w, r, encodeGIF)
	}))
	url := fmt.Sprintf(":%d", *port)
	if *verbose {
		log.Printf("Starting server on %s", url)
	}
	server := &http.Server{Addr: url}
	log.Fatal(server.ListenAndServe())
}

func publishImage(cm *snapshot.Snapper, w http.ResponseWriter, r *http.Request, encode func(http.ResponseWriter, frame.Frame) error) {
	if *verbose {
		log.Printf("URL request: %v", r.URL)
	}
	f, err := cm.Snap()
	if err != nil {
		log.Printf("Getframe: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Release()
	if err := encode(w, f); err != nil {
		log.Printf("Error writing image: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	} else if *verbose {
		log.Printf("Wrote image successfully\n")
	}
}
