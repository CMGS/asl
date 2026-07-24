package a

var first int

var second int // want `more than one top-level var block; merge into a single block`

func use() int { return first + second + late + lateVar }
