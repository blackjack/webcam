package webcam

import "fmt"

type PixelFormat uint32

// Struct that describes frame size supported by a webcam
// For fixed sizes min and max values will be the same and
// step value will be equal to '0'
type FrameSize struct {
	MinWidth  uint32
	MaxWidth  uint32
	StepWidth uint32

	MinHeight  uint32
	MaxHeight  uint32
	StepHeight uint32
}

func (s FrameSize) GetString() string {
	if s.StepWidth == 0 && s.StepHeight == 0 {
		return fmt.Sprintf("%dx%d", s.MaxWidth, s.MaxHeight)
	} else {
		return fmt.Sprintf("[%d-%d;%d]x[%d-%d;%d]", s.MinWidth, s.MaxWidth, s.StepWidth, s.MinHeight, s.MaxHeight, s.StepHeight)
	}
}
