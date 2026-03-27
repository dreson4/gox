//go:build ios

package linking

/*
#include <stdlib.h>

extern int goxOpenURL(const char *url);
*/
import "C"
import (
	"errors"
	"unsafe"
)

func platformOpenURL(url string) error {
	cURL := C.CString(url)
	defer C.free(unsafe.Pointer(cURL))

	result := C.goxOpenURL(cURL)
	if result != 0 {
		return errors.New("linking: failed to open URL")
	}
	return nil
}
