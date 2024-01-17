package resolver

import (
	"fmt"

	"github.com/yuemori/blueprinter/internal/parser"
)

// Derivation is a type that represents a derivation of a resolver.
type Derivation interface {
	isDerivation()
}

var (
	_ Derivation = (*FieldDecl)(nil)
	_ Derivation = (*PrivateFuncDecl)(nil)
)

// A FieldDecl is a type that represents a field of a resolver.
type FieldDecl struct {
	Name string
	Type parser.Type
}

func (*FieldDecl) isDerivation() {}

// A FuncDecl is a type that represents a function of a resolver.
type FuncDecl interface {
	FuncName() string
	FuncBody() string
	Imports() []string
	Pkg() string
	FuncReturn() string

	isFuncDecl()
}

var (
	_ FuncDecl = (*PublicFuncDecl)(nil)
	_ FuncDecl = (*PrivateFuncDecl)(nil)
)

// A PublicFuncDecl is a type that represents a public function of a resolver.
type PublicFuncDecl struct {
	fn      *parser.Func
	library string
	params  []Derivation
}

func (p *PublicFuncDecl) FuncName() string {
	return "Resolve" + p.fn.Name()
}

func (p *PublicFuncDecl) Pkg() string {
	return p.fn.ImportPath()
}

func (p *PublicFuncDecl) FuncReturn() string {
	return parser.QualifiedTypeName(p.fn.Results().At(0).Type())
}

func (p *PublicFuncDecl) FuncBody() string {
	args := make([]interface{}, 0)
	format := "\treturn "
	if p.fn.ImportPath() != p.library {
		format += "%s."
		args = append(args, p.fn.FullPkg())
	}
	format += "%s(\n"
	args = append(args, p.fn.Name())
	for _, param := range p.params {
		switch p := param.(type) {
		case *FieldDecl:
			format += "\t\t// %s\n"
			args = append(args, parser.TypeNamePrefixedByImportPath(p.Type))
			format += "\t\tf.%s,\n"
			args = append(args, p.Name)
		case *PrivateFuncDecl:
			format += "\t\t// %s\n"
			args = append(args, parser.TypeNamePrefixedByImportPath(p.ReturnType()))
			format += "\t\tf.%s(),\n"
			args = append(args, p.FuncName())
		}
	}
	format += "\t)"

	return fmt.Sprintf(format, args...)
}

func (p *PublicFuncDecl) Imports() []string {
	return []string{p.fn.FullImportPath()}
}

func (*PublicFuncDecl) isFuncDecl() {}

// A PrivateFuncDecl is a type that represents a private function of a resolver.
type PrivateFuncDecl struct {
	iface   *parser.Iface
	fn      *parser.Func
	params  []Derivation
	library string
}

func (i *PrivateFuncDecl) FuncReturn() string {
	if i.iface.ImportPath() == i.library {
		return i.iface.Name()
	}
	return fmt.Sprintf("%s.%s", i.iface.FullPkg(), i.iface.Name())
}

func (i *PrivateFuncDecl) Pkg() string {
	return i.iface.ImportPath()
}

func (i *PrivateFuncDecl) Imports() []string {
	return []string{i.iface.FullImportPath(), i.fn.FullImportPath()}
}

func (i *PrivateFuncDecl) FuncName() string {
	return i.iface.FullPkg() + "_" + i.iface.Name()
}

func (i *PrivateFuncDecl) ReturnType() parser.Type {
	return i.iface.Type()
}

func (p *PrivateFuncDecl) FuncBody() string {
	args := make([]interface{}, 0)
	format := "\treturn "
	if p.fn.ImportPath() != p.library {
		format += "%s."
		args = append(args, p.fn.FullPkg())
	}
	format += "%s(\n"
	args = append(args, p.fn.Name())
	for _, param := range p.params {
		switch p := param.(type) {
		case *FieldDecl:
			format += "\t\t// %s\n"
			args = append(args, parser.TypeNamePrefixedByImportPath(p.Type))
			format += "\t\tf.%s,\n"
			args = append(args, p.Name)
		case *PrivateFuncDecl:
			format += "\t\t// %s\n"
			args = append(args, parser.TypeNamePrefixedByImportPath(p.ReturnType()))
			format += "\t\tf.%s(),\n"
			args = append(args, p.FuncName())
		}
	}
	format += "\t)"

	return fmt.Sprintf(format, args...)
}

func (*PrivateFuncDecl) isFuncDecl()   {}
func (*PrivateFuncDecl) isDerivation() {}
