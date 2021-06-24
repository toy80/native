#ifdef TOY80_CGO

// macOS 的 cocoa 开发要点：
// * 开发语言是 Objective-C, 编译时要打开 -xobjective-c 开关
// * 我们不用xib/nib/storyboard, 也不带额外资源，这样不用打包为app即可运行 （和其他OS习惯一样）
// TODO： 

#import <Foundation/Foundation.h>
#import <Cocoa/Cocoa.h>
#import <QuartzCore/CAMetalLayer.h>
#import <AppKit/AppKit.h>

#include <dispatch/dispatch.h>
#include "native-cgo.h"


typedef struct NaWndData {
  uintptr_t view;    // NaWndData.ID  in go
  uintptr_t wc; // NaWndData.Ctx in go
  int32_t btnDown;
  float mouseX;
  float mouseY;
} NaWndData;

static dispatch_queue_t main_q;
static void updateKeyCode(uint16_t c, bool down, NSEventModifierFlags m);

@class NaWindowController;
@class NaView;

@interface NaAppDelegate : NSObject <NSApplicationDelegate>
@end

@interface NaWindowController : NSWindowController <NSWindowDelegate>
{
@public
  NaView * glview;
}
@end

@interface NaWindow : NSWindow
{
@public
//  NSTrackingArea* _ta;
}
@end


@interface NaViewController : NSViewController
@end

@interface NaView : NSView
{
@public
  NaWindowController* _wc;
//  NSTrackingArea* _ta;
}
@end

@implementation NaAppDelegate
- (void)applicationDidFinishLaunching:(NSNotification *)aNotification {
  na_emit_app_start();
}

- (BOOL)applicationShouldTerminateAfterLastWindowClosed:(NSApplication *)sender {
  return TRUE;
}

- (void)applicationWillTerminate:(NSNotification *)aNotification {
  //na_on_exit(0);
}
@end


@implementation NaViewController{
	CVDisplayLinkRef	_displayLink;
}


-(void) dealloc {
	CVDisplayLinkRelease(_displayLink);
	[super dealloc];
}

- (void)viewDidLoad {
  if (_displayLink) {
    return;
  }

  [super viewDidLoad];

  self.view.wantsLayer = YES;

  CVDisplayLinkCreateWithActiveCGDisplays(&_displayLink);
	CVDisplayLinkSetOutputCallback(_displayLink, &DisplayLinkCallback, self.view);
	CVDisplayLinkStart(_displayLink);
}

// - (void)viewWillAppear {
//   [super viewWillAppear];
// }

static CVReturn DisplayLinkCallback(CVDisplayLinkRef displayLink,
									const CVTimeStamp* now,
									const CVTimeStamp* outputTime,
									CVOptionFlags flagsIn,
									CVOptionFlags* flagsOut,
									void* target) {
  if(!target) {
    return kCVReturnError;
  }
	dispatch_sync(main_q, ^{ 
    na_emit_expose((uintptr_t) target, 0, 0, 0, 0);
  });
	return kCVReturnSuccess;
}


@end

@implementation NaView

- (instancetype)initWithFrame:(NSRect)frameRect {
    self = [super initWithFrame:frameRect];
    if (self) {
      NSTrackingArea * ta = [[NSTrackingArea alloc]
        initWithRect:(NSRect)self.bounds
        options: (/*NSTrackingActiveInKeyWindow |*/ NSTrackingActiveAlways | NSTrackingMouseEnteredAndExited |
          NSTrackingMouseMoved | NSTrackingInVisibleRect)
        owner: self
        userInfo: nil];
      [self addTrackingArea: ta];
      [ta release];
    }
    return self;
}

- (BOOL)wantsUpdateLayer {
  return YES;
}

+ (Class)layerClass {
  return [CAMetalLayer class];
}

- (CALayer *)makeBackingLayer {
  return [self.class.layerClass layer];
}


- (BOOL) acceptsFirstResponder {
  return YES;
}

- (BOOL)isOpaque {
    return YES;
}

- (void)drawRect:(NSRect)dirtyRect {
    [super drawRect:dirtyRect];
    float x = dirtyRect.origin.x;
    float y = dirtyRect.origin.y;
    float w = dirtyRect.size.width;
    float h = dirtyRect.size.height;
    y = self.bounds.size.height - y;
    na_emit_expose((uintptr_t)self, x, y, w, h);
}

// - (void)viewDidEndLiveResize {
//   [super viewDidEndLiveResize];
//   CGSize sz = _bounds.size;
//   na_on_resize(_wc, sz.width, sz.height);
// }

- (void)mouseEntered:(NSEvent *)theEvent {
  NSPoint pt = [self convertPoint:[theEvent locationInWindow] fromView:nil];
  pt.y = self.bounds.size.height - pt.y;
  na_emit_mouse_enter((uintptr_t)self, pt.x, pt.y);
}

- (void)mouseMoved:(NSEvent *)theEvent {
  [super mouseMoved: theEvent];
  NSPoint pt = [self convertPoint:[theEvent locationInWindow] fromView:nil];
  pt.y = self.bounds.size.height - pt.y;
  na_emit_mouse_move((uintptr_t)self, pt.x, pt.y);
}

- (void)mouseExited:(NSEvent *)theEvent {
  NSPoint pt = [self convertPoint:[theEvent locationInWindow] fromView:nil];
  pt.y = self.bounds.size.height - pt.y;
  na_emit_mouse_leave((uintptr_t)self, pt.x, pt.y);
}

- (void)mouseDown:(NSEvent *)theEvent {
  NSPoint pt = [self convertPoint:[theEvent locationInWindow] fromView:nil];
  pt.y = self.bounds.size.height - pt.y;
  na_emit_mouse_press((uintptr_t)self, NA_MOUSE_BTN_LEFT, pt.x, pt.y);
}

- (void)mouseUp:(NSEvent *)theEvent {
  NSPoint pt = [self convertPoint:[theEvent locationInWindow] fromView:nil];
  pt.y = self.bounds.size.height - pt.y;
  na_emit_mouse_release((uintptr_t)self, NA_MOUSE_BTN_LEFT, pt.x, pt.y);
}

- (void)rightMouseDown:(NSEvent *)theEvent {
  NSPoint pt = [self convertPoint:[theEvent locationInWindow] fromView:nil];
  pt.y = self.bounds.size.height - pt.y;
  na_emit_mouse_press((uintptr_t)self, NA_MOUSE_BTN_RIGHT, pt.x, pt.y);
}

- (void)rightMouseUp:(NSEvent *)theEvent {
  NSPoint pt = [self convertPoint:[theEvent locationInWindow] fromView:nil];
  pt.y = self.bounds.size.height - pt.y;
  na_emit_mouse_release((uintptr_t)self, NA_MOUSE_BTN_RIGHT, pt.x, pt.y);
}

- (void)keyDown:(NSEvent *)event {
  // na_printf("%s %04X %08X\n", "keyDown", event.keyCode, event.modifierFlags);
  updateKeyCode(event.keyCode, true, event.modifierFlags);
}


- (void)keyUp:(NSEvent *)event {
  //na_printf("%s %04X %08X\n", "keyUp", event.keyCode, event.modifierFlags);
  updateKeyCode(event.keyCode, false, event.modifierFlags);
}

@end

static uint8_t g_keyStates[256];

// cocoa的keyCode映射为Windows的VK code  // TODO: 键盘右半还没映射， 从TGB列开始
static uint8_t g_keyCodeMap[256] = {
  0x41,0x53,0x44,0x46, 0x5A,0x58,0x43,0x56, 0,0,0,0, 0x51,0x57,0x45,0x52,
  0,0,0x31,0x32, 0x33,0x34,0x36,0x35, 0, 0x39,0x37,0,0x38, 0x30,0,0,
  0,0,0,0, 0x0D,0,0,0, 0,0,0,0, 0,0,0,0,
  0x09,0x20,0,0x2E, 0, 0x1B,0,0, 0,0,0,0, 0,0,0,0,
  0,0,0,0, 0,0,0,0, 0,0,0,0, 0,0,0,0,
  0,0,0,0, 0,0,0,0, 0,0,0,0, 0,0,0,0,
  0,0,0,0, 0,0,0,0, 0,0,0,0, 0,0,0,0,
  0,0,0,0, 0,0,0,0, 0,0,0,0, 0,0,0,0,
  0,0,0,0, 0,0,0,0, 0,0,0,0, 0,0,0,0,
  0,0,0,0, 0,0,0,0, 0,0,0,0, 0,0,0,0,
  0,0,0,0, 0,0,0,0, 0,0,0,0, 0,0,0,0,
  0,0,0,0, 0,0,0,0, 0,0,0,0, 0,0,0,0,
  0,0,0,0, 0,0,0,0, 0,0,0,0, 0,0,0,0,
  0,0,0,0, 0,0,0,0, 0,0,0,0, 0,0,0,0,
  0,0,0,0, 0,0,0,0, 0,0,0,0, 0,0,0,0,
  0,0,0,0, 0,0,0,0, 0,0,0,0, 0,0,0,0,
};

static void updateKeyCode1(uint16_t c, uint8_t downCode, NSEventModifierFlags m) {

  if (m & NSEventModifierFlagNumericPad) // numeric keypad or an arrow key
  {
    switch (c) {
    case 0x7B: g_keyStates[0x25] = downCode; break;// left
    case 0x7E: g_keyStates[0x26] = downCode; break; // up
    case 0x7C: g_keyStates[0x27] = downCode; break;// right
    case 0x7D: g_keyStates[0x28] = downCode; break;// down
    }
    return;
  }

  if (m & NSEventModifierFlagFunction) {
    switch (c) {
      case 0x6F: g_keyStates[0x7B] = downCode; break; 
      case 0x67: g_keyStates[0x7A] = downCode; break; 
      case 0x6D: g_keyStates[0x79] = downCode; break; 
      case 0x65: g_keyStates[0x78] = downCode; break; 
      case 0x64: g_keyStates[0x77] = downCode; break; 
      case 0x62: g_keyStates[0x76] = downCode; break; 
      case 0x61: g_keyStates[0x75] = downCode; break; 
      case 0x0F: g_keyStates[0x74] = downCode; break; 
      case 0x76: g_keyStates[0x73] = downCode; break; 
      case 0x63: g_keyStates[0x72] = downCode; break; 
      case 0x78: g_keyStates[0x71] = downCode; break; 
      case 0x7A: g_keyStates[0x70] = downCode; break; 
    }
    return;
  }


  if(m & NSEventModifierFlagShift) {
    g_keyStates[0x10] = 0x10;
    g_keyStates[0xA0] = 0x10;
    g_keyStates[0xA0] = 0x10;
  } else {
    g_keyStates[0x10] = 0x00;
    g_keyStates[0xA0] = 0x00;
    g_keyStates[0xA0] = 0x00;
  }

  if (m & NSEventModifierFlagControl) {
    g_keyStates[0x02] = 0x10;
    g_keyStates[0xA2] = 0x10;
    g_keyStates[0xA3] = 0x10;
  } else {
    g_keyStates[0x02] = 0x00;
    g_keyStates[0xA2] = 0x00;
    g_keyStates[0xA3] = 0x00;
  }

  if (c >= 256) {
    return;
  }

  g_keyStates[g_keyCodeMap[c]] = downCode;
}

static void updateKeyCode(uint16_t c, bool down,NSEventModifierFlags m) {
  updateKeyCode1(c, down ? 0x10 : 0x00, m);
}

int na_get_keyboard_state(uint8_t keyStates[256]) {
  // 由于找不到类似于Windows GetKeyboardState的接口，我们通过NSView的keyDown和keyUp来模拟。
  // 此处有一个问题，单独按下CTRL等修饰键不产生keyDown或keyUp， 所以不能单独使用这些键。
  memcpy(keyStates, g_keyStates, 256);
  return 1; 
}


@implementation NaWindowController
- (void)windowDidChangeOcclusionState:(NSNotification *)notification {
  assert(na_is_main_thread());
  CGSize sz = glview.frame.size;
  na_emit_resize((uintptr_t)(void*)glview, sz.width, sz.height);
}

- (void)windowDidBecomeKey:(NSNotification *)notification {
  [self.window makeFirstResponder: self.window.contentView];
  self.window.acceptsMouseMovedEvents = YES;
    // na_printf("%s\n", "windowDidBecomeKey");
}

@end

@implementation NaWindow 
@end


na_bool_t na_init_library(uintptr_t sizeNaWndData) {
  na_printf("native.C.na_init_library()\n");
  assert(sizeNaWndData == sizeof(NaWndData));
  main_q = dispatch_get_main_queue();
  return 1;
}

void na_deinit_library() {}

void na_get_screen_size(int *width, int *height) {
  NSScreen *mainScreen = [NSScreen mainScreen];
  NSRect screenRect = [mainScreen visibleFrame];
  if (width) {
    *width = screenRect.size.width;
  }
  if (height) {
    *height = screenRect.size.height;
  }
}

na_bool_t na_get_mouse_pos(uintptr_t win, float *x, float *y) {
  if (win == 0 || (!x && !y)) {
    return 0;
  }

  // TODO:

  return 0;
}

na_bool_t na_create_window(NaWndData *win, uint64_t ws, int x, int y,
                            int width, int height) { @autoreleasepool{
  
  int sw=0, sh=0;
  na_get_screen_size(&sw, &sh);
  if (width <= 0) {
    width = sw / 2;
  }
  if (height <= 0) {
    height = sh / 2;
  }
  int x = (sw - width) / 2;
  int y = (sh - height) / 2;

  NSRect frame = NSMakeRect(x, y, width, height);
  NSWindowStyleMask mask = NSWindowStyleMaskClosable | NSWindowStyleMaskTitled | NSWindowStyleMaskMiniaturizable;
  if(ws & NA_WIN_HINT_RESIZABLE)
    mask |= NSWindowStyleMaskResizable;
  NaWindow* window = [[NaWindow alloc] initWithContentRect:frame
                                                  styleMask: mask
                                                    backing:NSBackingStoreBuffered
                                                      defer:NO];


  NaView* view = [[NaView alloc] initWithFrame: [window frame]];
  // // [pf release];
  // if (sharedOpenGLContex != nil) {
  //   NSOpenGLContext *ctx = [[NSOpenGLContext alloc]
  //                   initWithFormat: pf
  //                   shareContext:sharedOpenGLContex];
  //   view.openGLContext = ctx;
  // } else {
  //   sharedOpenGLContex = view.openGLContext;
  // }

  NaViewController* vc = [[NaViewController alloc] initWithNibName: nil bundle: nil];
  vc.view = view;

  window.contentViewController = vc;
  NaWindowController* wc = [[NaWindowController alloc] initWithWindow: window];
  window.delegate = wc;

  //wc->_assoc_data = assoc_data;
  view->_wc = wc;
  wc->glview = view;

  // NSUInteger count = [[[NSApplication sharedApplication] windows] count];
  win->view = (uintptr_t)(void*)view;
  win->wc   = (uintptr_t)(void*)wc;

  [vc viewDidLoad];

  return 1;
}}

void na_destroy_window(NaWndData *win) {
  if (!win) {
    return;
  }
  NaWindowController *wc = (NaWindowController *)win->wc;
  if (!wc) {
    return;
  }
  [wc close];
}

void na_show_window(NaWndData *win, int visible) {
  if (!win) {
    return;
  }
  NaWindowController *wc = (NaWindowController *)win->wc;
  if (!wc) {
    return;
  }
  if (visible) {
    [wc.window makeKeyAndOrderFront:nil];
  } else {
    // TODO:
  }
}

na_bool_t na_is_window_visible(NaWndData *win) { 
  if (!win) {
    return 0;
  }
  NaWindowController *wc = (NaWindowController *)win->wc;
  if (!wc) {
    return 0;
  }

  return wc.window.visible;
}

void na_set_window_title(NaWndData *win, const char *title) {
  if (!win) {
    return;
  }
  NaWindowController *wc = (NaWindowController *)win->wc;
  if (!wc) {
    return;
  }

  NSString* s = [[NSString alloc] initWithUTF8String: title];
  wc.window.title = s;
  [s release];
}

na_bool_t na_is_full_screen(NaWndData *win) { 
  if (!win) {
    return 0;
  }
  NaWindowController *wc = (NaWindowController *)win->wc;
  if (!wc) {
    return 0;
  }
  
  return (wc.window.styleMask & NSWindowStyleMaskFullScreen) == NSWindowStyleMaskFullScreen;
}

void na_set_focus(NaWndData *win) {
  if (!win) {
    return;
  }
  NaWindowController *wc = (NaWindowController *)win->wc;
  if (!wc) {
    return;
  }
    
  [wc.window makeKeyAndOrderFront: nil];
}

void na_toggle_full_screen(NaWndData *win) {
  if (!win) {
    return;
  }
  NaWindowController *wc = (NaWindowController *)win->wc;
  if (!wc) {
    return;
  }
  [wc.window toggleFullScreen:nil];
}

void na_quit_loop() {
  [NSApplication.sharedApplication terminate:nil];
}

// char *na_os_version() { 
//   NSOperatingSystemVersion version = NSProcessInfo.processInfo.operatingSystemVersion;
//   char* buf = (char*)malloc(128);
//   sprintf(buf, "MacOS %ld.%ld.%ld", version.majorVersion, version.minorVersion, version.patchVersion);
//   return buf;
// }

void na_expose_window(NaWndData *win, float x, float y, float width, float height) {
  if (!win) {
    return;
  }
  NaWindowController *wc = (NaWindowController *)win->wc;
  if (!wc) {
    return;
  }

  y =  wc->glview.bounds.size.height - y;
  NSRect invalidRect = NSMakeRect(x, y, width, height);
  [wc->glview setNeedsDisplayInRect: invalidRect];
}

na_bool_t na_is_main_thread() { return (na_bool_t) [NSThread isMainThread]; }

int na_dispatch_event() {
		return 0;
}

void na_yield() {}


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

int na_event_loop() {
  NSApplication * app = [NSApplication sharedApplication];
  // set the app delegate, don't need release.
  [app setDelegate:[[NaAppDelegate alloc] init]];
  // make window popup in the front as normal app
  [app setActivationPolicy:NSApplicationActivationPolicyRegular];
  [app activateIgnoringOtherApps: YES];
  [app run];
  return EXIT_SUCCESS;
}

void na_dispatch_sync() {
	dispatch_sync(main_q, ^{ na_dispatch_sync_callback(); });
}

void na_dispatch_async() {
	dispatch_async(main_q, ^{ na_dispatch_async_callback(); });
}


void na_confine_cursor(NaWndData *win, na_bool_t confine)
{
  // TODO: 
	if (confine) 
	{
	} 
	else 
	{
	}	
}

void na_show_cursor(NaWndData *win, na_bool_t show, float x, float y)
{
  // TODO: 
  
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


#endif
