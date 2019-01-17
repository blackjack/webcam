// package snapshot is an webcam stills capture module.
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
	Width    int
	Height   int
	Format   string
	Timeout  uint32
	Buffers  uint32
	newFrame func(int, int, []byte, func()) (frame.Frame, error)
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
func (c *Snapper) Open(device, format, resolution string) error {
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
	format_desc := c.cam.GetSupportedFormats()
	var pixelFormat webcam.PixelFormat
	var found bool
	for k, v := range format_desc {
		if v == format {
			found = true
			pixelFormat = k
			break
		}
	}
	if !found {
		return fmt.Errorf("%s: unsupported format: %s", device, format)
	}
	if c.newFrame, err = frame.GetFramer(format); err != nil {
		return err
	}

	// Build a map of resolution names from the description.
	sizeMap := make(map[string]webcam.FrameSize)
	for _, value := range c.cam.GetSupportedFrameSizes(pixelFormat) {
		sizeMap[value.GetString()] = value
	}

	sz, ok := sizeMap[resolution]
	if !ok {
		return fmt.Errorf("%s: unsupported resolution: %s (allowed: %v)", device, resolution, sizeMap)
	}

	_, w, h, err := c.cam.SetImageFormat(pixelFormat, uint32(sz.MaxWidth), uint32(sz.MaxHeight))

	if err != nil {
		return err
	}
	c.Width = int(w)
	c.Height = int(h)

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
	return c.newFrame(c.Width, c.Height, snap.frame, func() {
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

// Query returns a map of the supported formats and resolutions.
func (c *Snapper) Query() map[string][]string {
	m := map[string][]string{}
	formats := c.cam.GetSupportedFormats()
	for f, fs := range formats {
		r := []string{}
		for _, value := range c.cam.GetSupportedFrameSizes(f) {
			if value.StepWidth == 0 && value.StepHeight == 0 {
				r = append(r, fmt.Sprintf("%dx%d", value.MaxWidth, value.MaxHeight))
			}
		}
		m[fs] = r
	}
	return m
}

// GetControl returns the current value of a camera control.
func (c *Snapper) GetControl(id webcam.ControlID) (int32, error) {
	return c.cam.GetControl(id)
}

// SetControl sets the selected camera control.
func (c *Snapper) SetControl(id webcam.ControlID, value int32) error {
	return c.cam.SetControl(id, value)
}
