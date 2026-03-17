//go:build darwin

// Package ios implements the platform.Backend for iOS using UIKit via cgo.
//
// Architecture:
//   - Go exports functions that the Objective-C side calls (GoxGetTree)
//   - Go calls C functions to create and configure UIKit views
//   - The Objective-C bootstrap (AppDelegate) drives the app lifecycle
//
// The iOS native files (main.m, AppDelegate, bridge.m) are generated
// by the CLI into the ios/ directory and are never edited by the user.
package ios

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework UIKit -framework Foundation

#include <stdlib.h>

// These functions are implemented in bridge.m (generated native code).
// They are resolved at link time when building the iOS app.
//
// For now, we declare them here so Go code can reference them,
// but they won't be available until we link with the native side.

// View creation
long gox_create_view(void);
long gox_create_label(const char* text);

// View hierarchy
void gox_add_subview(long parent, long child);
void gox_set_root_view(long handle);

// Styling
void gox_set_background_color(long handle, const char* color);
void gox_set_frame(long handle, double x, double y, double w, double h);
void gox_set_padding(long handle, double top, double right, double bottom, double left);
void gox_set_font_size(long handle, double size);
void gox_set_font_weight(long handle, const char* weight);
void gox_set_text_color(long handle, const char* color);
void gox_set_text_align(long handle, const char* align);
void gox_set_border_radius(long handle, double radius);
void gox_set_opacity(long handle, double opacity);
void gox_set_flex_layout(long handle, const char* direction, const char* justify, const char* align);

// App lifecycle
void gox_run_app(int argc, char* argv[]);
*/
import "C"

import (
	"gox/internal/platform"
	"unsafe"
)

// Backend implements platform.Backend for iOS using UIKit.
type Backend struct{}

// New creates a new iOS backend.
func New() *Backend {
	return &Backend{}
}

func (b *Backend) CreateView() platform.ViewHandle {
	return platform.ViewHandle(C.gox_create_view())
}

func (b *Backend) CreateText(content string) platform.ViewHandle {
	cs := C.CString(content)
	defer C.free(unsafe.Pointer(cs))
	return platform.ViewHandle(C.gox_create_label(cs))
}

func (b *Backend) AddChild(parent, child platform.ViewHandle) {
	C.gox_add_subview(C.long(parent), C.long(child))
}

func (b *Backend) SetRootView(handle platform.ViewHandle) {
	C.gox_set_root_view(C.long(handle))
}

func (b *Backend) SetBackgroundColor(handle platform.ViewHandle, color string) {
	cs := C.CString(color)
	defer C.free(unsafe.Pointer(cs))
	C.gox_set_background_color(C.long(handle), cs)
}

func (b *Backend) SetFrame(handle platform.ViewHandle, x, y, width, height float64) {
	C.gox_set_frame(C.long(handle), C.double(x), C.double(y), C.double(width), C.double(height))
}

func (b *Backend) SetPadding(handle platform.ViewHandle, top, right, bottom, left float64) {
	C.gox_set_padding(C.long(handle), C.double(top), C.double(right), C.double(bottom), C.double(left))
}

func (b *Backend) SetFontSize(handle platform.ViewHandle, size float64) {
	C.gox_set_font_size(C.long(handle), C.double(size))
}

func (b *Backend) SetFontWeight(handle platform.ViewHandle, weight string) {
	cs := C.CString(weight)
	defer C.free(unsafe.Pointer(cs))
	C.gox_set_font_weight(C.long(handle), cs)
}

func (b *Backend) SetTextColor(handle platform.ViewHandle, color string) {
	cs := C.CString(color)
	defer C.free(unsafe.Pointer(cs))
	C.gox_set_text_color(C.long(handle), cs)
}

func (b *Backend) SetTextAlign(handle platform.ViewHandle, align string) {
	cs := C.CString(align)
	defer C.free(unsafe.Pointer(cs))
	C.gox_set_text_align(C.long(handle), cs)
}

func (b *Backend) SetBorderRadius(handle platform.ViewHandle, radius float64) {
	C.gox_set_border_radius(C.long(handle), C.double(radius))
}

func (b *Backend) SetOpacity(handle platform.ViewHandle, opacity float64) {
	C.gox_set_opacity(C.long(handle), C.double(opacity))
}

func (b *Backend) SetFlexLayout(handle platform.ViewHandle, direction, justify, align string) {
	cd := C.CString(direction)
	cj := C.CString(justify)
	ca := C.CString(align)
	defer C.free(unsafe.Pointer(cd))
	defer C.free(unsafe.Pointer(cj))
	defer C.free(unsafe.Pointer(ca))
	C.gox_set_flex_layout(C.long(handle), cd, cj, ca)
}

func (b *Backend) RunApp() {
	C.gox_run_app(0, nil)
}
