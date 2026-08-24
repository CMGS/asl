package a

import "b"

type cni struct{}

func newCNI() b.Network { return &cni{} }

func (c *cni) Up() {}

func audit() { _ = &cni{} } // want `unexported function audit declared above exported function Exported; move unexported helpers below the exported set`

func Exported() int { return low() }

func low() int { return 0 }
