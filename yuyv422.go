package frame

import (
	"fmt"
	"image"
	"image/color"
	"runtime"
)

type fYUYV422 struct {
	model   color.Model
	b       image.Rectangle
	stride  int
	frame   []byte
	release func()
}

// Register a framer factory for this format.
func init() {
	RegisterFramer("YUYV", newFramerYUYV422)
}

func newFramerYUYV422(w, h, stride int) func([]byte, func()) (Frame, error) {
	return func(b []byte, rel func()) (Frame, error) {
		return frameYUYV422(h * stride, stride, w, h, b, rel)
	}
}

// Wrap a raw webcam frame in a Frame so that it can be used as an image.
func frameYUYV422(size, stride, w, h int, b []byte, rel func()) (Frame, error) {
	if len(b) != size {
		if rel != nil {
			defer rel()
		}
		return nil, fmt.Errorf("Wrong frame length (exp: %d, read %d)", size, len(b))
	}
	f := &fYUYV422{model: color.YCbCrModel, b: image.Rect(0, 0, w, h), stride: stride, frame: b, release: rel}
	runtime.SetFinalizer(f, func(obj Frame) {
		obj.Release()
	})
	return f, nil
}

func (f *fYUYV422) ColorModel() color.Model {
	return f.model
}

func (f *fYUYV422) Bounds() image.Rectangle {
	return f.b
}

func (f *fYUYV422) At(x, y int) color.Color {
	index := f.stride*y + (x&^1)*2
	if x&1 == 0 {
		return color.YCbCr{f.frame[index], f.frame[index+1], f.frame[index+3]}
	} else {
		return color.YCbCr{f.frame[index+2], f.frame[index+1], f.frame[index+3]}
	}
}

// Done with frame, release back to camera (if required).
func (f *fYUYV422) Release() {
	if f.release != nil {
		f.release()
		// Make sure it only gets called once.
		f.release = nil
	}
}
