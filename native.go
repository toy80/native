package native

import (
	"errors"
	"runtime"

	"github.com/toy80/debug"
)

func init() {
	// many system call have poor multi-thread support, we stick the "main"
	// goroutine to the thread that invoke init functions
	runtime.LockOSThread()

	initNative()
}

const (
	WinHintResizable = 1 << iota
	WinHintFullScreen
	WinHintNoVideo
	WinHintMinimizeBox
	WinHintMaximizeBox
)

const (
	MouseBtnLeft = 1 << iota
	MouseBtnRight
	MouseBtnMiddle
)

const (
	EventDispatched    = 0
	EventRequestToQuit = 1
	EventQueueEmpty    = 2
)

var (
	ErrGeneric = errors.New("generic error")

	// map native window to Go Window
	winMap = make(map[uintptr]*NaWindow)

	KeepRunningWithoutWindow bool

	numWnds  int
	firstWnd bool
	quit     bool
	started  bool
	exitErr  error
	exited   bool
	//
	//gOnExit  func(error)
)

// IWindow is interface to window
type IWindow interface {
	OnWindowCreate()
	OnWindowActivate(active bool)
	OnWindowAppear(visible bool)
	OnWindowResize(width, height float32)
	OnMouseMove(x, y float32)
	OnMouseMoveRelative(dx, dy float32)
	OnMousePress(btn int, x, y float32)
	OnMouseRelease(btn int, x, y float32)
	OnMouseWheel(axis int, delta, x, y float32) // up is positive
	OnMouseEnter(x, y float32)
	OnMouseLeave(x, y float32)
	OnWindowExpose(x, y, width, height float32)
	OnTextInput(ch rune)
	WindowSize() (width, height float32)
	//ShowCursor(b bool)

	// 销毁窗口. 内部先调用Deinit,再销毁Native资源.
	Destroy()

	// 析构. 对于窗口来说, 要销毁请调用Destroy而不是Deinit
	Deinit()

	SetFocus()

	ConfineCursor(confine bool) // 把鼠标指针限制在窗口客户区
	IsCursorConfined() bool     // 鼠标指针是否限制在窗口客户区

	ShowCursor(show bool)  // 显示鼠标指针
	IsCursorVisible() bool // 鼠标指针是否可见

	ActivateWindow()
	IsWindowActive() bool
}

func (win *NaWindow) emitCreate() {
	win.self.OnWindowCreate()
}

// func (win *NaWindow) emitRequestToClose() bool {
// 	return win.self.OnWindowRequestToClose()
// }

// func (win *NaWindow) emitWindowClose() {
// 	win.self.OnWindowClose()
// }

func (win *NaWindow) emitActivate(active bool) {
	if win.active == active {
		return
	}
	win.active = active
	win.self.OnWindowActivate(active)
}

func (win *NaWindow) emitAppear(visible bool) {
	if win.visible == visible {
		return
	}
	win.visible = visible
	win.self.OnWindowAppear(visible)
}

func (win *NaWindow) emitResize(width, height float32) {
	if win.width == width && win.height == height {
		return
	}
	win.width, win.height = width, height
	win.self.OnWindowResize(width, height)
}

func (win *NaWindow) emitMouseMove(x, y float32) {
	win.mouseX, win.mouseY = x, y
	win.self.OnMouseMove(x, y)
}

func (win *NaWindow) emitMouseMoveRelative(dx, dy float32) {
	win.self.OnMouseMoveRelative(dx, dy)
}

func (win *NaWindow) emitMousePress(btn int, x, y float32) {
	win.self.OnMousePress(int(btn), x, y)
}

func (win *NaWindow) emitMouseRelease(btn int, x, y float32) {
	win.self.OnMouseRelease(int(btn), x, y)
}

func (win *NaWindow) emitMouseWheel(axis int, delta, x, y float32) {
	win.self.OnMouseWheel(axis, delta, x, y)
}

func (win *NaWindow) emitMouseEnter(x, y float32) {
	win.self.OnMouseEnter(x, y)
}

func (win *NaWindow) emitMouseLeave(x, y float32) {
	win.self.OnMouseLeave(x, y)
}

func (win *NaWindow) emitExpose(x, y, width, height float32) {
	win.self.OnWindowExpose(x, y, width, height)
}

func (win *NaWindow) emitTextInput(ch rune) {
	win.self.OnTextInput(ch)
}

// func (win *NaWindow) emitTimerTick() {
// 	win.self.OnTimerTick()
// }

func (win *NaWindow) OnWindowCreate() {}
func (win *NaWindow) OnWindowActivate(active bool) {
	debug.Trace(1)
	win.SetFocus()
}
func (win *NaWindow) OnWindowAppear(visible bool) {
	debug.Trace(1)
	win.SetFocus()
}
func (win *NaWindow) OnWindowResize(width, height float32)       {}
func (win *NaWindow) OnMouseMove(x, y float32)                   {}
func (win *NaWindow) OnMouseMoveRelative(dx, dy float32)         {}
func (win *NaWindow) OnMousePress(btn int, x, y float32)         {}
func (win *NaWindow) OnMouseRelease(btn int, x, y float32)       {}
func (win *NaWindow) OnMouseWheel(axis int, delta, x, y float32) {}
func (win *NaWindow) OnMouseEnter(x, y float32)                  {}
func (win *NaWindow) OnMouseLeave(x, y float32)                  {}
func (win *NaWindow) OnWindowExpose(x, y, width, height float32) {}
func (win *NaWindow) OnTextInput(ch rune)                        {}
func (win *NaWindow) OnTimerTick()                               {}

// WindowID 是平台相关的“窗口ID”
type WindowID struct {
	// 平台相关的“窗口ID”
	ID uintptr
	// 平台相关的额外信息，例如 X Window 的 Connection
	Ctx uintptr
}

// NaWindow represent a native window, it wraps a HWND on Windows, XCB visual or XLib NaWindow on Linux, NSView on MacOS.
type NaWindow struct {
	*NaWndData // 平台相关的数据, cgo分支分配于C内存

	self          IWindow
	width         float32
	height        float32
	saveMouseX    float32
	saveMouseY    float32
	visible       bool // on screen and neither minimized nor covered by another maximized window
	active        bool // at foreground and receive keyboard input
	hideCursor    bool // 系统光标是否隐藏
	confineCursor bool
	hints         uint64

	destroying bool // 已经开始销毁. 一旦开始直到最后都是true
	deinited   bool // go 对象的Deinit已经被调用
	destroyed  bool // 系统的"销毁窗口"已经被调用
}

func NumWindows() int {
	return int(numWnds)
}

func Running() bool {
	if quit {
		return false
	}
	if KeepRunningWithoutWindow {
		return true
	}
	if firstWnd {
		return true
	}
	return numWnds != 0
}

type WindowOptions struct {
	Hints         uint64
	Width, Height float32
}

func (win *NaWindow) Init(o *WindowOptions) (err error) {
	if win.self == nil {
		panic(`win.self == nil`)
	}

	if o == nil {
		debug.Println("window options is nil")
		o = &WindowOptions{}
	}

	err = win.sysCreateWindow(o.Hints, o.Width, o.Height)
	if err != nil {
		return
	}
	win.hints = o.Hints
	// inst := o.Instance
	// if inst == nil {
	// 	inst = defaultEngine
	// 	panic("nil engine")
	// }

	// if !inst.Initialized() {
	// 	if err = inst.InitDefaults(); err != nil {
	// 		return
	// 	}
	// }

	// if win.Surface, err = inst.CreateSurface(win.ID, win.hDC); err != nil {
	// 	destroyNativeWindow(win.ID)
	// 	win.ID = 0
	// 	return
	// }

	winMap[win.ID] = win
	numWnds++
	firstWnd = false

	win.emitCreate()

	return
}

func (win *NaWindow) Deinit() {

	// 销毁窗口的过程和接口, 要达成以下几个目标:
	//
	// 1. 用户可以通过系统UI来关闭窗口
	// 2. 程序内部可以通过代码来关闭窗口
	// 3. 程序可以在关闭窗口前插入确认步骤
	// 4. 程序里窗口对象正确析构, 窗口本身正确销毁
	// 5. 接口应该易用且无歧义, 且使用习惯和程序的其他接口一致
	//
	// 问: 正确析构和销毁的含义?
	// 答: Deinit()和"销毁窗口"两个步骤都要执行, 而且确保Close()在"销毁窗口"前执行
	//
	// 问: 窗口销毁事件(即WM_DESTROY)是否暴露给程序?
	// 答: 从程序开发的角度, 这个事件和Deinit()几乎没有区别, 所以应该把它和Deinit()无缝整合.
	//

	// 我们要求用 Destroy() 来触发析构, 这样上层的程序可以用IsDestroying()来判断是否在析构的过程中.
	debug.Assert(win.destroying, "don't call Deinit() directly for a window, call Destroy() instead.")

	// 标记Deinit()已经被调用, 避免再次进入
	win.deinited = true

	// OS可以直接销毁窗口, 这种情况下会先标记win.destroyed, 再调用Deinit()
	if win.ID != 0 && !win.destroyed {
		win.destroyNativeWindow()
	}
}

func (win *NaWindow) Destroy() {
	if !win.destroying {
		win.destroying = true
		// 默认的实现是不确认, 直接关闭
		win.invokeDeinitIfNot()
	}
}

func (win *NaWindow) invokeDeinitIfNot() {
	if !win.deinited {
		win.self.Deinit()
		debug.Assert(win.deinited, "NaWindow.Deinit() shoud be invoked in derived class's Deinit() function")
	}
}

// 收到销毁窗口事件
func (win *NaWindow) onNativeDestroy() {

	// "销毁窗口"是OS事件, 我们无法保证只收到1次. 至少在Windows系统里, 可以连续多次调用DestroyWindow.
	// 所以这里要判断 win.destroyed 标记
	if win.ID == 0 || win.destroyed {
		return // already destroyed
	}

	debug.Printf("NaWindow.onNativeDestroy(): %08X", win.ID)
	debug.Assert(IsMainThread(), "NaWindow.onNativeDestroy() must be called on main thread")

	// 保证不多次销毁.
	win.destroyed = true

	// 确保Deinit()在"销毁"前被调用
	win.invokeDeinitIfNot()

	delete(winMap, win.ID)
	win.releaseOtherNativeResource()
	numWnds--
}

func (win *NaWindow) onNativeRequestDestroy() {
	debug.Printf("NaWindow.onNativeRequestDestroy(): %08X", win.ID)
	win.self.Destroy()
}

func getWindow(hWnd uintptr) *NaWindow {
	return winMap[hWnd]
}

func (win *NaWindow) WindowSize() (float32, float32) {
	return win.width, win.height
}

func (win *NaWindow) SetSelf(self interface{}) {
	win.self = self.(IWindow)
}

func (win *NaWindow) Self() interface{} {
	return win.self
}

// IsDestroying 标识 窗口对象(Go+Native)是否已经开始销毁
func (win *NaWindow) IsDestroying() bool {
	return win.destroying
}

// IsDeinited 标识 Deinit() 是否被调用
func (win *NaWindow) IsDeinited() bool {
	return win.deinited
}

func (win *NaWindow) IsCursorConfined() bool {
	return win.confineCursor
}

func (win *NaWindow) IsCursorVisible() bool {
	return !win.hideCursor
}

func (win *NaWindow) IsWindowActive() bool {
	return win.active
}

func NumCPU() int {
	return runtime.NumCPU()
}

// func OSVersion() (s string) {
// 	return runtime.GOOS + "/" + runtime.GOARCH
// }

// Run the main event loop
func Run(appStart func() error) error {
	if started {
		panic("func Run() re-enter")
	}
	started = true

	mainLoop(appStart)

	return exitErr
}
