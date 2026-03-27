//go:build !ios && !android

package linking

func platformOpenURL(_ string) error {
	return nil
}
