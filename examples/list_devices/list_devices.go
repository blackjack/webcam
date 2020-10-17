package main

import (
	"flag"
	"fmt"

	"github.com/blackjack/webcam"
)

func main() {
	flag.Parse()

	devices, err := webcam.ListDevices()
	if err != nil {
		panic(err.Error())
	}
	if len(devices) == 0 {
		fmt.Println("No valid video devices found in %q", webcam.VIDEO4LINUX_DIR)
	} else {
		fmt.Println("Video devices found:")
		for devPath, name := range devices {
			fmt.Printf("  %q located in %s", name, devPath)
		}
	}
}

