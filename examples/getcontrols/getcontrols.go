// Example program that reads the list of available controls and prints them.
package main

import (
	"flag"
	"fmt"

	"github.com/blackjack/webcam"
)

var device = flag.String("input", "/dev/video0", "Input video device")

func main() {
	flag.Parse()
	cam, err := webcam.Open(*device)
	if err != nil {
		panic(err.Error())
	}
	defer cam.Close()

	fmap := cam.GetSupportedFormats()
	fmt.Println("Available Formats: ")
	for p, s := range fmap {
		var pix []byte
		for i := 0; i < 4; i++ {
			pix = append(pix, byte(p>>uint(i*8)))
		}
		fmt.Printf("ID:%08x ('%s') %s\n   ", p, pix, s)
		for _, fs := range cam.GetSupportedFrameSizes(p) {
			fmt.Printf(" %s", fs.GetString())
		}
		fmt.Printf("\n")
	}

	cmap := cam.GetControls()
	fmt.Println("Available controls: ")
	for id, c := range cmap {
		fmt.Printf("ID:%08x %-32s Type: %1d Min: %6d Max: %6d Step: %6d\n", id, c.Name, c.Type, c.Min, c.Max, c.Step)
	}
}
