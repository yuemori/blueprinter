package parser

import (
	"fmt"
	"go/ast"
	"go/types"
	"regexp"
	"strings"
)

var (
	// Match `provider:must_resolve` comment
	mustResolveRegexp = regexp.MustCompile("provider:must_resolve")
	// Match `provider:include` comment
	includeRegexp = regexp.MustCompile("provider:include")
	// Match `provider:resolve` comment
	resolveRegexp = regexp.MustCompile("provider:resolve *")
	// Match `provider:exclude` comment
	excludeRegexp = regexp.MustCompile("provider:exclude")
)

// A Object is a wrapper of types.Object.
type Object struct {
	object  types.Object
	comment *ast.CommentGroup
}

func newObject(object types.Object, comment *ast.CommentGroup) *Object {
	return &Object{
		object:  object,
		comment: comment,
	}
}

// Name returns the name of the object.
func (o *Object) Name() string {
	return o.object.Name()
}

// Pkg returns the package name of the object.
func (o *Object) Pkg() string {
	return o.object.Pkg().Name()
}

// FullPkg returns a string like: 'github_com_owner_repo_pkg'
// This string generated from o.ImportPath() by replacing '.', '/', '-' to '_'.
func (o *Object) FullPkg() string {
	str := o.ImportPath()

	for _, rep := range []string{".", "/", "-"} {
		str = strings.Replace(str, rep, "_", -1)
	}

	return str
}

// ImportPath returns the import path of the object.
func (o *Object) ImportPath() string {
	return o.object.Pkg().Path()
}

// FullImportPath returns a string like: 'github_com_owner_repo_pkg "github.com/owner/repo/pkg"'
func (o *Object) FullImportPath() string {
	return fmt.Sprintf("%s \"%s\"", o.FullPkg(), o.ImportPath())
}

// Interface returns the underlying Interface.
func (o *Object) Interface() (*Iface, bool) {
	_, ok := o.object.Type().Underlying().(*types.Interface)
	if !ok {
		return nil, false
	}
	return &Iface{o}, true
}

// Func returns the underlying Func.
func (o *Object) Func() (*Func, bool) {
	_, ok := o.object.(*types.Func)
	if !ok {
		return nil, false
	}
	return &Func{o}, true
}

// Struct returns the underlying Struct.
func (o *Object) Struct() (*Struct, bool) {
	_, ok := o.object.Type().Underlying().(*types.Struct)
	if !ok {
		return nil, false
	}
	return &Struct{o}, true
}

// Type returns the underlying type.
func (o *Object) Type() types.Type {
	return o.object.Type()
}

// Exported returns true if the object is exported.
func (o *Object) Exported() bool {
	return o.object.Exported()
}

// String returns a string like: 'github_com_owner_repo_pkg.ObjectName'
func (o *Object) String() string {
	return fmt.Sprintf("%s.%s", o.ImportPath(), o.object.Name())
}

// Comment returns the comment of the object.
func (o *Object) Comment() string {
	return o.comment.Text()
}

// IsExcluded returns true if the object has `provider:exclude` comment.
func (o *Object) IsExcluded() bool {
	if o.hasComment(excludeRegexp) {
		return true
	}
	return false
}

// MustBeResolved returns true if the object has `provider:must_resolve` comment.
func (f *Func) MustBeResolved() bool {
	return f.hasComment(mustResolveRegexp)
}

func (o *Object) IsMarkedAsBindable() bool {
	return o.hasComment(includeRegexp)
}

func (o *Object) IsResolve() bool {
	if o.comment == nil {
		return false
	}

	for _, comment := range o.comment.List {
		if resolveRegexp.MatchString(comment.Text) {
			return true
		}
	}

	return false
}

// Resolve returns the package path and the function name to be resolved.
// The format of the comment must be `provider:resolve path/to/package FuncName` or `provider:resolve FuncName`.
// If the object has no `provider:resolve` comment, this function returns false.
func (o *Object) ResolvedPkgAndFuncName() (string, string, error) {
	if o.comment == nil {
		return "", "", nil
	}

	for _, comment := range o.comment.List {
		if resolveRegexp.MatchString(comment.Text) {
			path := strings.TrimSpace(strings.TrimPrefix(comment.Text, "// provider:resolve"))
			paths := strings.Split(path, " ")
			if len(paths) == 1 {
				return o.ImportPath(), paths[0], nil
			}
			if len(paths) == 2 {
				return paths[0], paths[1], nil
			}

			return "", "", fmt.Errorf("resolver comment format must be `provider:resolve path/to/package FuncName` or `provider:resolve FuncName`, but: %s", comment.Text)
		}
	}

	return "", "", nil
}

func (o *Object) hasComment(r *regexp.Regexp) bool {
	if o.comment == nil {
		return false
	}
	for _, comment := range o.comment.List {
		if r.MatchString(comment.Text) {
			return true
		}
	}
	return false
}
