package webcam

import (
	"bytes"
	"encoding/binary"
	"golang.org/x/sys/unix"
	"unsafe"
	"github.com/fabian-z/webcam/ioctl"
)

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

var (
	VIDIOC_QUERYCAP        = ioctl.IoR(uintptr('V'), 0, unsafe.Sizeof(v4l2_capability{}))
	VIDIOC_ENUM_FMT        = ioctl.IoRW(uintptr('V'), 2, unsafe.Sizeof(v4l2_fmtdesc{}))
	VIDIOC_S_FMT           = ioctl.IoRW(uintptr('V'), 5, unsafe.Sizeof(v4l2_format{}))
	VIDIOC_REQBUFS         = ioctl.IoRW(uintptr('V'), 8, unsafe.Sizeof(v4l2_requestbuffers{}))
	VIDIOC_QUERYBUF        = ioctl.IoRW(uintptr('V'), 9, unsafe.Sizeof(v4l2_buffer{}))
	VIDIOC_QBUF            = ioctl.IoRW(uintptr('V'), 15, unsafe.Sizeof(v4l2_buffer{}))
	VIDIOC_DQBUF           = ioctl.IoRW(uintptr('V'), 17, unsafe.Sizeof(v4l2_buffer{}))
	//sizeof int32
	VIDIOC_STREAMON        = ioctl.IoW(uintptr('V'), 18, 4)
	VIDIOC_ENUM_FRAMESIZES = ioctl.IoRW(uintptr('V'), 74, unsafe.Sizeof(v4l2_frmsizeenum{}))

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

type v4l2_format struct {
	_type uint32
	union [204]uint8
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
	union     [8]uint8
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

func getFrameSize(fd uintptr, index uint32, code uint32) (frameSize [6]uint32, err error) {

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
		err = binary.Read(bytes.NewBuffer(frmsizeenum.union[:]), binary.LittleEndian, discrete)

		if err != nil {
			return
		}

		frameSize[0] = discrete.Width
		frameSize[1] = discrete.Width
		frameSize[2] = 0
		frameSize[3] = discrete.Height
		frameSize[4] = discrete.Height
		frameSize[5] = 0

	case V4L2_FRMSIZE_TYPE_CONTINUOUS:

	case V4L2_FRMSIZE_TYPE_STEPWISE:
		stepwise := &v4l2_frmsize_stepwise{}
		err = binary.Read(bytes.NewBuffer(frmsizeenum.union[:]), binary.LittleEndian, stepwise)

		if err != nil {
			return
		}

		frameSize[0] = stepwise.Min_width
		frameSize[1] = stepwise.Max_width
		frameSize[2] = stepwise.Step_width
		frameSize[3] = stepwise.Min_height
		frameSize[4] = stepwise.Max_height
		frameSize[5] = stepwise.Step_height

	}

	return

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
	err = binary.Write(pixbytes, binary.LittleEndian, pix)

	if err != nil {
		return
	}

	copy(format.union[:], pixbytes.Bytes())

	err = ioctl.Ioctl(fd, VIDIOC_S_FMT, uintptr(unsafe.Pointer(format)))

	if err != nil {
		return
	}

	pixReverse := &v4l2_pix_format{}
	err = binary.Read(bytes.NewBuffer(format.union[:]), binary.LittleEndian, pixReverse)

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

func mmapQueryBuffer(fd uintptr, index uint32, length *uint32) (start unsafe.Pointer, err error) {

	buffer := &v4l2_buffer{}

	buffer._type = V4L2_BUF_TYPE_VIDEO_CAPTURE
	buffer.memory = V4L2_MEMORY_MMAP
	buffer.index = index

	err = ioctl.Ioctl(fd, VIDIOC_QUERYBUF, uintptr(unsafe.Pointer(buffer)))

	if err != nil {
		return
	}

	var offset uint32
	err = binary.Read(bytes.NewBuffer(buffer.union[:]), binary.LittleEndian, offset)

	if err != nil {
		return
	}

	*length = buffer.length

	returnPointer, _, errno := unix.Syscall6(unix.SYS_MMAP, 0, uintptr(buffer.length), unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED, fd, uintptr(offset))

	start = unsafe.Pointer(returnPointer)

	if errno != 0 {
		err = errno
	}

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

func mmapReleaseBuffer(start unsafe.Pointer, length uint32) (err error) {

	_, _, errno := unix.Syscall(unix.SYS_MUNMAP, uintptr(start), uintptr(length), uintptr(0))

	if errno != 0 {
		err = errno
	}

	return

}

func startStreaming(fd uintptr) (err error) {

	var uintPointer uint32 = V4L2_BUF_TYPE_VIDEO_CAPTURE
	err = ioctl.Ioctl(fd, VIDIOC_STREAMON, uintptr(unsafe.Pointer(&uintPointer)))
	return

}

func waitForFrame(fd uintptr, timeout uint32) (count int, err error) {

	for {

		fds := &unix.FdSet{}
		fds.Bits[0] = int64(fd)

		tv := &unix.Timeval{}
		tv.Sec = int64(timeout)

		count, err = unix.Select(int(fd)+1, fds, nil, nil, tv)

		if count < 0 {

			if err == unix.EINTR {
				continue
			}

			return

		}

		return

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