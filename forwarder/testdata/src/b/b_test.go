package b

import "testing"

func TestHelper(t *testing.T) {
	if helper() != 1 {
		t.Fatal("helper")
	}
}
