package parser

import (
	"go/types"
	"strings"
)

type Type types.Type

type ObjectCache struct {
	objects []*Object
}

func Identical(t1, t2 Type) bool {
	return types.Identical(t1, t2)
}

func AssignableTo(t1, t2 Type) bool {
	return types.AssignableTo(t1, t2)
}

func newObjectCache() *ObjectCache {
	return &ObjectCache{
		objects: make([]*Object, 0),
	}
}

func (c *ObjectCache) Get(pkg, name string) (*Object, bool) {
	for _, obj := range c.objects {
		if obj.Name() == name && obj.ImportPath() == pkg {
			return obj, true
		}
	}
	return nil, false
}

func IsEmpty(typ Type) bool {
	switch t := typ.(type) {
	case *types.Interface:
		return t.Empty()
	case *types.Named:
		return IsEmpty(t.Underlying())
	default:
		return false
	}
}

// TypeNamePrefixedByImportPath returns a string like: 'github_com_owner_repo_pkg.TypeName'
func TypeNamePrefixedByImportPath(t Type) string {
	return types.TypeString(t, nil)
}

// QualifiedTypeName returns a string like: 'github_com_owner_repo_pkg.TypeName'
// see: https://go.dev/ref/spec#Qualified_identifiers
func QualifiedTypeName(t Type) string {
	return types.TypeString(t, func(pkg *types.Package) string {
		str := pkg.Path()
		for _, rep := range []string{".", "/", "-"} {
			str = strings.Replace(str, rep, "_", -1)
		}
		return str
	})
}

func (c *ObjectCache) Add(obj *Object) {
	c.objects = append(c.objects, obj)
}

func (c *ObjectCache) All() []*Object {
	return c.objects
}

func (c *ObjectCache) Implementations(iface *Iface) []Type {
	binds := make([]Type, 0)
	for _, obj := range c.objects {
		if _, ok := obj.Struct(); ok {
			if types.Implements(obj.Type(), iface.Interface()) {
				binds = append(binds, obj.Type())
				continue
			}
			ptr := types.NewPointer(obj.Type())
			if types.Implements(ptr, iface.Interface()) {
				binds = append(binds, ptr)
			}
		}
	}
	return binds
}

func (c *ObjectCache) Ifaces() []*Iface {
	ifaces := make([]*Iface, 0)
	for _, obj := range c.objects {
		if it, ok := obj.Interface(); ok {
			ifaces = append(ifaces, it)
		}
	}
	return ifaces
}

func (c *ObjectCache) Funcs() []*Func {
	funcs := make([]*Func, 0)
	for _, obj := range c.objects {
		if fn, ok := obj.Func(); ok {
			funcs = append(funcs, fn)
		}
	}
	return funcs
}
