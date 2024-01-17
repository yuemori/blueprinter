package parser

import "go/types"

type Func struct {
	*Object
}

func (f *Func) Exported() bool {
	fn, _ := f.object.(*types.Func)
	return fn.Exported()
}

func (f *Func) signature() *types.Signature {
	sig, _ := f.object.Type().(*types.Signature)
	return sig
}

func (f *Func) Params() *types.Tuple {
	return f.signature().Params()
}

func (f *Func) Results() *types.Tuple {
	return f.signature().Results()
}

func (f *Func) IsConstructor() bool {
	return f.Results().Len() == 1 && f.Params().Len() > 0
}

func (f *Func) ShouldTryToResolve() bool {
	return f.Exported() && !f.IsExcluded() && f.IsConstructor()
}

func (f *Func) IsBindable() bool {
	if f.Results().Len() != 1 {
		return false
	}

	if f.Params().Len() != 0 {
		return true
	}

	if f.IsMarkedAsBindable() {
		return true
	}
	return false
}
