package a

import "context"

type registry struct {
	build  func(ctx context.Context, name string, replicas int) (string, error) // want `func type func\(context.Context, string, int\) \(string, error\) spelled 3 times; declare a named type`
	kill   func(ctx context.Context, vmID string, force bool) error // want `func type func\(context.Context, string, bool\) error spelled 2 times; declare a named type`
	launch func(ctx context.Context, cold string, warm bool) error
	short  func() error
}

type mirror struct {
	build func(context.Context, string, int) (string, error) // want `func type func\(context.Context, string, int\) \(string, error\) spelled 3 times; declare a named type`
	kill  func(ctx context.Context, vmID string, force bool) error // want `func type func\(context.Context, string, bool\) error spelled 2 times; declare a named type`
}

func defaultBuild() func(context.Context, string, int) (string, error) { // want `func type func\(context.Context, string, int\) \(string, error\) spelled 3 times; declare a named type`
	return nil
}

func alone() func(ctx context.Context, other string) error { return nil }
