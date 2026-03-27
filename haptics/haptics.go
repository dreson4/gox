// Package haptics provides tactile feedback on supported devices.
// On iOS it uses UIImpactFeedbackGenerator and UINotificationFeedbackGenerator.
package haptics

// ImpactStyle represents the weight of an impact haptic.
type ImpactStyle int

const (
	Light  ImpactStyle = iota
	Medium
	Heavy
)

func (s ImpactStyle) String() string {
	switch s {
	case Light:
		return "light"
	case Medium:
		return "medium"
	case Heavy:
		return "heavy"
	default:
		return "medium"
	}
}

// NotificationType represents the type of notification haptic.
type NotificationType int

const (
	Success NotificationType = iota
	Warning
	Error
)

func (n NotificationType) String() string {
	switch n {
	case Success:
		return "success"
	case Warning:
		return "warning"
	case Error:
		return "error"
	default:
		return "success"
	}
}

// Impact triggers an impact haptic with the given style.
func Impact(style ImpactStyle) {
	platformImpact(style)
}

// Notify triggers a notification haptic with the given type.
func Notify(typ NotificationType) {
	platformNotify(typ)
}

// Selection triggers a light selection haptic.
func Selection() {
	platformSelection()
}
