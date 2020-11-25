// Example program that reads the list of available controls and prints them.
package main

import (
	"flag"
	"fmt"
	"sort"

	"github.com/blackjack/webcam"
)

var device = flag.String("input", "/dev/video0", "Input video device")

type control struct {
	id       webcam.ControlID
	name     string
	min, max int32
}

func main() {
	flag.Parse()
	cam, err := webcam.Open(*device)
	if err != nil {
		panic(fmt.Errorf("%s: %v", *device, err.Error()))
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
	var clist []control
	for id, cm := range cmap {
		var c control
		c.id = id
		c.name = cm.Name
		c.min = cm.Min
		c.max = cm.Max
		clist = append(clist, c)
	}
	sort.Slice(clist, func(i, j int) bool {
		return clist[i].name < clist[j].name
	})
	for _, cl := range clist {
		fmt.Printf("ID:%08x %-32s  Min: %4d  Max: %5d\n", cl.id,
			cl.name, cl.min, cl.max)
	}
}
