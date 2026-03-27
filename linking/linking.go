// Package linking provides URL opening and deep link handling.
// On iOS it uses UIApplication openURL, on Android Intent.ACTION_VIEW.
package linking

// OpenURL opens the given URL in the default handler (browser, app, etc.).
func OpenURL(url string) error {
	return platformOpenURL(url)
}
