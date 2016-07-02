// Library for working with webcams and other video capturing devices.
// It depends entirely on v4l2 framework, thus will compile and work
// only on Linux machine
package webcam

import (
	"errors"
	"golang.org/x/sys/unix"
	"os"
	"reflect"
	"unsafe"
)

// Webcam object
type Webcam struct {
	file    *os.File
	fd      uintptr
	buffers [][]byte
}

// Open a webcam with a given path
// Checks if device is a v4l2 device and if it is
// capable to stream video
func Open(path string) (*Webcam, error) {

	file, err := os.OpenFile(path, unix.O_RDWR|unix.O_NONBLOCK, 0666)
	fd := file.Fd()

	if fd < 0 || err != nil {
		return nil, err
	}

	supportsVideoCapture, supportsVideoStreaming, err := checkCapabilities(fd)

	if err != nil {
		return nil, err
	}

	if !supportsVideoCapture {
		return nil, errors.New("Not a video capture device")
	}

	if !supportsVideoStreaming {
		return nil, errors.New("Device does not support the streaming I/O method")
	}

	w := new(Webcam)
	w.fd = fd
	w.file = file
	return w, nil
}

// Returns image formats supported by the device alongside with
// their text description
// Not that this function is somewhat experimental. Frames are not ordered in
// any meaning, also duplicates can occur so it's up to developer to clean it up.
// See http://linuxtv.org/downloads/v4l-dvb-apis/vidioc-enum-framesizes.html
// for more information
func (w *Webcam) GetSupportedFormats() map[PixelFormat]string {

	result := make(map[PixelFormat]string)
	var err error
	var code uint32
	var desc string
	var index uint32

	for index = 0; err == nil; index++ {
		code, desc, err = getPixelFormat(w.fd, index)

		if err != nil {
			break
		}

		result[PixelFormat(code)] = desc
	}

	return result
}

// Returns supported frame sizes for a given image format
func (w *Webcam) GetSupportedFrameSizes(f PixelFormat) []FrameSize {
	result := make([]FrameSize, 0)

	var index uint32
	var err error

	for index = 0; err == nil; index++ {
		s, err := getFrameSize(w.fd, index, uint32(f))

		if err != nil {
			break
		}

		result = append(result, s)
	}

	return result
}

// Sets desired image format and frame size
// Note, that device driver can change that values.
// Resulting values are returned by a function
// alongside with an error if any
func (w *Webcam) SetImageFormat(f PixelFormat, width, height uint32) (PixelFormat, uint32, uint32, error) {

	code := uint32(f)
	cw := width
	ch := height

	err := setImageFormat(w.fd, &code, &width, &height)

	if err != nil {
		return 0, 0, 0, err
	} else {
		return PixelFormat(code), cw, ch, nil
	}
}

// Start streaming process
func (w *Webcam) StartStreaming() error {

	var buf_count uint32 = 256

	err := mmapRequestBuffers(w.fd, &buf_count)

	if err != nil {
		return errors.New("Failed to map request buffers: " + string(err.Error()))
	}

	if buf_count < 2 {
		return errors.New("Insufficient buffer memory")
	}

	w.buffers = make([][]byte, buf_count, buf_count)
	for index, _ := range w.buffers {
		var length uint32

		buffer, err := mmapQueryBuffer(w.fd, uint32(index), &length)

		if err != nil {
			return errors.New("Failed to map memory: " + string(err.Error()))
		}

		w.buffers[index] = buffer
	}

	for index, _ := range w.buffers {

		err := mmapEnqueueBuffer(w.fd, uint32(index))

		if err != nil {
			return errors.New("Failed to enqueue buffer: " + string(err.Error()))
		}

	}

	err = startStreaming(w.fd)

	if err != nil {
		return errors.New("Failed to start streaming: " + string(err.Error()))
	}

	return nil
}

// Read a single frame from the webcam
// If frame cannot be read at the moment
// function will return empty slice
func (w *Webcam) ReadFrame() ([]byte, error) {
	var index uint32
	var length uint32

	err := mmapDequeueBuffer(w.fd, &index, &length)

	if err != nil {
		return nil, err
	}

	result := w.buffers[int(index)][:length]

	err = mmapEnqueueBuffer(w.fd, index)

	return result, err

}

// Wait until frame could be read
func (w *Webcam) WaitForFrame(timeout uint32) error {

	count, err := waitForFrame(w.fd, timeout)

	if count < 0 || err != nil {
		return err
	} else if count == 0 {
		return new(Timeout)
	} else {
		return nil
	}
}

// Close the device
func (w *Webcam) Close() error {
	for _, buffer := range w.buffers {
		err := mmapReleaseBuffer(buffer)
		if err != nil {
			return err
		}
	}

	err := w.file.Close()

	return err
}

func gobytes(p unsafe.Pointer, n int) []byte {

	h := reflect.SliceHeader{uintptr(p), n, n}
	s := *(*[]byte)(unsafe.Pointer(&h))

	return s
}
