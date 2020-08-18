package common

import "testing"

func AssertTrue(t *testing.T, a bool) {
	t.Helper()
	if !a {
		t.Errorf("Not True %t", a)
	}
}
