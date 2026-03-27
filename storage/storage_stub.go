//go:build !ios && !android

package storage

import (
	"errors"
	"sync"
)

var (
	stubMu   sync.Mutex
	stubData = make(map[string]string)
)

func platformSet(key, value string) error {
	stubMu.Lock()
	defer stubMu.Unlock()
	stubData[key] = value
	return nil
}

func platformGet(key string) (string, error) {
	stubMu.Lock()
	defer stubMu.Unlock()
	v, ok := stubData[key]
	if !ok {
		return "", errors.New("storage: key not found")
	}
	return v, nil
}

func platformDelete(key string) error {
	stubMu.Lock()
	defer stubMu.Unlock()
	delete(stubData, key)
	return nil
}

// Reset clears all stored data. Used in tests.
func Reset() {
	stubMu.Lock()
	defer stubMu.Unlock()
	stubData = make(map[string]string)
}
