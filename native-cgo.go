// +build linux darwin forcecgo,windows

package native

// #cgo CFLAGS: -DTOY80_CGO
//
// #include <stdlib.h>
// #include <string.h>
// #include "native-cgo.h"
import "C"

import (
	"errors"
	"log"
	"strings"
	"unsafe"

	"github.com/toy80/debug"
)

func initNative() {
	if C.na_init_library(C.uintptr_t(unsafe.Sizeof(*(*NaWndData)(nil)))) == 0 {
		panic("C.na_init_library() failure")
	}
	if WinHintResizable != C.NA_WIN_HINT_RESIZABLE {
		panic("WinHintResizable != C.NA_WIN_HINT_RESIZABLE")
	}
	if WinHintFullScreen != C.NA_WIN_HINT_FULLSCREEN {
		panic("WinHintFullScreen != C.NA_WIN_HINT_FULLSCREEN")
	}
	if WinHintMinimizeBox != C.NA_WIN_HINT_MINIMIZE_BOX {
		panic("WinHintMinimizeBox != C.NA_WIN_HINT_MINIMIZE_BOX")
	}
	if WinHintMaximizeBox != C.NA_WIN_HINT_MAXIMIZE_BOX {
		panic("WinHintMaximizeBox != C.NA_WIN_HINT_MAXIMIZE_BOX")
	}

	if MouseBtnLeft != C.NA_MOUSE_BTN_LEFT {
		panic("MouseBtnLeft != C.NA_MOUSE_BTN_LEFT")
	}
	if MouseBtnRight != C.NA_MOUSE_BTN_RIGHT {
		panic("MouseBtnRight != C.NA_MOUSE_BTN_RIGHT")
	}
	if MouseBtnMiddle != C.NA_MOUSE_BTN_MIDDLE {
		panic("MouseBtnMiddle != C.NA_MOUSE_BTN_MIDDLE")
	}
}

// IsMainThread reports whether current thread is the main thread
func IsMainThread() bool {
	return C.na_is_main_thread() != 0
}

func GetCursorPos(wnd uintptr) (x, y float32, err error) {
	win := getWindow(uintptr(wnd))
	if win == nil {
		return 0, 0, ErrGeneric
	}
	// TODO: 从win.mouseX读鼠标位置是权宜之计, 因为Linux的na_get_mouse_pos暂未实现,
	// 所以所有cgo分支的SysSpecial都记下了鼠标位置.
	x, y = win.mouseX, win.mouseY
	return
}

func GetScreenSize(id int) (width, height float32, err error) {
	// TODO:
	return float32(1920), float32(1080), nil
}

func (win *NaWindow) sysCreateWindow(hints uint64, width, height float32) (err error) {
	x, y, w, h := int32(0), int32(0), int32(width), int32(height)
	if sw, sh, err1 := GetScreenSize(0); err1 == nil {
		if w <= 0 {
			w = int32(sw) / 2
		}
		if h <= 0 {
			h = int32(sh) / 2
		}
		x, y = (int32(sw)-w)/2, (int32(sh)-h)/2
	}
	if w <= 0 {
		w = 100
	}
	if h <= 0 {
		h = 100
	}

	sz := unsafe.Sizeof(*(*NaWndData)(nil))
	ptr := C.malloc(C.size_t(sz))
	C.memset(ptr, 0, C.size_t(sz))
	win.NaWndData = (*NaWndData)(ptr)
	if C.na_create_window((*C.NaWndData)(unsafe.Pointer(win.NaWndData)), C.uint64_t(hints), C.int(x), C.int(y), C.int(w), C.int(h)) == 0 {
		err = errors.New("failed to create native window")
		C.free(unsafe.Pointer(win.NaWndData))
		win.NaWndData = nil
		return
	}
	win.width, win.height = width, height
	return
}

func (win *NaWindow) destroyNativeWindow() error {
	debug.Trace(0)
	C.na_destroy_window((*C.NaWndData)(unsafe.Pointer(win.NaWndData)))
	return nil
}

func (win *NaWindow) releaseOtherNativeResource() {
	C.free(unsafe.Pointer(win.NaWndData))
	win.NaWndData = nil
}

func (win *NaWindow) SetVisible(visible bool) {
	var v C.int
	if visible {
		v = 1
	} else {
		v = 0
	}
	C.na_show_window((*C.NaWndData)(unsafe.Pointer(win.NaWndData)), v)
}

func (win *NaWindow) ExposeRect(x, y, width, height float32) {
	// TODO:
}

func YieldThread() {
	//debug.Println("Yield()")
	C.na_yield()
}

// extern NaWndData * na_get_windata(uintptr_t win);

//export na_get_windata
func na_get_windata(wnd C.uintptr_t) *C.NaWndData {
	win := getWindow(uintptr(wnd))
	if win == nil {
		return nil
	}
	return (*C.NaWndData)(unsafe.Pointer(win.NaWndData))
}

//export na_report
func na_report(msg *C.char, topanic C.int) {
	s := strings.TrimSpace(C.GoString(msg))
	log.Println("na:", s)
	if topanic != 0 {
		panic(s)
	}
}

//export na_emit_close
func na_emit_close(wnd C.uintptr_t) {
	win := getWindow(uintptr(wnd))
	if win != nil {
		win.onNativeRequestDestroy()
	}
}

//export na_emit_expose
func na_emit_expose(wnd C.uintptr_t, x, y, width, height C.int) {
	win := getWindow(uintptr(wnd))
	if win != nil {
		win.emitExpose(float32(x), float32(y), float32(width), float32(height))
		// if runtime.GOOS == "darwin" { // TODO: better handle timer
		// 	win.emitTimerTick()
		// }
	}
}

//export na_emit_destroy
func na_emit_destroy(wnd C.uintptr_t) {
	win := getWindow(uintptr(wnd))
	if win != nil {
		win.onNativeDestroy()
	}
}

//export na_emit_mouse_press
func na_emit_mouse_press(wnd C.uintptr_t, btn C.int, x, y C.float) {
	win := getWindow(uintptr(wnd))
	if win != nil {
		win.emitMousePress(int(btn), float32(x), float32(y))
	}
}

//export na_emit_mouse_release
func na_emit_mouse_release(wnd C.uintptr_t, btn C.int, x, y C.float) {
	win := getWindow(uintptr(wnd))
	if win != nil {
		win.emitMouseRelease(int(btn), float32(x), float32(y))
	}
}

//export na_emit_mouse_wheel
func na_emit_mouse_wheel(wnd C.uintptr_t, vertical C.int, delta, x, y C.float) {
	win := getWindow(uintptr(wnd))
	if win != nil {
		win.emitMouseWheel(int(vertical), float32(delta), float32(x), float32(y))
	}
}

//export na_emit_mouse_move
func na_emit_mouse_move(wnd C.uintptr_t, x, y C.float) {
	win := getWindow(uintptr(wnd))
	if win != nil {
		if slowRelativeMouse {
			// TODO: 目前只有Windows 版用了Raw Input加速, 其他平台只能从普通的Mouse Move生成
			dx, dy := float32(x)-win.mouseX, float32(y)-win.mouseY
			win.emitMouseMoveRelative(dx, dy)
		}
		win.mouseX, win.mouseY = float32(x), float32(y)
		win.emitMouseMove(float32(x), float32(y))
	}
}

//export na_emit_mouse_move_relative
func na_emit_mouse_move_relative(wnd C.uintptr_t, dx, dy C.float) {
	win := getWindow(uintptr(wnd))
	if win != nil {
		win.emitMouseMoveRelative(float32(dx), float32(dy))
	}
}

//export na_emit_mouse_enter
func na_emit_mouse_enter(wnd C.uintptr_t, x, y C.float) {
	// TODO:
	// win := getWindow(uintptr(wnd))
	// if win != nil {
	// }
}

//export na_emit_mouse_leave
func na_emit_mouse_leave(wnd C.uintptr_t, x, y C.float) {
	// TODO:
	// win := getWindow(uintptr(wnd))
	// if win != nil {
	// }
}

//export na_emit_resize
func na_emit_resize(wnd C.uintptr_t, width, height C.float) {
	win := getWindow(uintptr(wnd))
	if win != nil {
		win.emitResize(float32(width), float32(height))
	}
}

//export na_emit_appear
func na_emit_appear(wnd C.uintptr_t, visible C.int) {
	win := getWindow(uintptr(wnd))
	if win != nil {
		if visible == 0 {
			win.emitAppear(false)
		} else {
			win.emitAppear(true)
		}
	}
}

//export na_emit_activate
func na_emit_activate(wnd C.uintptr_t, active C.int) {
	win := getWindow(uintptr(wnd))
	if win != nil {
		if active == 0 {
			win.emitActivate(false)
		} else {
			win.emitActivate(true)
		}
	}

}

//export na_emit_text_input
func na_emit_text_input(wnd C.uintptr_t, ch C.int) {
	win := getWindow(uintptr(wnd))
	if win != nil {
		win.emitTextInput(rune(ch))
	}
}

func (win *NaWindow) SetTitle(text string) {
	pstr := C.CString(text)
	defer C.free(unsafe.Pointer(pstr))
	C.na_set_window_title((*C.NaWndData)(unsafe.Pointer(win.NaWndData)), pstr)
}

func (win *NaWindow) SetFocus() {
	C.na_set_focus((*C.NaWndData)(unsafe.Pointer(win.NaWndData)))
}

func GetKeyboardState(keyStates []KeyState) {
	var buf [256]KeyState
	C.na_get_keyboard_state((*C.uint8_t)(unsafe.Pointer(&buf[0])))
	copy(keyStates, buf[:])
}

func (win *NaWindow) ConfineCursor(confine bool) {
	var b C.int
	if confine {
		b = 1
	} else {
		b = 0
	}
	C.na_confine_cursor((*C.NaWndData)(unsafe.Pointer(win.NaWndData)), b)
	win.confineCursor = confine
}

func (win *NaWindow) ShowCursor(show bool) {
	if show {
		x, y := C.float(win.saveMouseX), C.float(win.saveMouseY)
		C.na_show_cursor((*C.NaWndData)(unsafe.Pointer(win.NaWndData)), 1, x, y)
	} else {
		//var x, y C.float
		C.na_show_cursor((*C.NaWndData)(unsafe.Pointer(win.NaWndData)), 0, 0, 0)
		win.saveMouseX, win.saveMouseY = win.mouseX, win.mouseY
	}

	win.hideCursor = !show
}

func (win *NaWindow) ActivateWindow() {
	C.na_activate_window((*C.NaWndData)(unsafe.Pointer(win.NaWndData)))
}

func (win *NaWindow) ReadClipboard(contentType int) (data interface{}, err error) {
	if p := C.na_read_clipboard((*C.NaWndData)(unsafe.Pointer(win.NaWndData))); p != nil {
		data = C.GoString(p)
		C.free(unsafe.Pointer(p))
	} else {
		err = errors.New("unkown error")
	}
	return
}

func (win *NaWindow) WriteClipboard(data interface{}) (err error) {
	pstr := C.CString(data.(string))
	defer C.free(unsafe.Pointer(pstr))

	if ret := C.na_write_clipboard((*C.NaWndData)(unsafe.Pointer(win.NaWndData)), pstr); ret == 0 {
		err = errors.New("unkown error")
		return
	}
	return
}

// func OSVersion() (s string) {
// 	buf := C.na_os_version()
// 	s = C.GoString(buf)
// 	C.free(unsafe.Pointer(buf))
// 	return
// }

// Exit the event loop
func Exit(err error) {
	debug.Println("exit event loop with err = ", err)
	exitErr = err
	C.na_quit_loop()
}

// List of all windows
func List() (ret []*NaWindow) {
	for _, v := range winMap {
		ret = append(ret, v)
	}
	return
}
