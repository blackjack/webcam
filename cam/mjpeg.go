package cam

import (
    "bytes"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	_ "image/gif"
    "os"
)

type FrameMJPEG struct {
    img    image.Image
	release func()
}

// Register this framer for this format.
func init() {
    RegisterFramer("Motion-JPEG", newFrameMJPEG)
}

// Wrap a mjpeg still in a Frame so that it can be used as an image.
func newFrameMJPEG(x int, y int, f []byte, rel func()) (Frame, error) {
    file, err := os.Create("/tmp/xfile")
    if err != nil {
        return nil, err
    }
    file.Write(f)
    file.Close()
    j, _, err := image.Decode(bytes.NewBuffer(f))
    if err != nil {
		if rel != nil {
			rel()
		}
		return nil, err
	}
	return &FrameMJPEG{img: j, release: rel}, nil
}

func (f *FrameMJPEG) ColorModel() color.Model {
	return f.img.ColorModel()
}

func (f *FrameMJPEG) Bounds() image.Rectangle {
	return f.img.Bounds()
}

func (f *FrameMJPEG) At(x, y int) color.Color {
    return f.img.At(x, y)
}

// Done with frame, release back to camera (if required).
func (f *FrameMJPEG) Release() {
	if f.release != nil {
		f.release()
	}
}
