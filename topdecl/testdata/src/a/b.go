package a

var first int

var second int // want `more than one top-level var block; merge into a single block`

var _ = first // want `more than one top-level var block; merge into a single block`

var _ any = (*checked)(nil)

type checked struct{}

func use() int { return first + second + late + lateVar }
