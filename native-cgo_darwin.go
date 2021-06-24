// +build darwin

package native

// #cgo CFLAGS: -xobjective-c
// #cgo LDFLAGS: -framework Foundation -framework Cocoa -framework QuartzCore -framework AppKit
// #include "native-cgo.h"
import "C"
import (
	"errors"
)

const slowRelativeMouse = true

type NaWndData struct {
	WindowID
	btnDown int32
	mouseX  float32
	mouseY  float32
}

// 用来把函数传给主线程调用, darwin版用GDC来实现, 所以结构里没有chan
type mtcall struct {
	fn  func() error
	err error
}

var (
	mainSyncQ        = make(chan *mtcall, 1)
	mainAsyncQ       = make(chan func() error, 128)
	callbackAppStart func() error
)

//export na_dispatch_sync_callback
func na_dispatch_sync_callback() {
	cb := <-mainSyncQ
	cb.err = cb.fn()
}

func MainDispatchSync(fn func() error) error {
	if fn == nil {
		return errors.New("nil func")
	}

	if IsMainThread() {
		// invoke directly
		return fn()
	}

	// macOS用dispatch_sync来实现相应的功能
	// fn 不能直接传给 Objective-C, 所以保存在 Go这边，并编上号
	cb := &mtcall{fn: fn}

	mainSyncQ <- cb

	C.na_dispatch_sync()

	return cb.err
}

//export na_dispatch_async_callback
func na_dispatch_async_callback() {
	fn := <-mainAsyncQ
	fn()
}

func MainDispatchAsync(fn func() error) {
	if fn == nil {
		return
	}

	// 对于异步调用来说，即使在主线程也不直接调用

	mainAsyncQ <- fn
	C.na_dispatch_async()
}

func mainLoop(appStart func() error) {
	callbackAppStart = appStart
	C.na_event_loop()
}

//export na_emit_app_start
func na_emit_app_start() {
	if callbackAppStart == nil {
		return
	}
	if err := callbackAppStart(); err != nil {
		Exit(err)
	}
}
