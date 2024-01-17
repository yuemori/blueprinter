package parser

import "go/types"

type Struct struct {
	*Object
}

func (s *Struct) Type() *types.Struct {
	return s.object.Type().Underlying().(*types.Struct)
}
