package a

var ordered int

const after = 3 // want `const block below the var block; declare const first`

func useOrdered() int { return ordered + after }
