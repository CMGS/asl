package a

type widget struct{ n int }

// newWidget is a constructor for a file-local type: exempt.
func newWidget() *widget { return &widget{} }

func (w *widget) grow() { w.n++ }

func helper() int { return 1 } // want `unexported function helper declared above exported function Second`

func First() int { return helper() }

func also() int { return 2 } // want `unexported function also declared above exported function Second`

func Second() int { return also() }

func trailing() int { return 3 }
