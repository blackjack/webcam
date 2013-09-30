package webcam

// #include "webcam_wrapper.h"
import "C"
import "unsafe"

func GetImg(path string) []byte {
	dev := C.CString(path)
	defer C.free(unsafe.Pointer(dev))
	var length C.int
	buf := C.go_get_webcam_frame(dev)
	result := C.GoBytes(unsafe.Pointer(buf.start), C.int(buf.length))
	if buf.start != C.NULL {
		C.free(unsafe.Pointer(buf.start))
	}
}
