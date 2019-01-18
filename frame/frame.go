// package frame wraps raw webcam frames as an image.
package frame

import (
	"fmt"
	"image"

	"github.com/aamcrae/webcam"
)

type FourCC string

type Frame interface {
	image.Image
	Release()
}

var frameHandlers = map[FourCC]func(int, int, []byte, func()) (Frame, error){}

// RegisterFramer registers a frame handler for a format.
// Note that only one handler can be registered for any single format.
func RegisterFramer(format FourCC, handler func(int, int, []byte, func()) (Frame, error)) {
	frameHandlers[format] = handler
}

// GetFramer returns a function that wraps the frame for this format.
func GetFramer(format FourCC) (func(int, int, []byte, func()) (Frame, error), error) {
	if f, ok := frameHandlers[format]; ok {
		return f, nil
	}
	return nil, fmt.Errorf("No handler for format '%s'", format)
}

// PixelFormatToFourCC converts the v4l2 PixelFormat to a FourCC.
func PixelFormatToFourCC(pf webcam.PixelFormat) FourCC {
	b := make([]byte, 4)
	b[0] = byte(pf)
	b[1] = byte(pf >> 8)
	b[2] = byte(pf >> 16)
	b[3] = byte(pf >> 24)
	return FourCC(b)
}

// FourCCToPixelFormat converts the four character string to a v4l2 PixelFormat.
func FourCCToPixelFormat(f FourCC) (webcam.PixelFormat, error) {
	if len(f) != 4 {
		return 0, fmt.Errorf("%s: Illegal FourCC", f)
	}
	return webcam.PixelFormat(uint32(f[0]) | uint32(f[1]) << 8 | uint32(f[2]) << 16 | uint32(f[3]) << 24), nil
}
