// Example program that reads the list of available controls and prints them.
package main

import "github.com/blackjack/webcam"
import "fmt"

func main() {
	cam, err := webcam.Open("/dev/video0")
	if err != nil {
		panic(err.Error())
	}
	defer cam.Close()

    cmap := cam.GetControls()

	fmt.Println("Available controls: ")
    for id, c := range cmap {
        fmt.Printf("ID:%08x %-32s  Min: %4d  Max: %5d\n", id, c.Name, c.Min, c.Max)
    }
}
