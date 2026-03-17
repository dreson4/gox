package gox

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation
#include <stdlib.h>

void GoxNSLog(const char *msg);
*/
import "C"

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"unsafe"
)

// nsLog writes a message through NSLog so it appears in iOS system logs.
func nsLog(msg string) {
	cs := C.CString(msg)
	C.GoxNSLog(cs)
	C.free(unsafe.Pointer(cs))
}

// NSLogHandler is a slog.Handler that routes all Go log output through NSLog,
// making it visible in iOS simulator logs and Xcode console.
type NSLogHandler struct {
	level slog.Level
	attrs []slog.Attr
	group string
}

func (h *NSLogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *NSLogHandler) Handle(_ context.Context, r slog.Record) error {
	var b strings.Builder
	b.WriteString(r.Level.String())
	b.WriteString(" ")
	b.WriteString(r.Message)

	// Add pre-set attrs
	for _, a := range h.attrs {
		fmt.Fprintf(&b, " %s=%v", a.Key, a.Value)
	}

	// Add record attrs
	r.Attrs(func(a slog.Attr) bool {
		key := a.Key
		if h.group != "" {
			key = h.group + "." + key
		}
		fmt.Fprintf(&b, " %s=%v", key, a.Value)
		return true
	})

	nsLog(b.String())
	return nil
}

func (h *NSLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &NSLogHandler{
		level: h.level,
		attrs: append(h.attrs, attrs...),
		group: h.group,
	}
}

func (h *NSLogHandler) WithGroup(name string) slog.Handler {
	g := name
	if h.group != "" {
		g = h.group + "." + name
	}
	return &NSLogHandler{
		level: h.level,
		attrs: h.attrs,
		group: g,
	}
}

func init() {
	slog.SetDefault(slog.New(&NSLogHandler{level: slog.LevelDebug}))
}
