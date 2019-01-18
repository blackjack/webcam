package frame

import (
	"flag"
	"fmt"
	"image"
	"image/color"
)

type fYUYV422 struct {
	model   color.Model
	b       image.Rectangle
	width   int
	frame   []byte
	release func()
}

var padded = flag.Bool("padded", false, "Frame has padding")

// Register a framer factory for this format.
func init() {
	RegisterFramer("YUYV", newFramerYUYV422)
}

func newFramerYUYV422(w, h int) func([]byte, func()) (Frame, error) {
	var size, bw int
	if *padded {
		bw = (w + 31) &^ 31
		size = 2 * bw * ((h + 15) &^ 15)
	} else {
		size = 2 * h * w
	}
	return func(b []byte, rel func()) (Frame, error) {
		return frameYUYV422(size, bw, w, h, b, rel)
	}
}

// Wrap a raw webcam frame in a Frame so that it can be used as an image.
func frameYUYV422(size, bw, w, h int, b []byte, rel func()) (Frame, error) {
	if len(b) != size {
		if rel != nil {
			defer rel()
		}
		return nil, fmt.Errorf("Wrong frame length (exp: %d, read %d)", size, len(b))
	}
	return &fYUYV422{model: color.YCbCrModel, b: image.Rect(0, 0, w, h), width: bw, frame: b, release: rel}, nil
}

func (f *fYUYV422) ColorModel() color.Model {
	return f.model
}

func (f *fYUYV422) Bounds() image.Rectangle {
	return f.b
}

func (f *fYUYV422) At(x, y int) color.Color {
	index := f.width*y*2 + (x&^1)*2
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
	}
}
