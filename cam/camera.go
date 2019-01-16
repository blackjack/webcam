// package cam is an snapshot interface to a webcam.
// Still frames are returned as an image.Image.
package cam

import (
	"fmt"

	"github.com/aamcrae/webcam"
)

type snapshot struct {
	frame []byte
	index uint32
}

type Camera struct {
	cam      *webcam.Webcam
	Width    int
	Height   int
	Format   string
	Timeout  uint32
	newFrame func(int, int, []byte, func()) (Frame, error)
	stop     chan struct{}
	stream   chan snapshot
}

// OpenCamera opens the webcam and creates the channels ready for use.
func OpenCamera(name string) (*Camera, error) {
	c, err := webcam.Open(name)
	if err != nil {
		return nil, err
	}
	camera := &Camera{cam: c, Timeout: 5}
	camera.stop = make(chan struct{}, 1)
	camera.stream = make(chan snapshot, 0)
	return camera, nil
}

// Close releases all current frames and shuts down the webcam.
func (c *Camera) Close() {
	c.stop <- struct{}{}
	// Flush any remaining frames.
	for f := range c.stream {
		c.cam.ReleaseFrame(f.index)
	}
	c.cam.StopStreaming()
	c.cam.Close()
}

// Init initialises the webcam ready for use, and begins streaming.
func (c *Camera) Init(format string, resolution string) error {
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
		return fmt.Errorf("Camera does not support this format: %s", format)
	}
	var err error
	if c.newFrame, err = GetFramer(format); err != nil {
		return err
	}

	// Build a map of resolution names from the description.
	sizeMap := make(map[string]webcam.FrameSize)
	for _, value := range c.cam.GetSupportedFrameSizes(pixelFormat) {
		sizeMap[value.GetString()] = value
	}

	sz, ok := sizeMap[resolution]
	if !ok {
		return fmt.Errorf("Unsupported resolution: %s (allowed: %v)", resolution, sizeMap)
	}

	_, w, h, err := c.cam.SetImageFormat(pixelFormat, uint32(sz.MaxWidth), uint32(sz.MaxHeight))

	if err != nil {
		return err
	}
	c.Width = int(w)
	c.Height = int(h)

	c.cam.SetBufferCount(16)
	c.cam.SetAutoWhiteBalance(true)
	if err := c.cam.StartStreaming(); err != nil {
		return err
	}
	go c.capture()
	return nil
}

// GetFrame returns one frame from the camera.
func (c *Camera) GetFrame() (Frame, error) {
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
func (c *Camera) capture() {
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
		case c.stream <- snapshot{frame, index}:
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
func (c *Camera) Query() map[string][]string {
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
func (c *Camera) GetControl(name string) (int32, error) {
	id, err := getControlID(name)
	if err != nil {
		return 0, err
	}
	return c.cam.GetControl(id)
}

// SetControl sets the selected camera control.
func (c *Camera) SetControl(name string, value int32) error {
	id, err := getControlID(name)
	if err != nil {
		return err
	}
	return c.cam.SetControl(id, value)
}

// getControlID returns the appropriate ControlID for a user-friendly control name.
func getControlID(name string) (webcam.ControlID, error) {
	var controls map[string]webcam.ControlID = map[string]webcam.ControlID{
		"focus":                0x009a090a,
		"power_line_frequency": 0x00980918,
		"brightness":           0x00980900,
		"contrast":             0x00980901,
	}
	id, ok := controls[name]
	if !ok {
		return 0, fmt.Errorf("%s: unknown control")
	}
	return id, nil
}
