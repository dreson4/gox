package gox

import "sync/atomic"

var perfEnabled atomic.Bool

// EnablePerfMonitor turns on the performance overlay.
func EnablePerfMonitor() { perfEnabled.Store(true) }

// DisablePerfMonitor turns off the performance overlay.
func DisablePerfMonitor() { perfEnabled.Store(false) }

// PerfEnabled returns whether the perf monitor is active.
func PerfEnabled() bool { return perfEnabled.Load() }

// PerfData holds timing from one Go render cycle (nanoseconds).
type PerfData struct {
	RenderNs   int64 `json:"renderNs"`
	LayoutNs   int64 `json:"layoutNs"`
	MarshalNs  int64 `json:"marshalNs"`
	FrameCount int   `json:"frameCount"`
}
