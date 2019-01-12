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

    clist := cam.GetControlList()

	fmt.Println("Available controls: ")
    for _, c := range clist {
        fmt.Printf("%32s  Min: %4d  Max: %5d\n", c.Name, c.Min, c.Max)
    }
}
