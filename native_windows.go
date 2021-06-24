package native

import (
	"reflect"
	"syscall"
	"unsafe"
)

type (
	HWND    = uintptr
	UINT    = uint32 // TODO:
	WPARAM  = uintptr
	LPARAM  = uintptr
	LRESULT = uintptr
	DWORD   = uint32
	BOOL    = uint32

	LPCWSTR *uint16
	//LPWSTR  = LPCWSTR
)

// TODO:  能不能用 //go:notinheap 来优化?

type sWNDCLASSEXW struct {
	cbSize        UINT
	style         UINT
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     uintptr
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  LPCWSTR
	lpszClassName LPCWSTR
	hIconSm       uintptr
}

type sRECT struct {
	left   int32
	top    int32
	right  int32
	bottom int32
}

type sPOINT struct {
	x, y int32
}

type sPAINTSTRUCT struct {
	hdc         uintptr
	fErase      BOOL
	rcPaint     sRECT
	fRestore    BOOL
	fIncUpdate  BOOL
	rgbReserved [32]byte
}

type sMSG struct {
	hwnd    HWND
	message UINT
	wParam  WPARAM
	lParam  LPARAM
	time    DWORD
	pt      sPOINT
}

type sTRACKMOUSEEVENT struct {
	cbSize      DWORD
	dwFlags     DWORD
	hwndTrack   HWND
	dwHoverTime DWORD
}

type sMINMAXINFO struct {
	ptReserved     sPOINT
	ptMaxSize      sPOINT
	ptMaxPosition  sPOINT
	ptMinTrackSize sPOINT
	ptMaxTrackSize sPOINT
}

type sRAWINPUTDEVICELIST struct {
	hDevice uintptr
	dwType  DWORD
}

type sRAWINPUTHEADER struct {
	dwType  DWORD
	dwSize  DWORD
	hDevice uintptr
	wParam  WPARAM
}

type sRAWINPUTDEVICE struct {
	usUsagePage uint16
	usUsage     uint16
	dwFlags     DWORD
	hwndTarget  HWND
}

const (
	ridiPREPARSEDDATA = 0x20000005
	ridiDEVICENAME    = 0x20000007
	ridiDEVICEINFO    = 0x2000000b
)

type sRID_DEVICE_INFO_MOUSE struct {
	dwId                DWORD
	dwNumberOfButtons   DWORD
	dwSampleRate        DWORD
	fHasHorizontalWheel BOOL
}

type sRID_DEVICE_INFO_KEYBOARD struct {
	dwType                 DWORD
	dwSubType              DWORD
	dwKeyboardMode         DWORD
	dwNumberOfFunctionKeys DWORD
	dwNumberOfIndicators   DWORD
	dwNumberOfKeysTotal    DWORD
}
type sRID_DEVICE_INFO_HID struct {
	dwVendorId      DWORD
	dwProductId     DWORD
	dwVersionNumber DWORD
	usUsagePage     uint16
	usUsage         uint16
}

const (
	rimTYPEMOUSE    = 0
	rimTYPEKEYBOARD = 1
	rimTYPEHID      = 2
)

type sRID_DEVICE_INFO struct {
	cbSize DWORD
	dwType DWORD
	data   [20]byte
}

const (
	mouseMOVE_RELATIVE      = 0
	mouseMOVE_ABSOLUTE      = 1
	mouseVIRTUAL_DESKTOP    = 0x02
	mouseATTRIBUTES_CHANGED = 0x04
	mouseMOVE_NOCOALESCE    = 0x08
)

const (
	riMOUSE_LEFT_BUTTON_DOWN   = 0x0001
	riMOUSE_LEFT_BUTTON_UP     = 0x0002
	riMOUSE_MIDDLE_BUTTON_DOWN = 0x0010
	riMOUSE_MIDDLE_BUTTON_UP   = 0x0020
	riMOUSE_RIGHT_BUTTON_DOWN  = 0x0004
	riMOUSE_RIGHT_BUTTON_UP    = 0x0008
	riMOUSE_BUTTON_1_DOWN      = riMOUSE_LEFT_BUTTON_DOWN
	riMOUSE_BUTTON_1_UP        = riMOUSE_LEFT_BUTTON_UP
	riMOUSE_BUTTON_2_DOWN      = riMOUSE_RIGHT_BUTTON_DOWN
	riMOUSE_BUTTON_2_UP        = riMOUSE_RIGHT_BUTTON_UP
	riMOUSE_BUTTON_3_DOWN      = riMOUSE_MIDDLE_BUTTON_DOWN
	riMOUSE_BUTTON_3_UP        = riMOUSE_MIDDLE_BUTTON_UP
	riMOUSE_BUTTON_4_DOWN      = 0x0040
	riMOUSE_BUTTON_4_UP        = 0x0080
	riMOUSE_BUTTON_5_DOWN      = 0x100
	riMOUSE_BUTTON_5_UP        = 0x0200
	riMOUSE_WHEEL              = 0x0400
	riMOUSE_HWHEEL             = 0x0800
)

type sRAWMOUSE struct {
	usFlags            uint16
	ulButtons          uint32
	ulRawButtons       uint32
	lLastX             int32
	lLastY             int32
	ulExtraInformation uint32
}

func (p *sRAWMOUSE) usButtonFlags() uint16 {
	return uint16(p.ulButtons & 0xFFFF)
}

func (p *sRAWMOUSE) usButtonData() uint16 {
	return uint16((p.ulButtons >> 16) & 0xFFFF)
}

const (
	riKEY_MAKE  = 0
	riKEY_BREAK = 1
	riKEY_E0    = 2
	riKEY_E1    = 4
)

type sRAWKEYBOARD struct {
	MakeCode         uint16
	Flags            uint16
	Reserved         uint16
	VKey             uint16
	Message          uint32
	ExtraInformation uint32
}

type sRAWHID struct {
	dwSizeHid DWORD
	dwCount   DWORD
	bRawData  [1]byte
}

type sRAWINPUT struct {
	header sRAWINPUTHEADER
	data   [24]byte
}

func (p *sRAWINPUT) mouse() *sRAWMOUSE {
	return (*sRAWMOUSE)(unsafe.Pointer(&p.data))
}

func (p *sRAWINPUT) keyboard() *sRAWKEYBOARD {
	return (*sRAWKEYBOARD)(unsafe.Pointer(&p.data))
}

func (p *sRAWINPUT) hid() *sRAWINPUT {
	return (*sRAWINPUT)(unsafe.Pointer(&p.data))
}

const (
	csVREDRAW    = 0x0001
	csHREDRAW    = 0x0002
	csDBLCLKS    = 0x0008
	csOWNDC      = 0x0020
	csDROPSHADOW = 0x00020000
	// csCLASSDC         = 0x0040
	// csPARENTDC        = 0x0080
	// csNOCLOSE         = 0x0200
	// csSAVEBITS        = 0x0800
	// csBYTEALIGNCLIENT = 0x1000
	// csBYTEALIGNWINDOW = 0x2000
	// csGLOBALCLASS     = 0x4000
	// csIME             = 0x00010000
)

const (
	gwlpUSERDATA = 0xFFFFFFEB // int(-21)

)

const (
	smCXSCREEN = 0
	smCYSCREEN = 1
)

const (
	wsOVERLAPPED  = 0x00000000
	wsVISIBLE     = 0x10000000
	wsCAPTION     = 0x00C00000
	wsSYSMENU     = 0x00080000
	wsTHICKFRAME  = 0x00040000
	wsMINIMIZEBOX = 0x00020000
	wsMAXIMIZEBOX = 0x00010000
	// wsPOPUP        = 0x80000000
	// wsCHILD        = 0x40000000
	// wsMINIMIZE     = 0x20000000
	// wsDISABLED     = 0x08000000
	// wsCLIPSIBLINGS = 0x04000000
	// wsCLIPCHILDREN = 0x02000000
	// wsMAXIMIZE     = 0x01000000
	// wsBORDER       = 0x00800000
	// wsDLGFRAME     = 0x00400000
	// wsVSCROLL      = 0x00200000
	// wsHSCROLL      = 0x00100000
	// wsGROUP        = 0x00020000
	// wsTABSTOP      = 0x00010000
	//wsOVERLAPPEDWINDOW = wsOVERLAPPED | wsCAPTION | wsSYSMENU | wsTHICKFRAME | wsMINIMIZEBOX | wsMAXIMIZEBOX
)

// const (
// 	wsexAPPWINDOW = 0x00040000
// )

const (
	wmCREATE        = 0x0001
	wmDESTROY       = 0x0002
	wmMOVE          = 0x0003
	wmSIZE          = 0x0005
	wmACTIVATE      = 0x0006
	wmSETFOCUS      = 0x0007
	wmKILLFOCUS     = 0x0008
	wmENABLE        = 0x000A
	wmPAINT         = 0x000F
	wmCLOSE         = 0x0010
	wmQUIT          = 0x0012
	wmSHOWWINDOW    = 0x0018
	wmACTIVATEAPP   = 0x001C
	wmSETCURSOR     = 0x0020
	wmGETMINMAXINFO = 0x0024
	wmINPUT         = 0x00FF
	wmKEYDOWN       = 0x0100
	wmKEYUP         = 0x0101
	wmCHAR          = 0x0102
	wmSYSKEYDOWN    = 0x0104
	wmSYSKEYUP      = 0x0105
	wmUNICHAR       = 0x0109
	wmMOUSEMOVE     = 0x0200
	wmLBUTTONDOWN   = 0x0201
	wmLBUTTONUP     = 0x0202
	wmLBUTTONDBLCLK = 0x0203
	wmRBUTTONDOWN   = 0x0204
	wmRBUTTONUP     = 0x0205
	wmRBUTTONDBLCLK = 0x0206
	wmMBUTTONDOWN   = 0x0207
	wmMBUTTONUP     = 0x0208
	wmMBUTTONDBLCLK = 0x0209
	wmMOUSEWHEEL    = 0x020A
	wmXBUTTONDOWN   = 0x020B
	wmXBUTTONUP     = 0x020C
	wmXBUTTONDBLCLK = 0x020D
	wmMOUSEHWHEEL   = 0x020e
	wmMOUSEHOVER    = 0x02A1
	wmMOUSELEAVE    = 0x02A3
)

const (
	waINACTIVE    = 0
	waACTIVE      = 1
	waCLICKACTIVE = 2
)

const (
	pmREMOVE = 0x0001
)

// wParam for WM_SIZE
const (
	sizeRESTORED = iota
	sizeMINIMIZED
	sizeMAXIMIZED
	sizeMAXSHOW
	sizeMAXHIDE
)

type NaWndData struct {
	WindowID
	btnDown        int32
	mouseX         float32
	mouseY         float32
	restoreRect    sRECT
	restoreStyle   uint32
	restoreExStyle uint32
	trackMouse     uint8
	mouseHover     uint8
	fullScreen     uint8 // for restore
	_padding0      uint8
}

func utf16GoStr(p *uint16) string {
	if p == nil {
		return ""
	}
	var n int
	for *((*uint16)(unsafe.Pointer(uintptr(unsafe.Pointer(p)) + uintptr(n*2)))) != 0 {
		n++
	}
	n++
	var buf []uint16
	hd := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	hd.Data, hd.Len, hd.Cap = uintptr(unsafe.Pointer(p)), int(n), int(n)
	return syscall.UTF16ToString(buf)
}
