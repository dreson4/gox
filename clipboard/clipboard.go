// Package clipboard provides read/write access to the system clipboard.
// On iOS it uses UIPasteboard, on Android the ClipboardManager.
package clipboard

// Copy places text on the system clipboard.
func Copy(text string) {
	platformCopy(text)
}

// Read returns the current text content of the system clipboard.
// Returns an empty string if the clipboard is empty or contains non-text data.
func Read() string {
	return platformRead()
}
