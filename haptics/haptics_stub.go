//go:build !ios && !android

package haptics

func platformImpact(_ ImpactStyle)      {}
func platformNotify(_ NotificationType) {}
func platformSelection()                {}
