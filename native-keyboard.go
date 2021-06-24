package native

import "fmt"

// 按键的编码
type KeyCode byte

const (
	// use capital 'A'~'Z' for character key.
	// use '0'~'9' for number key.
	// LeftButton   KeyCode = 0x01
	// RightButton  KeyCode = 0x02
	KeyBackspace KeyCode = 0x08
	KeyTab       KeyCode = 0x09
	KeyEnter     KeyCode = 0x0D
	//KeyShift        KeyCode = 0x10
	//KeyControl      KeyCode = 0x11
	KeyAlt          KeyCode = 0x12
	KeyPause        KeyCode = 0x13
	KeyEsc          KeyCode = 0x1B
	KeySpace        KeyCode = 0x20
	KeyPageUp       KeyCode = 0x21
	KeyPageDown     KeyCode = 0x22
	KeyEnd          KeyCode = 0x23
	KeyHome         KeyCode = 0x24
	KeyLeft         KeyCode = 0x25
	KeyUp           KeyCode = 0x26
	KeyRight        KeyCode = 0x27
	KeyDown         KeyCode = 0x28
	KeyInsert       KeyCode = 0x2D
	KeyDelete       KeyCode = 0x2E
	KeyF1           KeyCode = 0x70
	KeyF2           KeyCode = 0x71
	KeyF3           KeyCode = 0x72
	KeyF4           KeyCode = 0x73
	KeyF5           KeyCode = 0x74
	KeyF6           KeyCode = 0x75
	KeyF7           KeyCode = 0x76
	KeyF8           KeyCode = 0x77
	KeyF9           KeyCode = 0x78
	KeyF10          KeyCode = 0x79
	KeyF11          KeyCode = 0x7A
	KeyF12          KeyCode = 0x7B
	KeyLeftShift    KeyCode = 0xA0
	KeyRightShift   KeyCode = 0xA1
	KeyLeftControl  KeyCode = 0xA2
	KeyRightControl KeyCode = 0xA3
)

func (k KeyCode) IsModifier() bool {
	switch k {
	case KeyAlt, KeyLeftShift, KeyRightShift, KeyLeftControl, KeyRightControl:
		return true
	}
	return false
}

func (k KeyCode) String() string {
	if k >= 'A' && k <= 'Z' || k >= '0' && k <= '9' {
		return string([]byte{byte(k)})
	}
	if k >= KeyF1 && k <= KeyF12 {
		return fmt.Sprintf("F%d", int(k)-int(KeyF1)+1)
	}

	switch k {
	case KeyBackspace:
		return "BACK_SPACE" // 0x08
	case KeyTab:
		return "TAB" // 0x09
	case KeyEnter:
		return "ENTER" // 0x0D
	// case KeyShift:
	// 	return "SHIFT" // 0x10
	// case KeyControl:
	// 	return "CTRL" // 0x11
	case KeyAlt:
		return "ALT" // 0x12
	case KeyPause:
		return "PAUSE" // 0x13
	case KeyEsc:
		return "ESC" // 0x1B
	case KeySpace:
		return "SPACE" // 0x20
	case KeyPageUp:
		return "PAGE_UP" // 0x21
	case KeyPageDown:
		return "PAGE_DOWN" // 0x22
	case KeyEnd:
		return "END" // 0x23
	case KeyHome:
		return "HOME" // 0x24
	case KeyLeft:
		return "LEFT" // 0x25
	case KeyUp:
		return "UP" // 0x26
	case KeyRight:
		return "RIGHT" // 0x27
	case KeyDown:
		return "DOWN" // 0x28
	case KeyInsert:
		return "INS" // 0x2D
	case KeyDelete:
		return "DEL" // 0x2E
	case KeyLeftShift:
		return "SHIFT_L" // 0xA0
	case KeyRightShift:
		return "SHIFT_R" // 0xA1
	case KeyLeftControl:
		return "CTRL_L" // 0xA2
	case KeyRightControl:
		return "CTRL_R" // 0xA3
	default:
		return fmt.Sprintf("0x%02X", int(k))
	}

}

// 按键的状态
type KeyState byte

func (ks KeyState) IsDown() bool {
	return ks&0xF0 != 0
}

func (ks KeyState) IsToggled() bool {
	return ks&0x0F != 0
}

func (ks KeyState) String() string {
	return fmt.Sprintf("%02X", int(ks))
}

type KeyboardFlags int

func (x KeyboardFlags) IsRepeat() bool {
	return x&0x01 != 0
}

func (x KeyboardFlags) IsCtrlDown() bool {
	return x&0x02 != 0
}

func (x KeyboardFlags) IsShiftDown() bool {
	return x&0x04 != 0
}

func (x KeyboardFlags) IsAltDown() bool {
	return x&0x08 != 0
}

type Keyboard struct {
	self   interface{}
	States [256]KeyState

	lastKey  KeyCode
	lastTime int64
}

func (pad *Keyboard) Init(self interface{}) {
	pad.self = self
	GetKeyboardState(pad.States[:])
}

func (pad *Keyboard) PollKeyboard(timestamp int64) {
	var newStates [256]KeyState
	GetKeyboardState(newStates[:])

	i, ok := pad.self.(interface {
		OnRawKeyboard(KeyCode, KeyState, KeyboardFlags)
	})

	if !ok {
		copy(pad.States[:], newStates[:])
		return
	}

	var flags KeyboardFlags

	if newStates[int(KeyLeftControl)].IsDown() || newStates[int(KeyRightControl)].IsDown() {
		flags |= 0x02
	}

	if newStates[int(KeyLeftShift)].IsDown() || newStates[int(KeyRightShift)].IsDown() {
		flags |= 0x04
	}

	if newStates[int(KeyAlt)].IsDown() {
		flags |= 0x08
	}

	const repeatDelay = 400
	const repeatPeriod = 200

	for k, v := range newStates[:] {
		if v == pad.States[k] {
			continue
		}
		// 键盘状态变化, 触发
		code := KeyCode(k)
		pad.States[k] = v
		i.OnRawKeyboard(code, v, flags)
		if !code.IsModifier() {
			pad.lastKey = code
			pad.lastTime = timestamp + (repeatDelay - repeatPeriod)
		} else {
			pad.lastKey = 0
		}
	}

	// 重复发送
	if pad.lastKey != 0 && pad.States[int(pad.lastKey)].IsDown() {
		if pad.lastTime+repeatPeriod < timestamp || pad.lastTime > timestamp+repeatDelay {
			pad.lastTime = timestamp
			i.OnRawKeyboard(pad.lastKey, pad.States[int(pad.lastKey)], flags|0x01)
		}
	}
}

func (pad Keyboard) IsDown(key KeyCode) bool {
	return pad.States[key].IsDown()
}

func (pad Keyboard) IsToggled(key KeyCode) bool {
	return pad.States[key].IsToggled()
}
