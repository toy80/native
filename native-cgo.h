#ifndef _TOY80_NATIVE_LAYER_H_
#define _TOY80_NATIVE_LAYER_H_

#include <stdint.h>

typedef struct NaWndData NaWndData;

enum {
  NA_WIN_HINT_RESIZABLE      = 0x0001,
  NA_WIN_HINT_FULLSCREEN     = 0x0002,
  NA_WIN_HINT_NO_VIDEO       = 0x0004,
  NA_WIN_HINT_MINIMIZE_BOX   = 0x0008,
  NA_WIN_HINT_MAXIMIZE_BOX   = 0x0010,
};

enum {
  NA_MOUSE_BTN_LEFT   = 0x0001,
  NA_MOUSE_BTN_RIGHT  = 0x0002,
  NA_MOUSE_BTN_MIDDLE = 0x0004,
};

enum {
  NA_EVENT_DISPATCHED      = 0,
  NA_EVENT_REQUEST_TO_QUIT = 1,
  NA_EVENT_QUEUE_EMPTY     = 2,
};

typedef int na_bool_t;

void      na_get_screen_size(int *width, int *height);
na_bool_t na_get_mouse_pos(uintptr_t win, float *x, float *y);
na_bool_t na_create_window(NaWndData *win, uint64_t ws, int x, int y, int width, int height);
void      na_destroy_window(NaWndData *win);
void      na_show_window(NaWndData *win, int visible);
na_bool_t na_is_window_visible(NaWndData *win);
void      na_set_window_title(NaWndData *win, const char *title);
na_bool_t na_is_full_screen(NaWndData *win);
void      na_toggle_full_screen(NaWndData *win);
void      na_quit_loop();
void      na_expose_window(NaWndData *win, float x, float y, float width, float height);
void      na_set_focus(NaWndData *win);
void      na_show_cursor(NaWndData *win, na_bool_t show, float x, float y);
void      na_confine_cursor(NaWndData *win, na_bool_t confine);
void      na_activate_window(NaWndData *win);
char *    na_read_clipboard(NaWndData *win);
na_bool_t na_write_clipboard(NaWndData *win, const char *pstr);

na_bool_t na_is_main_thread();
void      na_yield();
na_bool_t na_init_library(uintptr_t sizeNaWndData);
void      na_deinit_library();
int       na_get_keyboard_state(uint8_t keyStates[256]);
int       na_printf(const char *fmt, ...); // 同na_report,支持printf语法
int       na_event_loop();

// char *na_os_version(); // use free to release memory



// event handlers is implement in native.go
extern NaWndData * na_get_windata(uintptr_t win);

extern void na_emit_app_start();
extern void na_emit_close(uintptr_t win);
extern void na_emit_destroy(uintptr_t win);
extern void na_emit_resize(uintptr_t win, float width, float height);
extern void na_emit_mouse_move(uintptr_t win, float x, float y);
extern void na_emit_mouse_move_relative(uintptr_t win, float dx, float dy);
extern void na_emit_mouse_press(uintptr_t win, int btn, float x, float y);
extern void na_emit_mouse_release(uintptr_t win, int btn, float x, float y);
extern void na_emit_mouse_wheel(uintptr_t win, int vertical, float dz, float x, float y);
extern void na_emit_mouse_enter(uintptr_t win, float x, float y);
extern void na_emit_mouse_leave(uintptr_t win, float x, float y);
extern void na_emit_expose(uintptr_t win, int x, int y, int width, int height);
extern void na_emit_appear(uintptr_t win, int visible);
extern void na_emit_activate(uintptr_t win, int visible);
extern void na_emit_text_input(uintptr_t win, int ch);
extern void na_report(char *msg, int panic);

// void na_message_box(uintptr_t win, const char *title, const char *msg);
// int na_confirm_box(uintptr_t win, const char *title, const char *msg);

#ifdef __APPLE__
  void na_dispatch_sync();
  extern void na_dispatch_sync_callback();
  void na_dispatch_async();
  extern void na_dispatch_async_callback();
#else 
  int na_dispatch_event();
#endif

#endif /* _TOY80_NATIVE_LAYER_H_ */
