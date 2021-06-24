package main

import (
	"log"

	"github.com/toy80/native"
)

func appStart() (err error) {
	w0 := new(native.NaWindow)
	w0.SetSelf(w0)
	if err = w0.Init(&native.WindowOptions{
		Hints: native.WinHintResizable, //  | native.WinHintLimitFPS,
	}); err != nil {
		log.Println("error: w0.Init:", err)
		return
	}
	w0.SetTitle("w0 resizable")
	w0.SetVisible(true)

	w1 := new(native.NaWindow)
	w1.SetSelf(w1)
	if err = w1.Init(&native.WindowOptions{
		// Hints: native.WinHintResizable, //  | native.WinHintLimitFPS,
	}); err != nil {
		log.Println("error: w1.Init:", err)
		return
	}
	w1.SetTitle("w1 fix frame")
	w1.SetVisible(true)

	return nil
}

func main() {
	log.SetFlags(0)
	native.Run(appStart)
}
