package a

const limit = 1

var count int

func work() int { return count + limit }

const late = 2 // want `const declaration below the first func; move it into the top block`

var lateVar int // want `var declaration below the first func; move it into the top block`

var _ = work
