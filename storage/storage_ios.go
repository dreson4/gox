//go:build ios

package storage

/*
#include <stdlib.h>

extern int goxStorageSet(const char *key, const char *value);
extern const char* goxStorageGet(const char *key);
extern void goxStorageDelete(const char *key);
extern void GoxFreeString(const char *s);
*/
import "C"
import (
	"errors"
	"unsafe"
)

func platformSet(key, value string) error {
	cKey := C.CString(key)
	cVal := C.CString(value)
	defer C.free(unsafe.Pointer(cKey))
	defer C.free(unsafe.Pointer(cVal))

	result := C.goxStorageSet(cKey, cVal)
	if result != 0 {
		return errors.New("storage: failed to set value")
	}
	return nil
}

func platformGet(key string) (string, error) {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))

	cVal := C.goxStorageGet(cKey)
	if cVal == nil {
		return "", errors.New("storage: key not found")
	}
	defer C.GoxFreeString(cVal)
	return C.GoString(cVal), nil
}

func platformDelete(key string) error {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))

	C.goxStorageDelete(cKey)
	return nil
}
