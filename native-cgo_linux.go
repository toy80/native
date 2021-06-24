// +build linux

package native

// #cgo linux LDFLAGS: -lxcb -lxcb-icccm
// #include <xcb/xcb.h>
// #include "native-cgo.h"
import "C"

const slowRelativeMouse = true

type NaWndData struct {
	WindowID
	btnDown int32
	mouseX  float32
	mouseY  float32
}

func DispatchEvent() int {
	return int(C.na_dispatch_event())
}
