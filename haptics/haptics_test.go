package haptics

import "testing"

func TestImpactDoesNotPanic(t *testing.T) {
	Impact(Light)
	Impact(Medium)
	Impact(Heavy)
}

func TestNotifyDoesNotPanic(t *testing.T) {
	Notify(Success)
	Notify(Warning)
	Notify(Error)
}

func TestSelectionDoesNotPanic(t *testing.T) {
	Selection()
}

func TestImpactStyleString(t *testing.T) {
	tests := []struct {
		style ImpactStyle
		want  string
	}{
		{Light, "light"},
		{Medium, "medium"},
		{Heavy, "heavy"},
	}
	for _, tt := range tests {
		if got := tt.style.String(); got != tt.want {
			t.Errorf("ImpactStyle(%d).String() = %q, want %q", tt.style, got, tt.want)
		}
	}
}

func TestNotificationTypeString(t *testing.T) {
	tests := []struct {
		typ  NotificationType
		want string
	}{
		{Success, "success"},
		{Warning, "warning"},
		{Error, "error"},
	}
	for _, tt := range tests {
		if got := tt.typ.String(); got != tt.want {
			t.Errorf("NotificationType(%d).String() = %q, want %q", tt.typ, got, tt.want)
		}
	}
}
