package frame

import (
	"fmt"
	"image"
	"image/color"
)

type FrameYUYV422 struct {
	model   color.Model
	b       image.Rectangle
	frame   []byte
	release func()
}

// Register this framer for this format.
func init() {
	RegisterFramer("YUYV 4:2:2", newFrameYUYV422)
}

// Wrap a raw webcam frame in a Frame so that it can be used as an image.
func newFrameYUYV422(x int, y int, f []byte, rel func()) (Frame, error) {
	expLen := 2 * x * y
	if len(f) != expLen {
		if rel != nil {
			defer rel()
		}
		return nil, fmt.Errorf("Wrong frame length (exp: %d, read %d)", expLen, len(f))
	}
	return &FrameYUYV422{model: color.YCbCrModel, b: image.Rect(0, 0, x, y), frame: f, release: rel}, nil
}

func (f *FrameYUYV422) ColorModel() color.Model {
	return f.model
}

func (f *FrameYUYV422) Bounds() image.Rectangle {
	return f.b
}

func (f *FrameYUYV422) At(x, y int) color.Color {
	index := f.b.Max.X*y*2 + (x&^1)*2
	if x&1 == 0 {
		return color.YCbCr{f.frame[index], f.frame[index+1], f.frame[index+3]}
	} else {
		return color.YCbCr{f.frame[index+2], f.frame[index+1], f.frame[index+3]}
	}
}

// Done with frame, release back to camera (if required).
func (f *FrameYUYV422) Release() {
	if f.release != nil {
		f.release()
	}
}
