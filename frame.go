// package frame wraps raw webcam frames as an image.
package frame

import (
	"fmt"
	"image"

	"github.com/aamcrae/webcam"
)

type FourCC string

// Release is called when the frame is no longer in use.
// The implementation may set a finalizer on the frame as a precaution
// in case Release is not called (which would cause a kernel resource leak).
type Frame interface {
	image.Image
	Release()
}

var framerFactoryMap = map[FourCC]func(int, int) func([]byte, func()) (Frame, error){}

// RegisterFramer registers a framer factory for a format.
// Note that only one handler can be registered for any single format.
func RegisterFramer(format FourCC, factory func(int, int) func([]byte, func()) (Frame, error)) {
	framerFactoryMap[format] = factory
}

// GetFramer returns a function that wraps the frame for this format.
func GetFramer(format FourCC, w, h int) (func([]byte, func()) (Frame, error), error) {
	if factory, ok := framerFactoryMap[format]; ok {
		return factory(w, h), nil
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
	return webcam.PixelFormat(uint32(f[0]) | uint32(f[1])<<8 | uint32(f[2])<<16 | uint32(f[3])<<24), nil
}
