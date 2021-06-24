//+build windows,forcecgo

package native

// #cgo CFLAGS: -DUNICODE
//
// #include "native-cgo.h"
import "C"

const slowRelativeMouse = false

// func GetCursorPos(wnd uintptr) (x, y float32, err error) {
// 	win := getWindow(uintptr(wnd))
// 	if win == nil {
// 		return 0, 0, ErrGeneric
// 	}
// 	native := win.WindowID
// 	var cx, cy C.float
// 	C.na_get_mouse_pos((*C.WindowID)(unsafe.Pointer(&native)), &x, &y);
// 	x, y = float32(cx), float32(cy);
// 	return
// }

func DispatchEvent() int {
	return int(C.na_dispatch_event())
}
