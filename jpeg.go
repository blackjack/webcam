package frame

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
)

type FrameJPEG struct {
	img     image.Image
	release func()
}

// Register this framer for this format.
func init() {
	RegisterFramer("JPEG", newFrameJPEG)
}

// Wrap a jpeg block in a Frame so that it can be used as an image.
func newFrameJPEG(x int, y int, f []byte, rel func()) (Frame, error) {
	img, err := jpeg.Decode(bytes.NewBuffer(f))
	if err != nil {
		if rel != nil {
			rel()
		}
		return nil, err
	}
	return &FrameJPEG{img: img, release: rel}, nil
}

func (f *FrameJPEG) ColorModel() color.Model {
	return f.img.ColorModel()
}

func (f *FrameJPEG) Bounds() image.Rectangle {
	return f.img.Bounds()
}

func (f *FrameJPEG) At(x, y int) color.Color {
	return f.img.At(x, y)
}

// Done with frame, release back to camera (if required).
func (f *FrameJPEG) Release() {
	if f.release != nil {
		f.release()
	}
}
