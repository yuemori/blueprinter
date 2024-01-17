package parser

import "go/types"

// Iface is a wrapper of types.Interface.
type Iface struct {
	*Object
}

// Interface returns the underlying types.Interface.
func (i *Iface) Interface() *types.Interface {
	return i.object.Type().Underlying().(*types.Interface)
}
