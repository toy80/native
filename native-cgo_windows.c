#ifdef TOY80_CGO

#include <stdlib.h>
#include <assert.h>
#include <stdio.h>
#include <string.h>
#include <windows.h>
#include <windowsx.h>

#include "native-cgo.h"

typedef struct NaWndData
{
	HWND      hWnd;		 // NaWndData.ID  in go
	uintptr_t _unused; // NaWndData.Ctx in go
	int32_t   btnDown;
	float     mouseX;
	float     mouseY;
	RECT      restoreRect;
	uint32_t  restoreStyle;
	uint32_t  restoreExStyle;
	uint8_t   trackMouse;
	uint8_t   mouseHover;
	uint8_t   fullScreen;
	uint8_t   _padding0;
} NaWndData;

#define NA_WINDOW_CLASS_NAME L"Toy80_WinL"

static DWORD     g_idMainThread = 0;
static HINSTANCE g_hInstance    = 0;

static LRESULT CALLBACK NaWindowProc(HWND hwnd, UINT uMsg, WPARAM wParam, LPARAM lParam);

int na_printf(const char *fmt, ...)
{
	char buf[4096];
	va_list ap;
	int ret;
	va_start(ap, fmt);
	ret = vsprintf(buf, fmt, ap);
	va_end(ap);
	na_report(buf, 0);
	return ret;
}

na_bool_t na_init_library(uintptr_t sizeNaWndData)
{
	assert(sizeof(HWND) == sizeof(uintptr_t));
	assert(sizeNaWndData == sizeof(NaWndData));
	na_printf("native.C.na_init_library()\n");
	g_idMainThread = (DWORD)GetCurrentThreadId();
	g_hInstance = GetModuleHandleW(NULL);

	// register the window class
	HCURSOR hCursorArrow = LoadCursor(NULL, IDC_ARROW);
	DWORD style = CS_OWNDC | CS_VREDRAW | CS_HREDRAW | CS_DBLCLKS | CS_DROPSHADOW;
	WNDCLASSEXW wc;
	memset(&wc, 0, sizeof(wc));
	wc.cbSize = sizeof(wc);
	wc.style = style;												 //
	wc.lpfnWndProc = &NaWindowProc;					 //
	wc.cbClsExtra = 0;											 //
	wc.cbWndExtra = 0;											 // note: USERDATA is different than extra bytes
	wc.hInstance = g_hInstance;							 //
	wc.hIcon = NULL;												 // TODO: allow set window icon
	wc.hCursor = hCursorArrow;							 // TODO: allow disable cursor
	wc.hbrBackground = NULL;								 // don't draw anything with GDI
	wc.lpszMenuName = NULL;									 //
	wc.lpszClassName = NA_WINDOW_CLASS_NAME; //
	wc.hIconSm = 0;													 // TODO:

	BOOL bOK = RegisterClassExW(&wc);
	if (!bOK)
	{
		na_printf("RegisterClassExW() failure.\n");
		na_deinit_library();
		return 0;
	}

	return 1;
}

void na_deinit_library()
{
	//
}

na_bool_t na_is_main_thread()
{
	DWORD idThread = (DWORD)GetCurrentThreadId();
	return g_idMainThread == idThread;
}

na_bool_t na_get_mouse_pos(uintptr_t win, float *x, float *y)
{
	if (win == 0 || !x && !y)
	{
		return 0;
	}
	POINT pt;
	if (!GetCursorPos(&pt))
	{
		return 0;
	}
	if (win != 0)
	{
		if (!ScreenToClient((HWND)win, &pt))
		{
			return 0;
		}
	}
	if (x)
	{
		*x = (float)pt.x;
	}
	if (y)
	{
		*y = (float)pt.y;
	}
	return 0;
}

na_bool_t na_create_window(NaWndData *win, uint64_t ws, int x, int y, int width, int height)
{
	DWORD style = WS_OVERLAPPED | WS_CAPTION | WS_SYSMENU;

	if ((ws & NA_WIN_HINT_RESIZABLE) != 0)
	{
		style |= WS_THICKFRAME  | WS_MINIMIZEBOX | WS_MAXIMIZEBOX;
	}

	if ((ws & NA_WIN_HINT_MINIMIZE_BOX) != 0)
	{
		style |= WS_MINIMIZEBOX;
	}

	if ((ws & NA_WIN_HINT_MAXIMIZE_BOX) != 0)
	{
		style |= WS_MAXIMIZEBOX;
	}

	RECT rc;
	rc.left = x;
	rc.top = y;
	rc.right = x + width;
	rc.bottom = y + height;
	AdjustWindowRectEx(&rc, style, 0, 0);
	// debug.Printf("Win32API.CreateWindow(%+v)", rc)
	win->hWnd = CreateWindowExW(0,
															NA_WINDOW_CLASS_NAME,
															L"Toy80",
															style,
															rc.left, rc.top, rc.right - rc.left, rc.bottom - rc.top,
															0, 0, g_hInstance, 0);
	// TODO: store NaWndData into window.USERDATA
	if (win->hWnd == 0)
	{
		na_printf("CreateWindowExW() failure.\n");
		na_deinit_library();
		return 0;
	}

	RAWINPUTDEVICE rid[1];
	rid[0].usUsagePage = ((USHORT) 0x01); // HID_USAGE_PAGE_GENERIC
	rid[0].usUsage = ((USHORT) 0x02);     // HID_USAGE_GENERIC_MOUSE
	rid[0].dwFlags = RIDEV_INPUTSINK;   
	rid[0].hwndTarget = win->hWnd;
	BOOL bOK = RegisterRawInputDevices(rid, 1, sizeof(rid[0]));
	if (!bOK)
	{
		na_printf("RegisterRawInputDevices() failure.\n");
		na_deinit_library();
		return 0;
	}

	return 1;
}

void na_destroy_window(NaWndData *win)
{
	DestroyWindow((HWND)win->hWnd);
}

// static int btnNum(xcb_button_t btn)
// {
//   switch (btn)
//   {
//   case 1:
//     return NA_MOUSE_BTN_LEFT;
//   case 2:
//     return NA_MOUSE_BTN_MIDDLE;
//   case 3:
//     return NA_MOUSE_BTN_RIGHT;
//   default:
//     return 0;
//   }
// }

int na_dispatch_event()
{
	MSG msg;
	if (PeekMessageW(&msg, 0, 0, 0, PM_REMOVE))
	{
		if (msg.message != WM_QUIT)
		{
			TranslateMessage(&msg);
			DispatchMessageW(&msg);
			return NA_EVENT_DISPATCHED;
		}
		return NA_EVENT_REQUEST_TO_QUIT;
	}
	return NA_EVENT_QUEUE_EMPTY;
}

LRESULT CALLBACK NaWindowProc(HWND hWnd, UINT message, WPARAM wParam, LPARAM lParam)
{
	switch (message)
	{
	case WM_DESTROY:
		na_emit_destroy((uintptr_t)hWnd);
		return 0;
	case WM_PAINT:
	{
		PAINTSTRUCT ps;
		BeginPaint(hWnd, &ps);
		RECT rc = ps.rcPaint;
		na_emit_expose((uintptr_t)hWnd, rc.left, rc.top, rc.right - rc.left, rc.bottom - rc.top);
		EndPaint(hWnd, &ps);
	}
		return 0;
	case WM_SIZE:
		switch (wParam)
		{
		case SIZE_MINIMIZED:
			na_emit_appear((uintptr_t)hWnd, 0);
			break;
		case SIZE_MAXHIDE:
			na_emit_appear((uintptr_t)hWnd, 1);
			break;
		case SIZE_MAXIMIZED:
		case SIZE_MAXSHOW:
		default: // SIZE_RESTORED
			na_emit_appear((uintptr_t)hWnd, 1);
			na_emit_resize((uintptr_t)hWnd, LOWORD(lParam), HIWORD(lParam));
		}
		return 0;
	case WM_CLOSE:
		na_emit_close((uintptr_t)hWnd);
		return 0;
	case WM_LBUTTONDOWN:
	{
		NaWndData *win = na_get_windata((uintptr_t)hWnd);
		if (win->trackMouse != 0 && win->btnDown == 0)
		{
			SetCapture(hWnd);
		}
		win->btnDown |= NA_MOUSE_BTN_LEFT;
		na_emit_mouse_press((uintptr_t)hWnd, NA_MOUSE_BTN_LEFT, GET_X_LPARAM(lParam), GET_Y_LPARAM(lParam));
	}
		return 0;
	case WM_LBUTTONUP:
	{
		NaWndData *win = na_get_windata((uintptr_t)hWnd);

		win->btnDown &= ~NA_MOUSE_BTN_LEFT;

		na_emit_mouse_release((uintptr_t)hWnd, NA_MOUSE_BTN_LEFT, GET_X_LPARAM(lParam), GET_Y_LPARAM(lParam));
		if (win->trackMouse != 0 && win->btnDown == 0)
		{
			ReleaseCapture();
		}
	}
		return 0;
	case WM_RBUTTONDOWN:
	{
		NaWndData *win = na_get_windata((uintptr_t)hWnd);
		if (win->trackMouse != 0 && win->btnDown == 0)
		{
			SetCapture(hWnd);
		}
		win->btnDown |= NA_MOUSE_BTN_RIGHT;
		// na_emit_mouse_press((uintptr_t)hWnd, NA_MOUSE_BTN_LEFT, GET_X_LPARAM(lParam), GET_Y_LPARAM(lParam));
		na_emit_mouse_press((uintptr_t)hWnd, NA_MOUSE_BTN_RIGHT, GET_X_LPARAM(lParam), GET_Y_LPARAM(lParam));
	}
		return 0;
	case WM_RBUTTONUP:
	{
		NaWndData *win = na_get_windata((uintptr_t)hWnd);
		win->btnDown &= ~NA_MOUSE_BTN_RIGHT;
		na_emit_mouse_release((uintptr_t)hWnd, NA_MOUSE_BTN_RIGHT, GET_X_LPARAM(lParam), GET_Y_LPARAM(lParam));
		if (win->trackMouse != 0 && win->btnDown == 0)
		{
			ReleaseCapture();
		}
	}
		return 0;
	case WM_MOUSEMOVE:
	{
		NaWndData *win = na_get_windata((uintptr_t)hWnd);
		if (win->trackMouse != 0 && win->mouseHover == 0)
		{
			win->mouseHover = 1;
			TRACKMOUSEEVENT tme = {
					0,
					TME_LEAVE, // TME_LEAVE
					hWnd,
					HOVER_DEFAULT, //HOVER_DEFAULT
			};
			tme.cbSize = sizeof(tme);
			TrackMouseEvent(&tme); // one shot
			na_emit_mouse_enter((uintptr_t)hWnd, GET_X_LPARAM(lParam), GET_Y_LPARAM(lParam));
		}
		na_emit_mouse_move((uintptr_t)hWnd, GET_X_LPARAM(lParam), GET_Y_LPARAM(lParam));
	}
		return 0;
	case WM_MOUSEWHEEL:
		na_emit_mouse_wheel((uintptr_t)hWnd, 1, GET_WHEEL_DELTA_WPARAM(wParam), GET_X_LPARAM(lParam), GET_Y_LPARAM(lParam));
		return 0;
	case 0x020e: //WM_MOUSEHWHEEL
		na_emit_mouse_wheel((uintptr_t)hWnd, 0, GET_WHEEL_DELTA_WPARAM(wParam), GET_X_LPARAM(lParam), GET_Y_LPARAM(lParam));
		return 0;
	case WM_MOUSELEAVE:
	{
		NaWndData *win = na_get_windata((uintptr_t)hWnd);
		if (win->trackMouse != 0 && win->mouseHover != 0)
		{
			win->mouseHover = 0;
			POINT pt = {0, 0};
			GetCursorPos(&pt);
			ScreenToClient(hWnd, &pt);
			na_emit_mouse_leave((uintptr_t)hWnd, pt.x, pt.y);
		}
	}
		return 0;
	case WM_GETMINMAXINFO:
	{
		MINMAXINFO *pMMI = (MINMAXINFO *)lParam;
		pMMI->ptMinTrackSize.x = 100;
		pMMI->ptMinTrackSize.y = 100;
	}
		return 0;
	case WM_SYSKEYDOWN:
		if (wParam == VK_F4 && (lParam&0x20000000) != 0)
		{ 
			break; // ALT+F4
		}
		return 0;
	case WM_KEYDOWN:
	case WM_KEYUP:
	case WM_SYSKEYUP:
		return 0; // 我们用自己的方式处理键盘事件, 以消除F10等默认按键的影响. 当然Win键之类是拦截不了的.
	case WM_ACTIVATE:
		na_emit_activate((uintptr_t)hWnd, wParam != 0);
		break;
	case WM_INPUT:
	{
		UINT dwSize = sizeof(RAWINPUT);
		static BYTE lpb[sizeof(RAWINPUT)];
		// na_printf("sizeof(RAWINPUT)=%d sizeof(RAWINPUTHEADER))=%d", dwSize, sizeof(RAWINPUTHEADER));
		GetRawInputData((HRAWINPUT)lParam, RID_INPUT, lpb, &dwSize, sizeof(RAWINPUTHEADER));

		RAWINPUT* raw = (RAWINPUT*)lpb;

		if (raw->header.dwType == RIM_TYPEMOUSE) 
		{
			int xPosRelative = raw->data.mouse.lLastX;
			int yPosRelative = raw->data.mouse.lLastY;
			na_emit_mouse_move_relative((uintptr_t)hWnd, (float)(xPosRelative), (float)(yPosRelative));
		} 
		break;
	}
	case WM_CHAR:
		na_emit_text_input((uintptr_t)hWnd, wParam);
		return 0;
	default:
		break;
	}

	return DefWindowProc(hWnd, message, wParam, lParam);
}

void na_yield()
{
	Sleep(1);
}

void na_quit_loop()
{
	PostQuitMessage(0);
}

char * utf16_to_utf8(const WCHAR * utf16)
{
	if (utf16 == 0) 
	{
		return 0;
	}
	int sz = WideCharToMultiByte(CP_UTF8, 0, utf16, -1, NULL, 0, NULL, NULL);
	char * utf8 = (char *) malloc(sz);	
	WideCharToMultiByte(CP_UTF8, 0, utf16, -1, utf8, sz, NULL, NULL);
	return utf8;
}

WCHAR * utf8_to_utf16(const char * utf8)
{
	if (utf8 == 0)
	{
		return 0;
	}
	int sz = MultiByteToWideChar(CP_UTF8, 0, utf8, -1, NULL, 0);
	WCHAR * utf16 = (WCHAR *) malloc(2*sz);	
	MultiByteToWideChar(CP_UTF8, 0, utf8, -1, utf16, sz);
	return utf16;
}

void na_set_window_title(NaWndData *win, const char *title)
{
	if (win == 0 || win->hWnd == 0)
	{
		return;
	}
	WCHAR * utf16 = utf8_to_utf16(title);
	SetWindowTextW(win->hWnd, utf16);
	free((void*)utf16);
}

// void na_get_size(NaWndData * win, float *width, float *height)
// {
//   if(width)
//   {
//     *width = 0;
//   }
//   if(height)
//   {
//     *height = 0;
//   }

// }

void na_show_window(NaWndData *win, int visible)
{
	if (win == 0 || win->hWnd == 0)
	{
		return;
	}
	ShowWindow((HWND)win->hWnd, visible);
}

int na_get_keyboard_state(uint8_t keyStates[256])
{
	return GetKeyboardState(keyStates);
}

void na_set_focus(NaWndData *win)
{
	if (win == 0 || win->hWnd == 0)
	{
		return;
	}
	SetFocus((HWND)win->hWnd);
}

void na_confine_cursor(NaWndData *win, na_bool_t confine)
{
	if (confine) 
	{
		RECT rect = {};
		BOOL ret = GetClientRect((HWND)win->hWnd, &rect);
		if (!ret) 
		{
			return;
		}

		ret = ClientToScreen((HWND)win->hWnd, (POINT*)(void*)(&rect));
		if (!ret) 
		{
			return;
		}

		ret = ClientToScreen((HWND)win->hWnd, (POINT*)((void*)(&rect)+8));
		if (!ret) 
		{
			return;
		}

		ClipCursor(&rect);
	} 
	else 
	{
		ClipCursor(0);
	}	
}

void na_show_cursor(NaWndData *win, na_bool_t show, float x, float y)
{
	if (show) 
	{
		int32_t n = -1;
		while (n < 0) 
		{
			n = ShowCursor(1);
		}
		POINT pt = {(int)(x), (int)(y)};
		ClientToScreen((HWND)win->hWnd, &pt);
		SetCursorPos(pt.x, pt.y);
	} 
	else 
	{
		int32_t n = 0;
		while (n >= 0) 
		{
			n = ShowCursor(0);
		}
	}	
}


void na_activate_window(NaWndData *win)
{
	SetActiveWindow((HWND)win->hWnd);
}

char * na_read_clipboard(NaWndData *win)
{
  if(!OpenClipboard((HWND)win->hWnd))
	{
		return 0;
	}
	HGLOBAL hMem = GetClipboardData(CF_UNICODETEXT);
	char * ret = 0;
	if (hMem != NULL)
	{
		WCHAR * lpszClipboard = (WCHAR*)GlobalLock(hMem);
		ret = utf16_to_utf8(lpszClipboard);
		GlobalUnlock(hMem);
	}
	CloseClipboard();
	return ret;
}

na_bool_t na_write_clipboard(NaWndData *win, const char *pstr)
{
  if(!OpenClipboard((HWND)win->hWnd))
	{
		return 0;
	}
	EmptyClipboard();
	WCHAR * pUtf16 = utf8_to_utf16(pstr);
	if (pUtf16 == NULL)
	{
		CloseClipboard();
		return 0;
	}

	int n = 0;
	for(const WCHAR *p=pUtf16; *p != 0; p++)
	{
		n++;
	}
	n++;

	HGLOBAL hMem = GlobalAlloc(GMEM_MOVEABLE, n*2);
	if (hMem == NULL)
	{
		free(pUtf16);
		CloseClipboard();
		return 0;
	}

	LPTSTR lptstrCopy = (WCHAR*)GlobalLock(hMem);
	memcpy(lptstrCopy, pUtf16, n*2);
	free(pUtf16);
	GlobalUnlock(hMem);

	if (!SetClipboardData(CF_UNICODETEXT, hMem))
	{
		GlobalFree(hMem);
	}

	if (!CloseClipboard())
	{
		return 0;
	}

	return 1;
}



#endif // TOY80_CGO
