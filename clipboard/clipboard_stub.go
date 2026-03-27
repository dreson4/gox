//go:build !ios && !android

package clipboard

import "sync"

var (
	stubMu   sync.Mutex
	stubText string
)

func platformCopy(text string) {
	stubMu.Lock()
	defer stubMu.Unlock()
	stubText = text
}

func platformRead() string {
	stubMu.Lock()
	defer stubMu.Unlock()
	return stubText
}

// Reset clears the stub clipboard. Used in tests.
func Reset() {
	stubMu.Lock()
	defer stubMu.Unlock()
	stubText = ""
}
