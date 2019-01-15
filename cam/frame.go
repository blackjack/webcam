package cam

import (
	"fmt"
	"image"
)

type Frame interface {
	image.Image
	Release()
}

var frameHandlers = map[string]func(int, int, []byte, func()) (Frame, error){}

// RegisterFramer registers a frame handler for a particular format.
// Note that only one handler can be registered for any format.
func RegisterFramer(format string, handler func(int, int, []byte, func()) (Frame, error)) {
    frameHandlers[format] = handler
}

// GetFramer returns a function that wraps the frame for this format.
func GetFramer(format string) (func(int, int, []byte, func()) (Frame, error), error) {
	if f, ok := frameHandlers[format]; ok {
		return f, nil
	}
	return nil, fmt.Errorf("No handler for format '%s'", format)
}
