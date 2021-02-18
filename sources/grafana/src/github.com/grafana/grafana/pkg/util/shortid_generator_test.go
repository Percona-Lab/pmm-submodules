package util

import "testing"

func TestAllowedCharMatchesUidPattern(t *testing.T) {
	for _, c := range allowedChars {
		if !IsValidShortUID(string(c)) {
			t.Fatalf("charset for creating new shortids contains chars not present in uid pattern")
		}
	}
}
