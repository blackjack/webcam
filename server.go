package main

import (
    "flag"
    "fmt"
    "image/png"
    "log"
    "net/http"
)

var port = flag.Int("port", 8080, "Web server port number")
var device = flag.String("input", "/dev/video0", "Input video device")
var resolution = flag.String("resolution", "800x600", "Selected resolution of camera")
var format = flag.String("format", "YUYV 4:2:2", "Selected pixel format of camera")
var verbose = flag.Bool("v", false, "Log more information")

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
    http.Handle("/image", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
            readImage(cam, w, r)
        }))
    url := fmt.Sprintf(":%d", *port)
    if *verbose {
        log.Printf("Starting server on %s", url)
    }
    s := &http.Server{Addr: url}
    log.Fatal(s.ListenAndServe())
}

func readImage(cam *Camera, w http.ResponseWriter, r *http.Request) {
    if *verbose {
        log.Printf("URL request: %v", r.URL)
    }
    frame, err := cam.GetFrame()
    if err != nil {
        log.Fatalf("Getframe: %v", err)
    }
    w.Header().Set("Content-Type", "image/png")
    if err := png.Encode(w, frame); err != nil {
        log.Printf("Error writing image: %v\n", err)
    } else if *verbose {
        log.Printf("Wrote image successfully\n")
    }
    frame.Release()
}
