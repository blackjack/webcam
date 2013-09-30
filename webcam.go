package webcam

// #include "webcam_wrapper.h"
import "C"
import "unsafe"

func GetImg(path string) []byte {
	dev := C.CString(path)
	var length C.int
	arr := C.go_get_webcam_frame(dev, &length)
	return C.GoBytes(unsafe.Pointer(arr), length)
}
