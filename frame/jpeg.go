package frame

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
)

type fJPEG struct {
	img     image.Image
	release func()
}

// Register this framer for this format.
func init() {
	RegisterFramer("JPEG", newJPEGFramer)
}

// Return a framer for JPEG.
func newJPEGFramer(w, h int) func([]byte, func()) (Frame, error) {
	return jpegFramer
}

// Wrap a jpeg block in a Frame so that it can be used as an image.
func jpegFramer(f []byte, rel func()) (Frame, error) {
	img, err := jpeg.Decode(bytes.NewBuffer(f))
	if err != nil {
		if rel != nil {
			rel()
		}
		return nil, err
	}
	return &fJPEG{img: img, release: rel}, nil
}

func (f *fJPEG) ColorModel() color.Model {
	return f.img.ColorModel()
}

func (f *fJPEG) Bounds() image.Rectangle {
	return f.img.Bounds()
}

func (f *fJPEG) At(x, y int) color.Color {
	return f.img.At(x, y)
}

// Done with frame, release back to camera (if required).
func (f *fJPEG) Release() {
	if f.release != nil {
		f.release()
	}
}
