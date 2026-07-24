package a

import "testing"

const limit = 1

type fixture struct{ n int } // want `helper type above test funcs; move it below the last test`

func (f fixture) value() int { return f.n } // want `helper value above test funcs; move it below the last test`

func newFixture() fixture { return fixture{n: limit} } // want `helper newFixture above test funcs; move it below the last test`

func TestFirst(t *testing.T) {
	if newFixture().value() != limit {
		t.Fatal("value")
	}
}

func stray() {} // want `helper stray above test funcs; move it below the last test`

func TestSecond(t *testing.T) { stray() }

type below struct{}

func (below) ok() {}

func helperBelow() below { return below{} }
