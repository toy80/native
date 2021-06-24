#ifdef TOY80_CGO

#include <stdlib.h>
#include <stdio.h>
#include <stdarg.h>
#include <unistd.h>
#include <errno.h>
#include <assert.h>
#include <string.h>
#include <pthread.h>
#include <sys/utsname.h>
#include <xcb/xcb.h>
#include <xcb/xcb_icccm.h>

#include "native-cgo.h"

// TODO: 考虑用XGB代替XCB

typedef struct NaWndData
{
	uintptr_t wnd;  // NaWndData.ID  in go
	uintptr_t conn; // NaWndData.Ctx in go
	int32_t btnDown;
	float mouseX;
	float mouseY;
} NaWndData;


pthread_t g_mainThread;
xcb_connection_t *g_connection;
int g_preferedScreen;

xcb_atom_t atom_WM_PROTOCOLS;
xcb_atom_t atom_WM_DELETE_WINDOW;

// xcb_atom_t atom_NET_ACTIVE_WINDOW; 
// 注: X Window 似乎没有和 WM_ACTIVATE 直接对应的事件, _NET_ACTIVE_WINDOW 属性是存在root窗口下的.
// 如果要利用 _NET_ACTIVE_WINDOW 来模拟, 那么就要接收root的 property change 事件, 然后读_NET_ACTIVE_WINDOW, 看它是哪个窗口.
// 这样处理显然比较复杂, 也许我们应该换个思路, 例如在鼠标点击时SetFocus...
// 当然WM_ACTIVATE还有其他作用, 例如可以在窗口切到后台时降低帧率...

// xcb_intern_atom_reply_t *atom_NET_WM_NAME;

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
  na_printf("native.C.na_init_library()\n");
  assert(sizeNaWndData == sizeof(NaWndData));
  g_mainThread = pthread_self();
  g_connection = xcb_connect(NULL, &g_preferedScreen);
  int rc = xcb_connection_has_error(g_connection);
  if (rc != 0)
  {
    na_printf("xcb_connect() failure: err=%d\n", rc);
    na_deinit_library();
    return 0;
  }
  
  xcb_intern_atom_reply_t * reply = 0;
  reply = xcb_intern_atom_reply(g_connection, xcb_intern_atom(g_connection, 1, 12, "WM_PROTOCOLS"), 0);
  if (reply)
  {
    atom_WM_PROTOCOLS = reply->atom;
    free(reply);
  }
  reply = xcb_intern_atom_reply(g_connection, xcb_intern_atom(g_connection, 1, 16, "WM_DELETE_WINDOW"), 0);
  if (reply)
  {
    atom_WM_DELETE_WINDOW = reply->atom;
    free(reply);
  }

  // reply = xcb_intern_atom_reply(g_connection, xcb_intern_atom(g_connection, 1, 18, "_NET_ACTIVE_WINDOW"), 0);
  // if (reply)
  // {
  //   atom_NET_ACTIVE_WINDOW = reply->atom;
  //   free(reply);
  // }
  reply = 0;

  return 1;
}

void na_deinit_library()
{
  if (g_connection)
  {
    na_printf("native.C.na_deinit_library(): xcb_disconnect()\n");
    xcb_disconnect(g_connection);
    g_connection = 0;
  }
}

na_bool_t na_is_main_thread()
{
  return pthread_equal(g_mainThread, pthread_self());
}

na_bool_t na_get_mouse_pos(uintptr_t win, float *x, float *y)
{
  if (win == 0 || !x && !y)
  {
    return 0;
  }

  // TODO:
  // xcb_query_pointer_cookie_t xcb_query_pointer(xcb_connection_t *conn, xcb_window_t window);
  // xcb_query_pointer_reply_t *xcb_query_pointer_reply(xcb_connection_t *conn, xcb_query_pointer_cookie_t cookie, xcb_generic_error_t **e);

  return 0;
}

void na_yield()
{
  // sleep 1ms so that it is identical to Windows branch.
  // TODO: is pthread_yield works as expected?
  usleep(1000);
}

void na_quit_loop()
{
  // TODO:
}

na_bool_t na_create_window(NaWndData *win, uint64_t ws, int x, int y, int width, int height)
{
  const xcb_setup_t *setup = xcb_get_setup(g_connection);
  xcb_screen_iterator_t iter = xcb_setup_roots_iterator(setup);

  xcb_screen_t *screen = NULL;
  for (int i = 0; i < g_preferedScreen; i++)
  {
    xcb_screen_next(&iter);
  }
  screen = iter.data;

  if (screen == NULL)
  {
    return 0;
  }

  xcb_window_t wnd = xcb_generate_id(g_connection);
  uint32_t value_mask = /*XCB_CW_BACK_PIXEL |*/ XCB_CW_EVENT_MASK;
  uint32_t value_list[32];
  //value_list[0] = screen->black_pixel;
  value_list[0] =
      XCB_EVENT_MASK_EXPOSURE |
      XCB_EVENT_MASK_STRUCTURE_NOTIFY |
      XCB_EVENT_MASK_KEY_PRESS |
      XCB_EVENT_MASK_KEY_RELEASE |
      XCB_EVENT_MASK_BUTTON_PRESS |
      XCB_EVENT_MASK_BUTTON_RELEASE |
      XCB_EVENT_MASK_ENTER_WINDOW |
      XCB_EVENT_MASK_LEAVE_WINDOW |
      XCB_EVENT_MASK_POINTER_MOTION;

  // xcb_create_window(g_connection, XCB_COPY_FROM_PARENT, wnd,
  // 	screen->root, (int16_t)(x), (int16_t)(y),
  // 	(uint16_t)(width), (uint16_t)(height), 0,
  // 	XCB_WINDOW_CLASS_INPUT_OUTPUT, screen->root_visual,
  // 	value_mask, value_list);

  xcb_void_cookie_t cookie_window =
      xcb_create_window_checked(
          g_connection,
          screen->root_depth,
          wnd,
          screen->root,
          (int16_t)(x), (int16_t)(y), (uint16_t)(width), (uint16_t)(height),
          0,
          XCB_WINDOW_CLASS_INPUT_OUTPUT,
          screen->root_visual,
          value_mask, value_list);

  xcb_generic_error_t *error = xcb_request_check(g_connection, cookie_window);
  if (error)
  {
    na_printf("xcb_create_window_checked(%d,%d,%d,%d) failure: err=%d\n", x, y, width, height, error->error_code);
    free(error);
    na_deinit_library();
    return 0;
  }

  if ((ws & NA_WIN_HINT_RESIZABLE) == 0)
  {
    xcb_size_hints_t hints = {0};
    xcb_icccm_size_hints_set_size(&hints, 0, width, height);
    xcb_icccm_size_hints_set_min_size(&hints, width, height);
    xcb_icccm_size_hints_set_max_size(&hints, width, height);
    xcb_icccm_set_wm_size_hints(g_connection, wnd, XCB_ATOM_WM_NORMAL_HINTS, &hints);
  }
  else
  {
    // avoid zero area window
    xcb_size_hints_t hints = {0};
    xcb_icccm_size_hints_set_size(&hints, 0, width, height);
    xcb_icccm_size_hints_set_min_size(&hints, 100, 100);
    xcb_icccm_size_hints_set_max_size(&hints, 7680, 4320); // 8K UHD
    xcb_icccm_set_wm_size_hints(g_connection, wnd, XCB_ATOM_WM_NORMAL_HINTS, &hints);
  }


  // xcb_intern_atom_cookie_t cookie3 = xcb_intern_atom(g_connection, 0, 12, "_NET_WM_NAME");
  // atom_NET_WM_NAME = xcb_intern_atom_reply(g_connection, cookie3, 0);
  if (atom_WM_DELETE_WINDOW != 0 && atom_WM_PROTOCOLS != 0)
  {
    xcb_change_property(g_connection, XCB_PROP_MODE_REPLACE, wnd,
                        atom_WM_PROTOCOLS, 4, 32, 1, &(atom_WM_DELETE_WINDOW));
  }

  xcb_flush(g_connection);

  win->wnd = wnd;
  win->conn = (uintptr_t)(void *)g_connection;
  return 1;
}

void na_destroy_window(NaWndData *win)
{
  assert((xcb_connection_t *)win->conn == g_connection);
  xcb_destroy_window((xcb_connection_t *)win->conn, win->wnd);
  xcb_flush((xcb_connection_t *)win->conn);
}

static int btnNum(xcb_button_t btn)
{
  switch (btn)
  {
  case 1:
    return NA_MOUSE_BTN_LEFT;
  case 2:
    return NA_MOUSE_BTN_MIDDLE;
  case 3:
    return NA_MOUSE_BTN_RIGHT;
  default:
    return 0;
  }
}

na_bool_t na_dispatch_event()
{
  xcb_generic_event_t *evt = xcb_poll_for_event(g_connection);

  if (!evt)
  {
    return NA_EVENT_QUEUE_EMPTY;
  }

  switch (evt->response_type & ~0x80)
  {
  case XCB_CLIENT_MESSAGE:
  {
    xcb_client_message_event_t *e = (xcb_client_message_event_t *)evt;
    if (atom_WM_DELETE_WINDOW != 0 && e->data.data32[0] == atom_WM_DELETE_WINDOW)
    {
      na_emit_close(e->window); // TODO: destroy or close?
    }
    break;
  }
  case XCB_CONFIGURE_NOTIFY:
  {
    xcb_configure_notify_event_t *e = (xcb_configure_notify_event_t *)evt;
    na_emit_resize(e->event, e->width, e->height);
    break;
  }
  case XCB_MAP_NOTIFY:
  {
    xcb_map_notify_event_t *e = (xcb_map_notify_event_t *)evt;
    na_emit_appear(e->event, 1);
    break;
  }
  case XCB_UNMAP_NOTIFY:
  {
    xcb_unmap_notify_event_t *e = (xcb_unmap_notify_event_t *)evt;
    na_emit_appear(e->event, 0);
    break;
  }
  case XCB_DESTROY_NOTIFY:
  {
    xcb_destroy_notify_event_t *e = (xcb_destroy_notify_event_t *)evt;
    na_emit_destroy(e->window);
    break;
  }
  case XCB_EXPOSE:
  {
    xcb_expose_event_t *e = (xcb_expose_event_t *)evt;
    na_emit_expose(e->window, e->x, e->y, e->width, e->height);
    break;
  }
  case XCB_BUTTON_PRESS:
  {
    xcb_button_press_event_t *e = (xcb_button_press_event_t *)evt;
    int btn = btnNum(e->detail);
    if (btn != 0)
    {
      na_emit_mouse_press(e->event, btn, e->event_x, e->event_y);
    }
    //na_printf("XCB_BUTTON_PRESS detail=%X\n", e->detail);
    break;
  }
  case XCB_BUTTON_RELEASE:
  {
    xcb_button_release_event_t *e = (xcb_button_press_event_t *)evt;
    int btn = btnNum(e->detail);
    if (btn != 0)
    {
      na_emit_mouse_release(e->event, btn, e->event_x, e->event_y);
    }
    else if (e->detail == 4)
    {
      // up
      na_emit_mouse_wheel(e->event, 1, +120, e->event_x, e->event_y);
    }
    else if (e->detail == 5)
    {
      // down
      na_emit_mouse_wheel(e->event, 1, -120, e->event_x, e->event_y);
    }
    //na_printf("XCB_BUTTON_RELEASE detail=%X\n", e->detail);
    break;
  }
  case XCB_MOTION_NOTIFY:
  {
    xcb_motion_notify_event_t *e = (xcb_motion_notify_event_t *)evt;
    na_emit_mouse_move(e->event, e->event_x, e->event_y);
    break;
  }
  case XCB_KEY_PRESS:
  {
    xcb_key_press_event_t *e = (xcb_key_press_event_t *)evt;
    break;
  }
  case XCB_KEY_RELEASE:
  {
    xcb_key_release_event_t *e = (xcb_key_release_event_t *)evt;
    break;
  }
  default:
    break;
  }
  free(evt);
  return NA_EVENT_DISPATCHED;
}

void na_set_window_title(NaWndData *win, const char *title)
{
  if (win == 0 || win->wnd == 0)
  {
    return;
  }
  xcb_change_property((xcb_connection_t *)win->conn,
                      XCB_PROP_MODE_REPLACE,
                      (xcb_window_t)win->wnd,
                      XCB_ATOM_WM_NAME, // TODO: _NET_WM_NAME? xcb_icccm_set_wm_name_checked?
                      XCB_ATOM_STRING,
                      8,
                      strlen(title),
                      title);
  xcb_flush((xcb_connection_t *)win->conn);
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
  if (win == 0 || win->wnd == 0)
  {
    return;
  }
  if (visible)
  {
    xcb_map_window_checked((xcb_connection_t *)win->conn, (xcb_window_t)win->wnd);
  }
  else
  {
    xcb_unmap_window_checked((xcb_connection_t *)win->conn, (xcb_window_t)win->wnd);
  }
  xcb_flush((xcb_connection_t *)win->conn);
}

// 这个表格用来把X11的键码映射为Windows下的VK Code
// TODO: retrive from X11?
static uint8_t keyCodeMap[256] = {
  0,0,0,0, 0,0,0,0, 0,0x1B,'1', '2','3','4','5', '6',
  '7','8','9', '0', '-','=',0x08,0x09, 'Q','W','E','R', 'T','Y','U','I',
  'O','P','[',']', 0x0D, 0xA2, 'A','S','D', 'F', 'G', 'H', 'J', 'K', 'L', ';',
  '\'','`',0xA0, '\\', 'Z', 'X','C','V','B', 'N','M','<','>', '/',0xA1,0,
  0x12,0x20,0,0x70, 0x71,0x72,0x73,0x74, 0x75,0x76,0x77,0x78, 0x79,0,0,0,
  0,0,0,0, 0,0,0,0, 0,0,0,0, 0,0,0,0x7A,
  0x7B,0,0,0, 0,0,0,0, 0,0xA3,0,0, 0,0,0x24,0x26,
  0x21,0x25,0x27,0x23, 0x28,0x22,0x2D,0x2E, 0,0,0,0, 0,0,0,0x13,
  0,0,0,0, 0,0,0,0, 0,0,0,0, 0,0,0,0,
  0,0,0,0, 0,0,0,0, 0,0,0,0, 0,0,0,0,
  0,0,0,0, 0,0,0,0, 0,0,0,0, 0,0,0,0,
  0,0,0,0, 0,0,0,0, 0,0,0,0, 0,0,0,0,
  0,0,0,0, 0,0,0,0, 0,0,0,0, 0,0,0,0,
  0,0,0,0, 0,0,0,0, 0,0,0,0, 0,0,0,0,
  0,0,0,0, 0,0,0,0, 0,0,0,0, 0,0,0,0,
  0,0,0,0, 0,0,0,0, 0,0,0,0, 0,0,0,0,
};

int na_get_keyboard_state(uint8_t keyStates[256])
{

  xcb_query_keymap_cookie_t cookie = xcb_query_keymap(g_connection);
  xcb_query_keymap_reply_t *keys = xcb_query_keymap_reply(g_connection, cookie, NULL);

  size_t sz = sizeof(keys->keys);
  // for (size_t i = 0; i < sz; i++)
  // {
  //   if (keys->keys[i] != 0)
  //   {
  //     na_printf("sym %d %02X\n", i, (int)keys->keys[i]);
  //   }
  // }
  size_t count = sz * 8;
  if (count > 256)
  {
    count = 256;
  }

  // TODO: need to convert codes, xcb_key_symbols_get_keycode?

  for (size_t i = 0; i < count; i++)
  {
    uint8_t b = keys->keys[i / 8];
    uint8_t mask = (uint8_t)(1) <<  (i % 8);
    if ((b & mask) != 0) 
    {
      keyStates[keyCodeMap[i]] = 0x10;
       //na_printf("state %d 0x%02X= %02X\n", i, i, (int)keyStates[i] );
    }
    else 
    {
      keyStates[keyCodeMap[i]] = 0x00;
    }

  }

  free(keys);
}

void na_set_focus(NaWndData *win)
{
  /*xcb_void_cookie_t*/ xcb_set_input_focus((xcb_connection_t *)win->conn, XCB_INPUT_FOCUS_NONE, (xcb_window_t)win->wnd, XCB_CURRENT_TIME);
}


void na_confine_cursor(NaWndData *win, na_bool_t confine)
{
  // TODO: XGrabPointer 只能捕获鼠标, 不能把鼠标限制在矩形里

	if (confine) 
	{
    xcb_grab_pointer_cookie_t cookie;
    xcb_grab_pointer_reply_t *reply;

    cookie = xcb_grab_pointer(
        (xcb_connection_t *)win->conn,
        1, // 允许自身接收事件            
        (xcb_window_t)win->wnd,        
        XCB_NONE,            
        XCB_GRAB_MODE_ASYNC, 
        XCB_GRAB_MODE_ASYNC, 
        (xcb_window_t)win->wnd,            // confine_to 
        XCB_NONE,             
        XCB_CURRENT_TIME
    );

    if ((reply = xcb_grab_pointer_reply((xcb_connection_t *)win->conn, cookie, NULL))) {
        if (reply->status == XCB_GRAB_STATUS_SUCCESS)
            printf("successfully grabbed the pointer\n");
        free(reply);
    }
	} 
	else 
	{
		xcb_ungrab_pointer((xcb_connection_t *)win->conn, XCB_CURRENT_TIME); 
	}	
}

void na_show_cursor(NaWndData *win, na_bool_t show, float x, float y)
{
  // TODO: #include <X11/extensions/Xfixes.h>
  //   XFixesHideCursor(dpy, window);

	if (show) 
	{

	} 
	else 
	{

	}	
}


void na_activate_window(NaWndData *win)
{
	// TODO: 
}

char * na_read_clipboard(NaWndData *win)
{
  return 0; // TODO:
}

na_bool_t na_write_clipboard(NaWndData *win, const char *pstr)
{
  return 0; // TODO: 
}


#endif // TOY80_CGO
