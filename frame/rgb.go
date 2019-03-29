package frame

import (
	"fmt"
	"image"
	"image/color"
	"runtime"
)

type fRGB struct {
	model   color.Model
	b       image.Rectangle
	stride  int
	size	int
	roffs	int
	goffs	int
	boffs	int
	frame   []byte
	release func()
}

// Register framers for these formats.
func init() {
	RegisterFramer("RGB3", newFramerRGB3)
	RegisterFramer("BGR3", newFramerBGR3)
}

// Return a function that is used as a framer for RGB3.
func newFramerRGB3(w, h, stride, size int) func([]byte, func()) (Frame, error) {
	return newRGBFramer(w, h, stride, size, 0, 1, 2)
}

// Return a function that is used as a framer for BGR3.
func newFramerBGR3(w, h, stride, size int) func([]byte, func()) (Frame, error) {
	return newRGBFramer(w, h, stride, size, 2, 1, 0)
}

// Return a function that is used as a generic RGB framer.
func newRGBFramer(w, h, stride, size, r, g, b int) func([]byte, func()) (Frame, error) {
	return func(buf []byte, rel func()) (Frame, error) {
		return frameRGB(size, stride, w, h, r, g, b, buf, rel)
	}
}

// Wrap a raw webcam frame in a Frame so that it can be used as an image.
func frameRGB(size, stride, w, h, rof, gof, bof int, b []byte, rel func()) (Frame, error) {
	if len(b) != size {
		if rel != nil {
			defer rel()
		}
		return nil, fmt.Errorf("Wrong frame length (exp: %d, read %d)", size, len(b))
	}
	f := &fRGB{model: color.RGBAModel, b: image.Rect(0, 0, w, h), stride: stride,
		roffs: rof, goffs: gof, boffs: bof, frame: b, release: rel}
	runtime.SetFinalizer(f, func(obj Frame) {
		obj.Release()
	})
	return f, nil
}

func (f *fRGB) ColorModel() color.Model {
	return f.model
}

func (f *fRGB) Bounds() image.Rectangle {
	return f.b
}

func (f *fRGB) At(x, y int) color.Color {
	i := f.stride*y + x*3
	return color.RGBA{f.frame[i+f.roffs], f.frame[i+f.goffs], f.frame[i+f.boffs], 0xFF}
}

// Done with frame, release back to camera (if required).
func (f *fRGB) Release() {
	if f.release != nil {
		f.release()
		// Make sure it only gets called once.
		f.release = nil
	}
}
