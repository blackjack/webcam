package main

import (
	"fmt"
	"image"
)

type Frame interface {
	image.Image
	Release()
}

var frameHandlers = map[string]func(int, int, []byte, func()) (Frame, error){}

func RegisterFramer(format string, handler func(int, int, []byte, func()) (Frame, error)) {
    frameHandlers[format] = handler
}

// Return a function that wraps the frame for this format.
func GetFramer(format string) (func(int, int, []byte, func()) (Frame, error), error) {
	if f, ok := frameHandlers[format]; ok {
		return f, nil
	}
	return nil, fmt.Errorf("No handler for format '%s'", format)
}
