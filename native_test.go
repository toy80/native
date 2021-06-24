package native

import (
	"testing"
)

func TestIsMainThread(t *testing.T) {
	_ = IsMainThread()
}

func TestGetCursorPos(t *testing.T) {
	x, y, err := GetCursorPos(88)
	if err == nil {
		t.Fatalf("expect err == nil, got %v", err)
	}
	if x != 0 || y != 0 {
		t.Fatalf("expect x == 0 and y == 0, got x=%f, y=%f", x, y)
	}
}

func TestGetScreenSize(t *testing.T) {
	w, h, err := GetScreenSize(0)
	if err != nil {
		t.Fatalf("expect err == nil, got %v", err)
	}
	if w == 0 || h == 0 {
		t.Fatalf("expect w != 0 and h != 0, got w=%f, h=%f", w, h)
	}
}

func TestYield(t *testing.T) {
	YieldThread()
}
