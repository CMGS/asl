package a

import "testing"

func testHelper() int { return 1 }

func TestHelper(t *testing.T) {
	if testHelper() != 1 {
		t.Fatal("helper")
	}
}

func belowTests() int { return 0 }
