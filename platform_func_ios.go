//go:build ios

package gox

/*
#include <stdlib.h>
extern void goxDismissKeyboard(void);
extern void goxSetStatusBar(const char *style);
*/
import "C"
import "unsafe"

const currentPlatform = "ios"

func platformDismissKeyboard() {
	C.goxDismissKeyboard()
}

func platformSetStatusBar(style string) {
	cStyle := C.CString(style)
	defer C.free(unsafe.Pointer(cStyle))
	C.goxSetStatusBar(cStyle)
}
