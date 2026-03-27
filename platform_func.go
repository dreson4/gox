package gox

// Platform returns the current platform: "ios" or "android".
func Platform() string {
	return currentPlatform
}

// DismissKeyboard hides the on-screen keyboard if visible.
func DismissKeyboard() {
	platformDismissKeyboard()
}

// SetStatusBar sets the status bar appearance.
// Valid styles: "light" (white text), "dark" (black text), "auto" (system default).
func SetStatusBar(style string) {
	platformSetStatusBar(style)
}
