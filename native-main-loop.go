// +build linux windows

package native

import (
	"errors"
	"log"
	"time"

	"github.com/toy80/debug"
)

// 主循环, 适用于linux和windows, darwin则直接调用Framework的功能

type mtcall struct {
	result *chan error
	fn     func() error
}

var (
	mainSyncQ  = make(chan mtcall, 1)
	mainAsyncQ = make(chan mtcall, 128)
)

func mainLoop(appStart func() error) {
	if appStart != nil {
		if err := appStart(); err != nil {
			exitErr = err
			return
		}
	}
	if numWnds == 0 && !KeepRunningWithoutWindow {
		log.Println("no window created, the program may exit immediately.")
	}
	mainLoop1(nil)
}

func mainLoop1(quit <-chan interface{}) (ret interface{}) {
	debug.Trace(0)
	if !IsMainThread() {
		panic("Loop only work in GUI thread")
	}

	//numBusy := 0

	t0 := time.Now()

_OUTER:
	for {
		// poll and procceed events
		hasEvent := false
		switch DispatchEvent() {
		case EventRequestToQuit:
			break _OUTER
		case EventDispatched:
			hasEvent = true
		case EventQueueEmpty:
			hasEvent = false
		}

		// if received quit signal from the chan, break the loop
		select {
		case x := <-quit:
			ret = x
			break _OUTER
		default:
		}

		// if the program (should be) stop, break the loop
		if !Running() {
			ret = 0
			break
		}

		// invoke callbacks and draw frame
		//	hasCallback := guiThreadRoutine()
		hasCallback := false

		busy := hasEvent || hasCallback
		// busy = emitTimerTick() || busy
		toDrawFrame := false
		if busy {
			t1 := time.Now()
			span := t1.Sub(t0)
			toDrawFrame = span > time.Millisecond
			//log.Println("toDrawFrame=", toDrawFrame, "span=", span)
		} else {
			toDrawFrame = true
		}

		if toDrawFrame {
			busy = drawVideoFrame() || busy
			t0 = time.Now()
		}

		// sleep when idle
		if !busy {
			YieldThread()
		}
	}
	return
}

func guiThreadRoutine() (act bool) {
	// TODO: handle panic/recover here?

	// invoke callbacks
	var cb mtcall
	select {
	case cb = <-mainSyncQ:
	case cb = <-mainAsyncQ:
	default:
	}
	if cb.fn != nil {
		act = true
		err := cb.fn()
		if cb.result != nil {
			select {
			case (*cb.result) <- err:
			default:
			}
		}
	}
	return
}

// func emitTimerTick() (busy bool) {
// 	for _, win := range winMap {
// 		if !win.IsDestroying() {
// 			win.emitTimerTick()
// 		}
// 	}
// 	return false
// }

func drawVideoFrame() (busy bool) {
	// draw each video window TODO: darwin 是每个窗口独立的循环, windows也应该是
	// TODO: map is unordered...
	for _, win := range winMap {
		if !win.destroying && !win.deinited && (win.hints&WinHintNoVideo) != WinHintNoVideo {
			// 如果有这个HINT, 那么即使画图也SLEEP, 效果就是CPU占用下降, FPS可能下降
			// if (win.hints & WinHintLimitFPS) == 0 {
			busy = true
			// }
			win.emitExpose(0, 0, win.width, win.height)
		}
	}
	return
}

func MainDispatchSync(fn func() error) error {
	if fn == nil {
		return errors.New("nil func")
	}
	// fmt.Println("native.IsGuiThread()=", native.IsGuiThread())
	if IsMainThread() {
		// invoke directly
		return fn()
	}
	// invoke via channel
	result := make(chan error)
	mainSyncQ <- mtcall{result: &result, fn: fn}
	return <-result
}

func mainDispatchAsyncWithResult(result *chan error, fn func() error) {
	if fn == nil {
		if result != nil {
			select {
			case (*result) <- errors.New("nil func"):
			default:
			}
			return
		}
	}
	mainAsyncQ <- mtcall{result: result, fn: fn}
}

func MainDispatchAsync(fn func() error) {
	mainDispatchAsyncWithResult(nil, fn)
}
