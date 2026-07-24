package a

type typed interface{ Type() string }

type wrapper[T typed] struct{ v T }

type inline[T interface{ Type() string }] struct{ v T } // want `inline interface constraint; declare a named constraint type`

func good[T typed](v T) string { return v.Type() }

func bad[T interface{ Type() string }](v T) string { return v.Type() } // want `inline interface constraint; declare a named constraint type`

func anyOK[T any](v T) T { return v }

func comparableOK[T comparable](v T) T { return v }
