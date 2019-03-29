// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package frame

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"runtime"
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
func newJPEGFramer(w, h, stride, size int) func([]byte, func()) (Frame, error) {
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
	fr := &fJPEG{img: img, release: rel}
	runtime.SetFinalizer(fr, func(obj Frame) {
		obj.Release()
	})
	return fr, nil
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
		// Make sure it only gets called once.
		f.release = nil
	}
}
