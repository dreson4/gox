//go:build ios

package clipboard

/*
#include <stdlib.h>

extern void goxClipboardCopy(const char *text);
extern const char* goxClipboardRead(void);
extern void GoxFreeString(const char *s);
*/
import "C"
import "unsafe"

func platformCopy(text string) {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	C.goxClipboardCopy(cText)
}

func platformRead() string {
	cText := C.goxClipboardRead()
	if cText == nil {
		return ""
	}
	defer C.GoxFreeString(cText)
	return C.GoString(cText)
}
