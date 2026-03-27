//go:build !ios && !android

package gox

const currentPlatform = "stub"

func platformDismissKeyboard() {}
func platformSetStatusBar(_ string) {}
