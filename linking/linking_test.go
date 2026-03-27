package linking

import "testing"

func TestOpenURLDoesNotError(t *testing.T) {
	if err := OpenURL("https://example.com"); err != nil {
		t.Fatalf("OpenURL: %v", err)
	}
}
