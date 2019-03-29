// package snapshot is a webcam stills capture module.
package snapshot

import (
	"fmt"

	"github.com/aamcrae/imageserver/frame"
	"github.com/aamcrae/webcam"
)

const (
	defaultTimeout = 5
	defaultBuffers = 16
)

type snap struct {
	frame []byte
	index uint32
}

type Snapper struct {
	cam      *webcam.Webcam
	Timeout  uint32
	Buffers  uint32
	framer func([]byte, func()) (frame.Frame, error)
	stop     chan struct{}
	stream   chan snap
}

// NewSnapper creates a new Snapper.
func NewSnapper() *Snapper {
	return &Snapper{Timeout: defaultTimeout, Buffers: defaultBuffers}
}

// Close releases all current frames and shuts down the webcam.
func (c *Snapper) Close() {
	if c.cam != nil {
		c.stop <- struct{}{}
		// Flush any remaining frames.
		for f := range c.stream {
			c.cam.ReleaseFrame(f.index)
		}
		c.cam.StopStreaming()
		c.cam.Close()
		c.cam = nil
	}
}

// Open initialises the webcam ready for use, and begins streaming.
func (c *Snapper) Open(device string, format frame.FourCC, w, h int) error {
	pf, err := frame.FourCCToPixelFormat(format)
	if err != nil {
		return err
	}
	if c.cam != nil {
		c.Close()
	}
	cam, err := webcam.Open(device)
	if err != nil {
		return err
	}
	c.cam = cam
	c.stop = make(chan struct{}, 1)
	c.stream = make(chan snap, 0)
	// Get the supported formats and their descriptions.
	mf := c.cam.GetSupportedFormats()
	_, ok := mf[pf]
	if !ok {
		return fmt.Errorf("%s: unsupported format: %s", device, format)
	}
	var found bool
	for _, value := range c.cam.GetSupportedFrameSizes(pf) {
		if Match(value, w, h) {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("%s: unsupported resolution: %dx%d", device, w, h)
	}
	npf, nw, nh, stride, err := c.cam.SetImageFormat(pf, uint32(w), uint32(h))

	if err != nil {
		return err
	}
	if npf != pf || w != int(nw) || h != int(nh) {
		fmt.Printf("Asked for %08x %dx%d, got %08x %dx%d\n", pf, w, h, npf, nw, nh)
	}
	if c.framer, err = frame.GetFramer(format, w, h, int(stride)); err != nil {
		return err
	}

	c.cam.SetBufferCount(c.Buffers)
	c.cam.SetAutoWhiteBalance(true)
	if err := c.cam.StartStreaming(); err != nil {
		return err
	}
	go c.capture()
	return nil
}

// Snap returns one frame from the camera.
func (c *Snapper) Snap() (frame.Frame, error) {
	snap, ok := <-c.stream
	if !ok {
		return nil, fmt.Errorf("No frame received")
	}
	return c.framer(snap.frame, func() {
		c.cam.ReleaseFrame(snap.index)
	})
}

// capture continually reads frames and either discards them or
// sends them to a channel that is ready to receive them.
func (c *Snapper) capture() {
	for {
		err := c.cam.WaitForFrame(c.Timeout)

		switch err.(type) {
		case nil:
		case *webcam.Timeout:
			continue
		default:
			panic(err)
		}

		frame, index, err := c.cam.GetFrame()
		if err != nil {
			panic(err)
		}
		select {
		// Only executed if stream is ready to receive.
		case c.stream <- snap{frame, index}:
		// Signal to stop streaming.
		case <-c.stop:
			// Finish up.
			c.cam.ReleaseFrame(index)
			close(c.stream)
			return
		default:
			c.cam.ReleaseFrame(index)
		}
	}
}

// GetControl returns the current value of a camera control.
func (c *Snapper) GetControl(id webcam.ControlID) (int32, error) {
	return c.cam.GetControl(id)
}

// SetControl sets the selected camera control.
func (c *Snapper) SetControl(id webcam.ControlID, value int32) error {
	return c.cam.SetControl(id, value)
}

// Return true if frame size can accomodate request.
func Match(fs webcam.FrameSize, w, h int) bool {
	return canFit(fs.MinWidth, fs.MaxWidth, fs.StepWidth, uint32(w)) &&
		   canFit(fs.MinHeight, fs.MaxHeight, fs.StepHeight, uint32(h))
}

func canFit(min, max, step, val uint32) bool {
	// Fixed size exact match.
	if min == max && step == 0 && val == min {
		return true
	}
	return step != 0 && val >= val && val <= max && ((val - min) % step) == 0
}
