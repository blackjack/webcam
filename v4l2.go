package webcam

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/blackjack/webcam/ioctl"
	"golang.org/x/sys/unix"
)

/*
	V4L2 constants, enums, and structs are derived from the header file
	'videodev2.h'. See the following reference for more information:
	https://www.kernel.org/doc/html/latest/userspace-api/media/v4l/videodev.html#videodev2-h
*/

type controlType int

const (
	c_int controlType = iota
	c_bool
	c_menu
)

type control struct {
	id     uint32
	name   string
	c_type controlType
	step   int32
	min    int32
	max    int32
}

const (
	V4L2_CAP_VIDEO_CAPTURE      uint32 = 0x00000001
	V4L2_CAP_STREAMING          uint32 = 0x04000000
	V4L2_BUF_TYPE_VIDEO_CAPTURE uint32 = 1
	V4L2_MEMORY_MMAP            uint32 = 1
	V4L2_FIELD_ANY              uint32 = 0
)

const (
	V4L2_FRMSIZE_TYPE_DISCRETE   uint32 = 1
	V4L2_FRMSIZE_TYPE_CONTINUOUS uint32 = 2
	V4L2_FRMSIZE_TYPE_STEPWISE   uint32 = 3
)

const (
	V4L2_FRMIVAL_TYPE_DISCRETE   uint32 = 1
	V4L2_FRMIVAL_TYPE_CONTINUOUS uint32 = 2
	V4L2_FRMIVAL_TYPE_STEPWISE   uint32 = 3
)

const (
	V4L2_CID_BASE               uint32 = 0x00980900
	V4L2_CID_AUTO_WHITE_BALANCE uint32 = V4L2_CID_BASE + 12
	V4L2_CID_PRIVATE_BASE       uint32 = 0x08000000
)

const (
	V4L2_CTRL_TYPE_INTEGER      uint32 = 1
	V4L2_CTRL_TYPE_BOOLEAN      uint32 = 2
	V4L2_CTRL_TYPE_MENU         uint32 = 3
	V4L2_CTRL_TYPE_BUTTON       uint32 = 4
	V4L2_CTRL_TYPE_INTEGER64    uint32 = 5
	V4L2_CTRL_TYPE_CTRL_CLASS   uint32 = 6
	V4L2_CTRL_TYPE_STRING       uint32 = 7
	V4L2_CTRL_TYPE_BITMASK      uint32 = 8
	V4L2_CTRL_TYPE_INTEGER_MENU uint32 = 9

	V4L2_CTRL_COMPOUND_TYPES uint32 = 0x0100
	V4L2_CTRL_TYPE_U8        uint32 = 0x0100
	V4L2_CTRL_TYPE_U16       uint32 = 0x0101
	V4L2_CTRL_TYPE_U32       uint32 = 0x0102
)

const (
	V4L2_CTRL_FLAG_DISABLED  uint32 = 0x00000001
	V4L2_CTRL_FLAG_NEXT_CTRL uint32 = 0x80000000
)

var (
	VIDIOC_QUERYCAP  = ioctl.IoR(uintptr('V'), 0, unsafe.Sizeof(v4l2_capability{}))
	VIDIOC_ENUM_FMT  = ioctl.IoRW(uintptr('V'), 2, unsafe.Sizeof(v4l2_fmtdesc{}))
	VIDIOC_S_FMT     = ioctl.IoRW(uintptr('V'), 5, unsafe.Sizeof(v4l2_format{}))
	VIDIOC_REQBUFS   = ioctl.IoRW(uintptr('V'), 8, unsafe.Sizeof(v4l2_requestbuffers{}))
	VIDIOC_QUERYBUF  = ioctl.IoRW(uintptr('V'), 9, unsafe.Sizeof(v4l2_buffer{}))
	VIDIOC_QBUF      = ioctl.IoRW(uintptr('V'), 15, unsafe.Sizeof(v4l2_buffer{}))
	VIDIOC_DQBUF     = ioctl.IoRW(uintptr('V'), 17, unsafe.Sizeof(v4l2_buffer{}))
	VIDIOC_G_PARM    = ioctl.IoRW(uintptr('V'), 21, unsafe.Sizeof(v4l2_streamparm{}))
	VIDIOC_S_PARM    = ioctl.IoRW(uintptr('V'), 22, unsafe.Sizeof(v4l2_streamparm{}))
	VIDIOC_G_CTRL    = ioctl.IoRW(uintptr('V'), 27, unsafe.Sizeof(v4l2_control{}))
	VIDIOC_S_CTRL    = ioctl.IoRW(uintptr('V'), 28, unsafe.Sizeof(v4l2_control{}))
	VIDIOC_QUERYCTRL = ioctl.IoRW(uintptr('V'), 36, unsafe.Sizeof(v4l2_queryctrl{}))
	//sizeof int32
	VIDIOC_STREAMON            = ioctl.IoW(uintptr('V'), 18, 4)
	VIDIOC_STREAMOFF           = ioctl.IoW(uintptr('V'), 19, 4)
	VIDIOC_G_INPUT             = ioctl.IoR(uintptr('V'), 38, 4)
	VIDIOC_S_INPUT             = ioctl.IoRW(uintptr('V'), 39, 4)
	VIDIOC_ENUM_FRAMESIZES     = ioctl.IoRW(uintptr('V'), 74, unsafe.Sizeof(v4l2_frmsizeenum{}))
	VIDIOC_ENUM_FRAMEINTERVALS = ioctl.IoRW(uintptr('V'), 75, unsafe.Sizeof(v4l2_frmivalenum{}))
	__p                        = unsafe.Pointer(uintptr(0))
	NativeByteOrder            = getNativeByteOrder()
)

type v4l2_capability struct {
	driver       [16]uint8
	card         [32]uint8
	bus_info     [32]uint8
	version      uint32
	capabilities uint32
	device_caps  uint32
	reserved     [3]uint32
}

type v4l2_fmtdesc struct {
	index       uint32
	_type       uint32
	flags       uint32
	description [32]uint8
	pixelformat uint32
	reserved    [4]uint32
}

type v4l2_frmsizeenum struct {
	index        uint32
	pixel_format uint32
	_type        uint32
	union        [24]uint8
	reserved     [2]uint32
}

type v4l2_frmsize_discrete struct {
	Width  uint32
	Height uint32
}

type v4l2_frmsize_stepwise struct {
	Min_width   uint32
	Max_width   uint32
	Step_width  uint32
	Min_height  uint32
	Max_height  uint32
	Step_height uint32
}

// v4l2_frmivalenum that contains a pixel format and size and receives a
// frame interval (time duration between frames of a video stream).
type v4l2_frmivalenum struct {
	index        uint32
	pixel_format uint32
	width        uint32
	height       uint32
	_type        uint32
	union        [24]uint8
	reserved     [2]uint32
}

// v4l2_frmival_stepwise represents the frame interval range
// as minimum, maximum, and step size intervals.
type v4l2_frmival_stepwise struct {
	min  v4l2_fract // minimum frame interval [s]
	max  v4l2_fract // maximum frame interval [s]
	step v4l2_fract // frame interval step size [s]
}

// Hack to make go compiler properly align union
type v4l2_format_aligned_union struct {
	data [200 - unsafe.Sizeof(__p)]byte
	_    unsafe.Pointer
}

type v4l2_format struct {
	_type uint32
	union v4l2_format_aligned_union
}

type v4l2_pix_format struct {
	Width        uint32
	Height       uint32
	Pixelformat  uint32
	Field        uint32
	Bytesperline uint32
	Sizeimage    uint32
	Colorspace   uint32
	Priv         uint32
	Flags        uint32
	Ycbcr_enc    uint32
	Quantization uint32
	Xfer_func    uint32
}

type v4l2_requestbuffers struct {
	count    uint32
	_type    uint32
	memory   uint32
	reserved [2]uint32
}

type v4l2_buffer struct {
	index     uint32
	_type     uint32
	bytesused uint32
	flags     uint32
	field     uint32
	timestamp unix.Timeval
	timecode  v4l2_timecode
	sequence  uint32
	memory    uint32
	union     [unsafe.Sizeof(__p)]uint8
	length    uint32
	reserved2 uint32
	reserved  uint32
}

type v4l2_timecode struct {
	_type    uint32
	flags    uint32
	frames   uint8
	seconds  uint8
	minutes  uint8
	hours    uint8
	userbits [4]uint8
}

type v4l2_queryctrl struct {
	id            uint32
	_type         uint32
	name          [32]uint8
	minimum       int32
	maximum       int32
	step          int32
	default_value int32
	flags         uint32
	reserved      [2]uint32
}

type v4l2_control struct {
	id    uint32
	value int32
}

type v4l2_fract struct {
	Numerator   uint32
	Denominator uint32
}

type v4l2_streamparm_union struct {
	capability     uint32
	output_mode    uint32
	time_per_frame v4l2_fract
	extended_mode  uint32
	buffers        uint32
	reserved       [4]uint32
	data           [200 - (10 * unsafe.Sizeof(uint32(0)))]byte
}

type v4l2_streamparm struct {
	_type uint32
	union v4l2_streamparm_union
}

func checkCapabilities(fd uintptr) (supportsVideoCapture bool, supportsVideoStreaming bool, err error) {

	caps := &v4l2_capability{}

	err = ioctl.Ioctl(fd, VIDIOC_QUERYCAP, uintptr(unsafe.Pointer(caps)))

	if err != nil {
		return
	}

	supportsVideoCapture = (caps.capabilities & V4L2_CAP_VIDEO_CAPTURE) != 0
	supportsVideoStreaming = (caps.capabilities & V4L2_CAP_STREAMING) != 0
	return

}

func getPixelFormat(fd uintptr, index uint32) (code uint32, description string, err error) {

	fmtdesc := &v4l2_fmtdesc{}

	fmtdesc.index = index
	fmtdesc._type = V4L2_BUF_TYPE_VIDEO_CAPTURE

	err = ioctl.Ioctl(fd, VIDIOC_ENUM_FMT, uintptr(unsafe.Pointer(fmtdesc)))

	if err != nil {
		return
	}

	code = fmtdesc.pixelformat
	description = CToGoString(fmtdesc.description[:])

	return
}

func getFrameSize(fd uintptr, index uint32, code uint32) (frameSize FrameSize, err error) {

	frmsizeenum := &v4l2_frmsizeenum{}
	frmsizeenum.index = index
	frmsizeenum.pixel_format = code

	err = ioctl.Ioctl(fd, VIDIOC_ENUM_FRAMESIZES, uintptr(unsafe.Pointer(frmsizeenum)))

	if err != nil {
		return
	}

	switch frmsizeenum._type {

	case V4L2_FRMSIZE_TYPE_DISCRETE:
		discrete := &v4l2_frmsize_discrete{}
		err = binary.Read(bytes.NewBuffer(frmsizeenum.union[:]), NativeByteOrder, discrete)

		if err != nil {
			return
		}

		frameSize.MinWidth = discrete.Width
		frameSize.MaxWidth = discrete.Width
		frameSize.StepWidth = 0
		frameSize.MinHeight = discrete.Height
		frameSize.MaxHeight = discrete.Height
		frameSize.StepHeight = 0

	case V4L2_FRMSIZE_TYPE_CONTINUOUS:

	case V4L2_FRMSIZE_TYPE_STEPWISE:
		stepwise := &v4l2_frmsize_stepwise{}
		err = binary.Read(bytes.NewBuffer(frmsizeenum.union[:]), NativeByteOrder, stepwise)

		if err != nil {
			return
		}

		frameSize.MinWidth = stepwise.Min_width
		frameSize.MaxWidth = stepwise.Max_width
		frameSize.StepWidth = stepwise.Step_width
		frameSize.MinHeight = stepwise.Min_height
		frameSize.MaxHeight = stepwise.Max_height
		frameSize.StepHeight = stepwise.Step_height
	}

	return
}

func getName(fd uintptr) (string, error) {
	var caps v4l2_capability
	if err := ioctl.Ioctl(fd, VIDIOC_QUERYCAP, uintptr(unsafe.Pointer(&caps))); err != nil {
		return "", err
	}

	return CToGoString(caps.card[:]), nil
}

func getFrameInterval(fd uintptr, index uint32, code uint32, width uint32, height uint32) (FrameRate, error) {
	frmivalEnum := &v4l2_frmivalenum{
		index:        index,
		pixel_format: code,
		width:        width,
		height:       height,
	}

	if err := ioctl.Ioctl(fd, VIDIOC_ENUM_FRAMEINTERVALS, uintptr(unsafe.Pointer(frmivalEnum))); err != nil {
		return FrameRate{}, err
	}

	switch frmivalEnum._type {

	case V4L2_FRMIVAL_TYPE_DISCRETE:
		discrete := &v4l2_fract{}
		if err := binary.Read(bytes.NewBuffer(frmivalEnum.union[:]), NativeByteOrder, discrete); err != nil {
			return FrameRate{}, err
		}
		return FrameRate{
			MinDenominator:  discrete.Denominator,
			MaxDenominator:  discrete.Denominator,
			StepDenominator: 0,
			MinNumerator:    discrete.Numerator,
			MaxNumerator:    discrete.Numerator,
			StepNumerator:   0,
		}, nil

	case V4L2_FRMIVAL_TYPE_CONTINUOUS:
		return FrameRate{}, fmt.Errorf("V4L2_FRMIVAL_TYPE_CONTINUOUS not implemented")

	case V4L2_FRMIVAL_TYPE_STEPWISE:
		stepwise := &v4l2_frmival_stepwise{}
		if err := binary.Read(bytes.NewBuffer(frmivalEnum.union[:]), NativeByteOrder, stepwise); err != nil {
			return FrameRate{}, err
		}
		return FrameRate{
			MinDenominator:  stepwise.min.Denominator,
			MaxDenominator:  stepwise.max.Denominator,
			StepDenominator: stepwise.step.Denominator,
			MinNumerator:    stepwise.min.Numerator,
			MaxNumerator:    stepwise.max.Numerator,
			StepNumerator:   stepwise.step.Numerator,
		}, nil
	}

	return FrameRate{}, fmt.Errorf("unknown frame interval type")
}

func getBusInfo(fd uintptr) (string, error) {
	var caps v4l2_capability
	if err := ioctl.Ioctl(fd, VIDIOC_QUERYCAP, uintptr(unsafe.Pointer(&caps))); err != nil {
		return "", err
	}

	return CToGoString(caps.bus_info[:]), nil
}

func setImageFormat(fd uintptr, formatcode *uint32, width *uint32, height *uint32) (err error) {

	format := &v4l2_format{
		_type: V4L2_BUF_TYPE_VIDEO_CAPTURE,
	}

	pix := v4l2_pix_format{
		Width:       *width,
		Height:      *height,
		Pixelformat: *formatcode,
		Field:       V4L2_FIELD_ANY,
	}

	pixbytes := &bytes.Buffer{}
	err = binary.Write(pixbytes, NativeByteOrder, pix)

	if err != nil {
		return
	}

	copy(format.union.data[:], pixbytes.Bytes())

	err = ioctl.Ioctl(fd, VIDIOC_S_FMT, uintptr(unsafe.Pointer(format)))

	if err != nil {
		return
	}

	pixReverse := &v4l2_pix_format{}
	err = binary.Read(bytes.NewBuffer(format.union.data[:]), NativeByteOrder, pixReverse)

	if err != nil {
		return
	}

	*width = pixReverse.Width
	*height = pixReverse.Height
	*formatcode = pixReverse.Pixelformat

	return

}

func mmapRequestBuffers(fd uintptr, buf_count *uint32) (err error) {

	req := &v4l2_requestbuffers{}
	req.count = *buf_count
	req._type = V4L2_BUF_TYPE_VIDEO_CAPTURE
	req.memory = V4L2_MEMORY_MMAP

	err = ioctl.Ioctl(fd, VIDIOC_REQBUFS, uintptr(unsafe.Pointer(req)))

	if err != nil {
		return
	}

	*buf_count = req.count

	return

}

func mmapQueryBuffer(fd uintptr, index uint32, length *uint32) (buffer []byte, err error) {

	req := &v4l2_buffer{}

	req._type = V4L2_BUF_TYPE_VIDEO_CAPTURE
	req.memory = V4L2_MEMORY_MMAP
	req.index = index

	err = ioctl.Ioctl(fd, VIDIOC_QUERYBUF, uintptr(unsafe.Pointer(req)))

	if err != nil {
		return
	}

	var offset uint32
	err = binary.Read(bytes.NewBuffer(req.union[:]), NativeByteOrder, &offset)

	if err != nil {
		return
	}

	*length = req.length

	buffer, err = unix.Mmap(int(fd), int64(offset), int(req.length), unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	return
}

func mmapDequeueBuffer(fd uintptr, index *uint32, length *uint32) (err error) {

	buffer := &v4l2_buffer{}

	buffer._type = V4L2_BUF_TYPE_VIDEO_CAPTURE
	buffer.memory = V4L2_MEMORY_MMAP

	err = ioctl.Ioctl(fd, VIDIOC_DQBUF, uintptr(unsafe.Pointer(buffer)))

	if err != nil {
		return
	}

	*index = buffer.index
	*length = buffer.bytesused

	return

}

func mmapEnqueueBuffer(fd uintptr, index uint32) (err error) {

	buffer := &v4l2_buffer{}

	buffer._type = V4L2_BUF_TYPE_VIDEO_CAPTURE
	buffer.memory = V4L2_MEMORY_MMAP
	buffer.index = index

	err = ioctl.Ioctl(fd, VIDIOC_QBUF, uintptr(unsafe.Pointer(buffer)))
	return

}

func mmapReleaseBuffer(buffer []byte) (err error) {
	err = unix.Munmap(buffer)
	return
}

func startStreaming(fd uintptr) (err error) {

	var uintPointer uint32 = V4L2_BUF_TYPE_VIDEO_CAPTURE
	err = ioctl.Ioctl(fd, VIDIOC_STREAMON, uintptr(unsafe.Pointer(&uintPointer)))
	return

}

func stopStreaming(fd uintptr) (err error) {

	var uintPointer uint32 = V4L2_BUF_TYPE_VIDEO_CAPTURE
	err = ioctl.Ioctl(fd, VIDIOC_STREAMOFF, uintptr(unsafe.Pointer(&uintPointer)))
	return

}

func waitForFrame(fd uintptr, timeout uint32) (count int, err error) {

	for {
		fds := &unix.FdSet{}
		fds.Set(int(fd))

		var oneSecInNsec int64 = 1e9
		timeoutNsec := int64(timeout) * oneSecInNsec
		nativeTimeVal := unix.NsecToTimeval(timeoutNsec)
		tv := &nativeTimeVal

		count, err = unix.Select(int(fd+1), fds, nil, nil, tv)

		if count < 0 && err == unix.EINTR {
			continue
		}
		return
	}

}

func getControl(fd uintptr, id uint32) (int32, error) {
	ctrl := &v4l2_control{}
	ctrl.id = id
	err := ioctl.Ioctl(fd, VIDIOC_G_CTRL, uintptr(unsafe.Pointer(ctrl)))
	return ctrl.value, err
}

func setControl(fd uintptr, id uint32, val int32) error {
	ctrl := &v4l2_control{}
	ctrl.id = id
	ctrl.value = val
	return ioctl.Ioctl(fd, VIDIOC_S_CTRL, uintptr(unsafe.Pointer(ctrl)))
}

func getInput(fd uintptr) (index int32, err error) {
	err = ioctl.Ioctl(fd, VIDIOC_G_INPUT, uintptr(unsafe.Pointer(&index)))
	return
}

func selectInput(fd uintptr, index uint32) (err error) {
	err = ioctl.Ioctl(fd, VIDIOC_S_INPUT, uintptr(unsafe.Pointer(&index)))
	return
}

func getFramerate(fd uintptr) (float32, error) {
	param := &v4l2_streamparm{}
	param._type = V4L2_BUF_TYPE_VIDEO_CAPTURE

	err := ioctl.Ioctl(fd, VIDIOC_G_PARM, uintptr(unsafe.Pointer(param)))
	if err != nil {
		return 0, err
	}
	tf := param.union.time_per_frame
	if tf.Denominator == 0 || tf.Numerator == 0 {
		return 0, fmt.Errorf("Invalid framerate (%d/%d)", tf.Denominator, tf.Numerator)
	}
	return float32(tf.Denominator) / float32(tf.Numerator), nil
}

func setFramerate(fd uintptr, num, denom uint32) error {
	param := &v4l2_streamparm{}
	param._type = V4L2_BUF_TYPE_VIDEO_CAPTURE
	param.union.time_per_frame.Numerator = num
	param.union.time_per_frame.Denominator = denom
	return ioctl.Ioctl(fd, VIDIOC_S_PARM, uintptr(unsafe.Pointer(param)))
}

func queryControls(fd uintptr) []control {
	controls := []control{}
	var err error
	// Don't use V42L_CID_BASE since it is the same as brightness.
	var id uint32
	for err == nil {
		id |= V4L2_CTRL_FLAG_NEXT_CTRL
		query := &v4l2_queryctrl{}
		query.id = id
		err = ioctl.Ioctl(fd, VIDIOC_QUERYCTRL, uintptr(unsafe.Pointer(query)))
		id = query.id
		if err == nil {
			if (query.flags & V4L2_CTRL_FLAG_DISABLED) != 0 {
				continue
			}
			var c control
			switch query._type {
			default:
				continue
			case V4L2_CTRL_TYPE_INTEGER, V4L2_CTRL_TYPE_INTEGER64:
				c.c_type = c_int
			case V4L2_CTRL_TYPE_BOOLEAN:
				c.c_type = c_bool
			case V4L2_CTRL_TYPE_MENU:
				c.c_type = c_menu
			}
			c.id = id
			c.name = CToGoString(query.name[:])
			c.min = query.minimum
			c.max = query.maximum
			c.step = query.step
			controls = append(controls, c)
		}
	}
	return controls
}

func getNativeByteOrder() binary.ByteOrder {
	var i int32 = 0x01020304
	u := unsafe.Pointer(&i)
	pb := (*byte)(u)
	b := *pb
	if b == 0x04 {
		return binary.LittleEndian
	} else {
		return binary.BigEndian
	}
}

func CToGoString(c []byte) string {
	n := -1
	for i, b := range c {
		if b == 0 {
			break
		}
		n = i
	}
	return string(c[:n+1])
}
