//+build !forcecgo

package native

import (
	"errors"
	"sync/atomic"
	"syscall"
	"unsafe"

	"github.com/toy80/debug"
)

import "C"

// 由于cgo版的C源码存在, 不开cgo就报错 "C source files not allowed when not using cgo"
// 所以非cgo版仍然打开cgo, 并用条件编译去掉所有C源码, 实测对编译速度影响极小.

var (
	kernel32dll            = syscall.NewLazyDLL("kernel32.dll")
	procGetCurrentThreadId = kernel32dll.NewProc("GetCurrentThreadId") // DWORD GetCurrentThreadId();
	procGetModuleHandleW   = kernel32dll.NewProc("GetModuleHandleW")   // HMODULE GetModuleHandleW(LPCWSTR lpModuleName);
	procSleep              = kernel32dll.NewProc("Sleep")              // void Sleep(DWORD dwMilliseconds);
	procLocalAlloc         = kernel32dll.NewProc("LocalAlloc")         // HLOCAL LocalAlloc(UINT uFlags, SIZE_T uBytes);
	procLocalFree          = kernel32dll.NewProc("LocalFree")          // HLOCAL LocalFree(HLOCAL hMem);
	procGlobalLock         = kernel32dll.NewProc("GlobalLock")         // LPVOID GlobalLock(HGLOBAL hMem);
	procGlobalUnlock       = kernel32dll.NewProc("GlobalUnlock")       // BOOL GlobalUnlock(HGLOBAL hMem);
	procGlobalAlloc        = kernel32dll.NewProc("GlobalAlloc")        // HGLOBAL GlobalAlloc(UINT   uFlags,SIZE_T dwBytes);
	procGlobalFree         = kernel32dll.NewProc("GlobalFree")         //HGLOBAL GlobalFree(HGLOBAL hMem);

	user32dll                   = syscall.NewLazyDLL("user32.dll")
	procRegisterClassExW        = user32dll.NewProc("RegisterClassExW")        // ATOM RegisterClassExW(const WNDCLASSEXW *Arg1);
	procLoadCursorW             = user32dll.NewProc("LoadCursorW")             // HCURSOR LoadCursorW(HINSTANCE hInstance, LPCWSTR lpCursorName);
	procGetCursorPos            = user32dll.NewProc("GetCursorPos")            // BOOL GetCursorPos(LPPOINT lpPoint);
	procSetCursorPos            = user32dll.NewProc("SetCursorPos")            // BOOL SetCursorPos(int X,int Y)
	procScreenToClient          = user32dll.NewProc("ScreenToClient")          // BOOL ScreenToClient(HWND hWnd, LPPOINT lpPoint);
	procClientToScreen          = user32dll.NewProc("ClientToScreen")          // BOOL ClientToScreen(HWND hWnd, LPPOINT lpPoint);
	procGetClientRect           = user32dll.NewProc("GetClientRect")           // BOOL GetClientRect(	HWND hWnd, LPRECT lpRect);
	procGetSystemMetrics        = user32dll.NewProc("GetSystemMetrics")        // int GetSystemMetrics(int nIndex)
	procAdjustWindowRectEx      = user32dll.NewProc("AdjustWindowRectEx")      // BOOL AdjustWindowRectEx(LPRECT lpRect, DWORD dwStyle, BOOL bMenu, DWORD dwExStyle)
	procCreateWindowExW         = user32dll.NewProc("CreateWindowExW")         // HWND CreateWindowExW(DWORD dwExStyle, LPCWSTR lpClassName, LPCWSTR lpWindowName, DWORD dwStyle, int X, int Y, int nWidth, int nHeight, HWND hWndParent, HMENU hMenu, HINSTANCE hInstance, LPVOID lpParam);
	procDefWindowProcW          = user32dll.NewProc("DefWindowProcW")          // LRESULT DefWindowProcW(HWND hWnd, UINT Msg, WPARAM wParam, LPARAM lParam)
	procDestroyWindow           = user32dll.NewProc("DestroyWindow")           // BOOL DestroyWindow(HWND hWnd);
	procShowWindow              = user32dll.NewProc("ShowWindow")              // BOOL ShowWindow(HWND hWnd, int nCmdShow);
	procFlashWindow             = user32dll.NewProc("FlashWindow")             // BOOL FlashWindow (HWND hWnd, BOOL bInvert);
	procCloseWindow             = user32dll.NewProc("CloseWindow")             // BOOL CloseWindow (HWND hWnd);
	procMoveWindow              = user32dll.NewProc("MoveWindow")              // BOOL MoveWindow (HWND hWnd, int X, int Y, int nWidth, int nHeight, BOOL bRepaint);
	procGetDC                   = user32dll.NewProc("GetDC")                   // HDC GetDC(HWND hWnd);
	procReleaseDC               = user32dll.NewProc("ReleaseDC")               // int ReleaseDC(HWND hWnd,HDC hDC);
	procBeginPaint              = user32dll.NewProc("BeginPaint")              // HDC BeginPaint(HWND hWnd,LPPAINTSTRUCT lpPaint);
	procEndPaint                = user32dll.NewProc("EndPaint")                // BOOL EndPaint(HWND hWnd,CONST PAINTSTRUCT *lpPaint);
	procInvalidateRect          = user32dll.NewProc("InvalidateRect")          // BOOL InvalidateRect(HWND hWnd,CONST RECT *lpRect,BOOL bErase);
	procPostQuitMessage         = user32dll.NewProc("PostQuitMessage")         // void PostQuitMessage(int nExitCode)
	procPeekMessageW            = user32dll.NewProc("PeekMessageW")            // BOOL PeekMessageW(LPMSG lpMsg, HWND  hWnd, UINT  wMsgFilterMin, UINT  wMsgFilterMax, UINT  wRemoveMsg);
	procTranslateMessage        = user32dll.NewProc("TranslateMessage")        // BOOL TranslateMessage(const MSG *lpMsg);
	procDispatchMessageW        = user32dll.NewProc("DispatchMessageW")        // LRESULT DispatchMessageW(const MSG *lpMsg);
	procSetCapture              = user32dll.NewProc("SetCapture")              // HWND SetCapture(HWND hWnd);
	procReleaseCapture          = user32dll.NewProc("ReleaseCapture")          // BOOL ReleaseCapture();
	procTrackMouseEvent         = user32dll.NewProc("TrackMouseEvent")         // BOOL TrackMouseEvent(LPTRACKMOUSEEVENT lpEventTrack);
	procSetWindowTextW          = user32dll.NewProc("SetWindowTextW")          // BOOL SetWindowTextW( HWND hWnd, LPCWSTR lpString);
	procSetWindowLongPtrW       = user32dll.NewProc("SetWindowLongPtrW")       // LONG_PTR SetWindowLongPtrW(HWND hWnd, int nIndex, LONG_PTR dwNewLong);
	procGetWindowLongPtrW       = user32dll.NewProc("GetWindowLongPtrW")       // LONG_PTR GetWindowLongPtrW(HWND hWnd, int nIndex);
	procGetWindowDC             = user32dll.NewProc("GetWindowDC")             // HDC GetWindowDC(HWND hWnd);
	procGetKeyboardState        = user32dll.NewProc("GetKeyboardState")        // BOOL GetKeyboardState(PBYTE lpKeyState);
	procSetFocus                = user32dll.NewProc("SetFocus")                // HWND SetFocus(HWND hWnd);
	procClipCursor              = user32dll.NewProc("ClipCursor")              // BOOL ClipCursor(const RECT *lpRect);
	procGetClipCursor           = user32dll.NewProc("GetClipCursor")           // BOOL GetClipCursor(LPRECT lpRect)
	procSetCursor               = user32dll.NewProc("SetCursor")               // HCURSOR SetCursor(HCURSOR hCursor);
	procShowCursor              = user32dll.NewProc("ShowCursor")              // int ShowCursor(BOOL bShow);
	procSetActiveWindow         = user32dll.NewProc("SetActiveWindow")         // HWND SetActiveWindow(HWND hWnd)
	procGetRawInputDeviceList   = user32dll.NewProc("GetRawInputDeviceList")   // UINT GetRawInputDeviceList(PRAWINPUTDEVICELIST pRawInputDeviceList,PUINT puiNumDevices,UINT cbSize);
	procGetRawInputDeviceInfoA  = user32dll.NewProc("GetRawInputDeviceInfoA")  // UINT GetRawInputDeviceInfoA(HANDLE hDevice,UINT uiCommand,LPVOID pData,PUINT pcbSize);
	procGetRawInputData         = user32dll.NewProc("GetRawInputData")         // UINT GetRawInputData(HRAWINPUT hRawInput,UINT uiCommand,LPVOID pData,PUINT pcbSize,UINT cbSizeHeader);
	procGetRawInputBuffer       = user32dll.NewProc("GetRawInputBuffer")       // UINT GetRawInputBuffer(PRAWINPUT pData, PUINT pcbSize, UINT cbSizeHeader);
	procRegisterRawInputDevices = user32dll.NewProc("RegisterRawInputDevices") // BOOL RegisterRawInputDevices(PCRAWINPUTDEVICE pRawInputDevices, UINT uiNumDevices, UINT cbSize);
	procOpenClipboard           = user32dll.NewProc("OpenClipboard")           // BOOL OpenClipboard(HWND);
	procCloseClipboard          = user32dll.NewProc("CloseClipboard")          // BOOL CloseClipboard();
	procGetClipboardData        = user32dll.NewProc("GetClipboardData")        // HANDLE GetClipboardData(UINT uFormat);
	procSetClipboardData        = user32dll.NewProc("SetClipboardData")        // HANDLE SetClipboardData(UINT   uFormat,HANDLE hMem);
	procEmptyClipboard          = user32dll.NewProc("EmptyClipboard")          // BOOL EmptyClipboard();

	// dwmapidll                        = syscall.NewLazyDLL("dwmapi.dll")
	// procDwmExtendFrameIntoClientArea = dwmapidll.NewProc("DwmExtendFrameIntoClientArea") // DWMAPI DwmExtendFrameIntoClientArea(HWND hWnd,const MARGINS *pMarInset);
)

var (
	// class name for register window class
	wndClassName = [...]uint16{'T', 'o', 'y', '8', '0', '_', 'W', 'i', 'n', 'L', 0}

	idMainThread uint32
	hInstance    uintptr
	hCursorArrow uintptr

	wndproc = syscall.NewCallback(wndprocFunc)
)

// type W32BackendCreateInfo struct {
// 	Hinstance uintptr
// 	Hwnd      uintptr
// 	Hdc       uintptr
// }

// func (win *Window) Native() interface{} {
// 	return &W32BackendCreateInfo{
// 		Hinstance: hInstance,
// 		Hwnd:      win.ID,
// 		Hdc:       win.Hdc,
// 	}
// }

func initNative() {
	ret, _, _ := procGetCurrentThreadId.Call()
	idMainThread = uint32(ret)

	hInstance, _, _ = procGetModuleHandleW.Call(0)

	// register the window class

	hCursorArrow, _, _ = procLoadCursorW.Call(0, 32512) // MAKEINTRESOURCE(32512)
	const style = csOWNDC | csVREDRAW | csHREDRAW | csDBLCLKS | csDROPSHADOW
	lpszClassName := LPCWSTR(&wndClassName[0])
	wc := sWNDCLASSEXW{
		cbSize:        80,            //
		style:         style,         //
		lpfnWndProc:   wndproc,       //
		cbClsExtra:    0,             //
		cbWndExtra:    0,             // note: USERDATA is different than extra bytes
		hInstance:     hInstance,     //
		hIcon:         0,             // TODO: allow set window icon
		hCursor:       hCursorArrow,  // 有两种方式隐藏光标, 简单起见用 ShowCursor 方式. 如果用 SetCursor 方式就要指定为NULL
		hbrBackground: 0,             // don't draw anything with GDI
		lpszMenuName:  nil,           //
		lpszClassName: lpszClassName, //
		hIconSm:       0,             // TODO:
	}
	if unsafe.Sizeof(wc) != 80 {
		panic("wrong size of WNDCLASSEXW")
	}

	ret, _, _ = procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	if ret == 0 {
		panic("RegisterClassExW() failure")
	}
}

// IsMainThread reports whether current thread is the main thread
func IsMainThread() bool {
	idCurrentThread, _, _ := procGetCurrentThreadId.Call()
	return uint32(idCurrentThread) == atomic.LoadUint32(&idMainThread)
}

// func getWndData(hWnd uintptr) *sWndData {
// 	ret, _, _ := procGetWindowLongPtrW.Call(hWnd, gwlpUSERDATA)
// 	// TODO: go vet: possible misuse of unsafe.Pointer
// 	return (*sWndData)(unsafe.Pointer(ret))
// }

func GetCursorPos(win uintptr) (x, y float32, err error) {
	if win == 0 {
		return 0, 0, ErrGeneric
	}
	var pt sPOINT
	ret, _, err := procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	if ret == 0 {
		return
	}
	if win != 0 {
		ret, _, err = procScreenToClient.Call(win, uintptr(unsafe.Pointer(&pt)))
		if ret == 0 {
			return
		}
	}
	return float32(pt.x), float32(pt.y), nil
}

func GetScreenSize(id int) (width, height float32, err error) {
	// TODO: multi-screen ?
	w, _, _ := procGetSystemMetrics.Call(smCXSCREEN)
	h, _, _ := procGetSystemMetrics.Call(smCYSCREEN)
	return float32(w), float32(h), nil
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
	// Windows 10 默认有一个边框, 可能要用 DWM API 来去除
	style := uintptr(wsOVERLAPPED | wsCAPTION | wsSYSMENU)
	if hints&WinHintResizable != 0 {
		style |= wsTHICKFRAME | wsMAXIMIZEBOX | wsMINIMIZEBOX
	}
	if hints&WinHintMaximizeBox != 0 {
		style |= wsMAXIMIZEBOX
	}
	if hints&WinHintMinimizeBox != 0 {
		style |= wsMINIMIZEBOX
	}
	rc := sRECT{left: x, top: y, right: x + w, bottom: y + h}
	_, _, _ = procAdjustWindowRectEx.Call(uintptr(unsafe.Pointer(&rc)), style, 0, 0)
	debug.Printf("Win32API.CreateWindow(%+v)", rc)
	title := [...]uint16{'T', 'o', 'y', '8', '0', 0}
	win.NaWndData = new(NaWndData)
	win.ID, _, _ = procCreateWindowExW.Call(
		0, // 不需要 WS_EX_APPWINDOW, 因为我们是独立的窗口, 已经在任务栏里了.
		uintptr(unsafe.Pointer(&wndClassName[0])),
		uintptr(unsafe.Pointer(&title[0])),
		style,
		uintptr(rc.left), uintptr(rc.top), uintptr(rc.right-rc.left), uintptr(rc.bottom-rc.top),
		0, 0, hInstance, 0)
	if win.ID == 0 {
		err = errors.New("failed to create native window")
		win.NaWndData = nil
		return
	}
	win.width, win.height = float32(w), float32(h)

	// 注册鼠标的 WM_INPUT 消息
	var rid = [1]sRAWINPUTDEVICE{{
		usUsagePage: 0x01,       // HID_USAGE_PAGE_GENERIC
		usUsage:     0x02,       // HID_USAGE_GENERIC_MOUSE
		dwFlags:     0x00000100, // RIDEV_INPUTSINK
		hwndTarget:  win.ID,
	}}
	ret, _, _ := procRegisterRawInputDevices.Call(uintptr(unsafe.Pointer(&rid)), 1, unsafe.Sizeof(rid[0]))
	if ret == 0 {
		err = errors.New("failed to register raw input devices")
		win.NaWndData = nil
		return
	}

	// if err := procDwmExtendFrameIntoClientArea.Find(); err == nil {
	// 	margins := [4]int32{20, 20, 20, 20} // L,R,T,B
	// 	_, _, _ = procDwmExtendFrameIntoClientArea.Call(win.ID, uintptr(unsafe.Pointer(&margins[0])))
	// }
	//win.DC, _, _ = procGetDC.Call(win.Wnd)
	// TODO: na_track_mouse((uintptr_t)hWnd, 1);
	err = nil
	return
}

func (win *NaWindow) destroyNativeWindow() error {
	ret, _, err := procDestroyWindow.Call(win.ID)
	if ret == 0 {
		return err
	}
	return nil
}

// MemAlloc allocate zeroed unmamaged memory block
// func MemAlloc(sz uintptr) unsafe.Pointer {
//	if sz == 0 {
//		sz = 1 // MemAlloc(0) should return a non nil pointer
//	}
// 	ret, _, _ := procLocalAlloc.Call(0x0040, sz) // 0x0040 = LMEM_FIXED | LMEM_ZEROINIT.
// 	return unsafe.Pointer(ret)
// }

// // MemFree release memory block that allocated with MemAlloc()
// func MemFree(p unsafe.Pointer) {
// 	procLocalFree.Call(uintptr(p))
// }

func defWindowProc(hWnd uintptr, message UINT, wParam WPARAM, lParam LPARAM) LRESULT {
	ret, _, _ := procDefWindowProcW.Call(hWnd, uintptr(message), uintptr(wParam), uintptr(lParam))
	return LRESULT(ret)
}

// func getWindow(hWnd uintptr) *Window {
// 	return winMap[hWnd]
// }

// the window proc
func wndprocFunc(hWnd uintptr, message UINT, wParam WPARAM, lParam LPARAM) LRESULT {
	if win := getWindow(hWnd); win != nil {
		return win.wndproc(message, wParam, lParam)
	}
	// fmt.Println("message=", message)
	// WM_CREATE goes here
	return defWindowProc(hWnd, message, wParam, lParam)
}

func (win *NaWindow) releaseOtherNativeResource() {
	// syscall版, NaWndData分配于go内存, 受GC管理
}

func getWidthParam(lp LPARAM) float32 {
	return float32(uint16(lp & 0xFFFF))
}

func getHeightParam(lp LPARAM) float32 {
	return float32(uint16((lp >> 16) & 0xFFFF))
}

func getXParam(lp LPARAM) float32 {
	return float32((int16)(uint16(lp & 0xFFFF)))
}

func getYParam(lp LPARAM) float32 {
	return float32((int16)(uint16((lp >> 16) & 0xFFFF)))
}

func getWheelDeltaParam(wp WPARAM) float32 {
	return float32((int16)(uint16((wp >> 16) & 0xFFFF)))
}

func (win *NaWindow) wndproc(message UINT, wParam WPARAM, lParam LPARAM) LRESULT {
	hWnd := win.ID
	switch message {
	case wmDESTROY:
		win.onNativeDestroy()
		return 0
	case wmPAINT:
		{
			var ps sPAINTSTRUCT
			_, _, _ = procBeginPaint.Call(hWnd, uintptr(unsafe.Pointer(&ps)))
			rc := ps.rcPaint
			win.emitExpose(float32(rc.left), float32(rc.top), float32(rc.right-rc.left), float32(rc.bottom-rc.top))
			_, _, _ = procEndPaint.Call(hWnd, uintptr(unsafe.Pointer(&ps)))
		}
		return 0
	case wmSIZE:
		switch wParam {
		case sizeMINIMIZED:
			// the window become invisible
			win.emitAppear(false)
		case sizeMAXHIDE:
			win.emitAppear(true)
		case sizeMAXIMIZED:
			// the window become visible
			fallthrough
		case sizeMAXSHOW:
			fallthrough
		default: // SIZE_RESTORED
			// normal resize
			win.emitAppear(true)
			win.emitResize(getWidthParam(lParam), getHeightParam(lParam))
		}
		return 0
	case wmCLOSE:
		win.onNativeRequestDestroy()
		return 0
	case wmLBUTTONDOWN:
		if win.trackMouse != 0 && win.btnDown == 0 {
			procSetCapture.Call(hWnd)
		}

		win.btnDown |= MouseBtnLeft
		// na_emit_mouse_press((uintptr_t)hWnd, NA_MOUSE_BTN_LEFT, getXParam(lParam), getYParam(lParam));
		win.emitMousePress(MouseBtnLeft, getXParam(lParam), getYParam(lParam))
		return 0
	case wmLBUTTONUP:
		win.btnDown &= ^MouseBtnLeft
		win.emitMouseRelease(MouseBtnLeft, getXParam(lParam), getYParam(lParam))
		if win.trackMouse != 0 && win.btnDown == 0 {
			procReleaseCapture.Call()
		}
		return 0
	case wmRBUTTONDOWN:
		if win.trackMouse != 0 && win.btnDown == 0 {
			procSetCapture.Call(hWnd)
		}
		win.btnDown |= MouseBtnRight
		// na_emit_mouse_press((uintptr_t)hWnd, NA_MOUSE_BTN_LEFT, getXParam(lParam), getYParam(lParam));
		win.emitMousePress(MouseBtnRight, getXParam(lParam), getYParam(lParam))
		return 0
	case wmRBUTTONUP:
		win.btnDown &= ^MouseBtnRight
		win.emitMouseRelease(MouseBtnRight, getXParam(lParam), getYParam(lParam))
		if win.trackMouse != 0 && win.btnDown == 0 {
			procReleaseCapture.Call()
		}
		return 0
	case wmMOUSEMOVE:
		if win.trackMouse != 0 && win.mouseHover == 0 {
			win.mouseHover = 1
			var tme = sTRACKMOUSEEVENT{
				0,
				0x00000002, // TME_LEAVE
				hWnd,
				0xFFFFFFFF, //HOVER_DEFAULT
			}
			tme.cbSize = uint32(unsafe.Sizeof(tme))
			procTrackMouseEvent.Call(uintptr(unsafe.Pointer(&tme))) // one shot
			win.emitMouseEnter(getXParam(lParam), getYParam(lParam))
		}
		win.emitMouseMove(getXParam(lParam), getYParam(lParam))
		return 0
	case wmMOUSEWHEEL:
		win.emitMouseWheel(1, getWheelDeltaParam(wParam), getXParam(lParam), getYParam(lParam))
		return 0
	case wmMOUSEHWHEEL:
		win.emitMouseWheel(0, getWheelDeltaParam(wParam), getXParam(lParam), getYParam(lParam))
		return 0
	case wmMOUSELEAVE:
		if win.trackMouse != 0 && win.mouseHover != 0 {
			win.mouseHover = 0
			x, y, _ := GetCursorPos(hWnd)
			win.emitMouseLeave(x, y)
		}
		return 0
	case wmGETMINMAXINFO:
		pMMI := (*sMINMAXINFO)(unsafe.Pointer(lParam))
		pMMI.ptMinTrackSize.x = 100
		pMMI.ptMinTrackSize.y = 100
		return 0
	case wmSYSKEYDOWN:
		if wParam == 0x73 && lParam&0x20000000 != 0 { // ALT+F4
			break
		}
		return 0
	case wmKEYDOWN, wmKEYUP, wmSYSKEYUP:
		// if message == wmKEYUP && wParam == 0x1B {
		// }
		// debug.Printf("wmKEYDOWN, wmKEYUP, wmSYSKEYDOWN, wmSYSKEYUP\n")
		return 0 // 我们用自己的方式处理键盘事件, 以消除F10等默认按键的影响. Win键之类由于系统的限制不能拦截.
	case wmACTIVATE:
		win.emitActivate(wParam != 0)
		break
		// case wmSETCURSOR: 目前用 ShowCursor方式
		// 	if win.cursor || (lParam&0xFFFF) != 1 { // 不在client区域时要显示光标
		// 		procSetCursor.Call(hCursorArrow)
		// 	} else {
		// 		procSetCursor.Call(0)
		// 	}
		// 	return 1
	case wmINPUT: // "低级"外设输入
		const sizeRAWINPUT = unsafe.Sizeof(*(*sRAWINPUT)(nil))
		const sizeRAWINPUTHEADER = unsafe.Sizeof(*(*sRAWINPUTHEADER)(nil))
		var raw sRAWINPUT
		dwSize := uint32(sizeRAWINPUT)
		procGetRawInputData.Call(uintptr(lParam), 0x10000003, /*RID_INPUT*/
			uintptr(unsafe.Pointer(&raw)), uintptr(unsafe.Pointer(&dwSize)), sizeRAWINPUTHEADER)
		if raw.header.dwType == rimTYPEMOUSE {
			mouse := raw.mouse()
			win.emitMouseMoveRelative(float32(mouse.lLastX), float32(mouse.lLastY))
		}
		break
	case wmCHAR: // 不用 WM_UNICHAR, 因为收不到消息. 据说那是给IME用的
		win.emitTextInput(rune(wParam))
		return 0
	}

	return defWindowProc(win.ID, message, wParam, lParam)
}

// func (win *Window) Size() (float32, float32) {
// 	return win.width, win.height
// }

func (win *NaWindow) SetVisible(visible bool) {
	debug.Printf("Window.SetVisible(%v)", visible)
	var b uintptr
	if visible {
		b = 1
	}
	_, _, _ = procShowWindow.Call(win.ID, b)
}

func (win *NaWindow) ExposeRect(x, y, width, height float32) {
	rc := sRECT{int32(x), int32(y), int32(x + width), int32(y + height)}
	procInvalidateRect.Call(win.ID, uintptr(unsafe.Pointer(&rc)), 0)
}

// YieldThread cause the current thread to sleep a minimun span.
func YieldThread() {
	// on Windows, Sleep(0) still full fill the CPU, so Sleep(1) required.
	_, _, _ = procSleep.Call(1)
}

func Exit(code int) {
	_, _, _ = procPostQuitMessage.Call(uintptr(code))
}

var msg sMSG

func DispatchEvent() int {
	if ok, _, _ := procPeekMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0, pmREMOVE); ok != 0 {
		if msg.message != wmQUIT {
			_, _, _ = procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
			_, _, _ = procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
			return EventDispatched
		}
		return EventRequestToQuit
	}
	return EventQueueEmpty
}

// func  onWndDestroy(win *Window) {
// 	numnumWnds--
// }

func (win *NaWindow) SetTitle(text string) {
	if ptr, err := syscall.UTF16PtrFromString(text); err == nil {
		_, _, _ = procSetWindowTextW.Call(win.ID, uintptr(unsafe.Pointer(ptr)))
	}
}

func GetKeyboardState(keyStates []KeyState) {
	var buf [256]KeyState
	_, _, _ = procGetKeyboardState.Call(uintptr(unsafe.Pointer(&buf[0])))
	copy(keyStates, buf[:])
}

func (win *NaWindow) SetFocus() {
	_, _, _ = procSetFocus.Call(win.ID)
}

// 隐藏光标, 并限制在窗口内部
func (win *NaWindow) ConfineCursor(confine bool) {
	win.confineCursor = confine
	if confine {
		var rect sRECT
		ret, _, _ := procGetClientRect.Call(win.ID, uintptr(unsafe.Pointer(&rect)))
		if ret == 0 {
			return
		}
		ret, _, _ = procClientToScreen.Call(win.ID, uintptr(unsafe.Pointer(&rect)))
		if ret == 0 {
			return
		}
		ret, _, _ = procClientToScreen.Call(win.ID, uintptr(unsafe.Pointer(&rect))+8)
		if ret == 0 {
			return
		}
		_, _, _ = procClipCursor.Call(uintptr(unsafe.Pointer(&rect)))
	} else {
		_, _, _ = procClipCursor.Call(0)
	}
}

func (win *NaWindow) ShowCursor(show bool) {
	if show {
		var n int32 = -1
		for n < 0 {
			n1, _, _ := procShowCursor.Call(1)
			n = int32(uint32(n1))
		}
		if win.hideCursor {
			// 恢复到隐藏前的位置
			pt := sPOINT{
				x: int32(win.saveMouseX),
				y: int32(win.saveMouseY),
			}
			procClientToScreen.Call(win.ID, uintptr(unsafe.Pointer(&pt)))
			procSetCursorPos.Call(uintptr(pt.x), uintptr(pt.y))
		}
	} else {
		if !win.hideCursor {
			win.saveMouseX, win.saveMouseY = win.mouseX, win.mouseY
		}
		var n int32 = 0
		for n >= 0 {
			n1, _, _ := procShowCursor.Call(0)
			n = int32(uint32(n1))
		}
	}
	win.hideCursor = !show
}

func (win *NaWindow) ActivateWindow() {
	procSetActiveWindow.Call(win.ID)
}

// contentType 暂时不起作用, 只支持 UNICODE 文字
func (win *NaWindow) ReadClipboard(contentType int) (data interface{}, err error) {
	ok, _, _ := procOpenClipboard.Call(win.ID)
	if ok == 0 {
		err = errors.New("OpenClipboard() failed")
		return
	}
	defer procCloseClipboard.Call()
	hMem, _, _ := procGetClipboardData.Call(13) // CF_UNICODETEXT=13 CF_TEXT=1
	if hMem == 0 {
		err = errors.New("GetClipboardData() failed")
		return
	}
	lpszClipboard, _, _ := procGlobalLock.Call(hMem)
	if lpszClipboard == 0 {
		err = errors.New("GlobalLock() failed")
		return
	}
	defer procGlobalUnlock.Call(hMem)

	data = utf16GoStr((*uint16)(unsafe.Pointer(lpszClipboard)))
	err = nil
	return
}

func (win *NaWindow) WriteClipboard(data interface{}) (err error) {
	text := data.(string) // 目前仅支持文字

	ok, _, _ := procOpenClipboard.Call(win.ID)
	if ok == 0 {
		err = errors.New("OpenClipboard() failed")
		return
	}
	defer procCloseClipboard.Call()

	_, _, _ = procEmptyClipboard.Call(win.ID)
	// if ok == 0 {
	// 	if err == nil {
	// 		err = errors.New("EmptyClipboard() failed")
	// 	}
	// 	return
	// }

	u16, _ := syscall.UTF16FromString(text)
	if len(u16) == 0 {
		err = errors.New("UTF16FromString() failed")
		return
	}

	hMem, _, _ := procGlobalAlloc.Call(0x0002 /*GMEM_MOVEABLE*/, uintptr(2*len(u16)))
	if hMem == 0 {
		err = errors.New("GlobalAlloc() failed")
		return
	}
	lpszClipboard, _, _ := procGlobalLock.Call(hMem)
	if lpszClipboard == 0 {
		procGlobalFree.Call(hMem)
		err = errors.New("GlobalLock() failed")
		return
	}
	defer procGlobalUnlock.Call(hMem)

	for i, c := range u16 {
		*((*uint16)(unsafe.Pointer(lpszClipboard + uintptr(i*2)))) = c
	}

	h, _, _ := procSetClipboardData.Call(13, hMem)
	if h == 0 {
		procGlobalFree.Call(hMem)
		err = errors.New("SetClipboardData() failed")
		return
	}
	return
}
