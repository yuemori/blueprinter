package resolver

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/yuemori/blueprinter/internal/parser"

	"github.com/pkg/errors"
)

type Data struct {
	PublicDecls  map[string][]*FuncData
	PrivateDecls map[string][]*FuncData
	Imports      []string
	Package      string
}

type FuncData struct {
	ImportPaths []string
	Pkg         string
	Receiver    string
	FuncName    string
	FuncReturn  string
	FuncImpl    string
}

func Resolve(cache *parser.ObjectCache, target, library string) (*Data, error, []error) {
	providerImpl, err := loadProviderImpl(cache, library, target)
	if err != nil {
		return nil, err, nil
	}

	resolver := NewResolver(providerImpl, cache, library)
	decls, errs := resolver.Resolve()
	if errs != nil {
		return nil, nil, errs
	}

	privates := make(map[string][]*FuncData)
	publics := make(map[string][]*FuncData)

	for _, decl := range decls {
		data := &FuncData{
			ImportPaths: decl.Imports(),
			Pkg:         decl.Pkg(),
			Receiver:    fmt.Sprintf("f *%s", target),
			FuncName:    decl.FuncName(),
			FuncReturn:  decl.FuncReturn(),
			FuncImpl:    decl.FuncBody(),
		}
		switch decl.(type) {
		case *PublicFuncDecl:
			if _, ok := publics[decl.Pkg()]; !ok {
				publics[decl.Pkg()] = make([]*FuncData, 0)
			}
			publics[decl.Pkg()] = append(publics[decl.Pkg()], data)
		case *PrivateFuncDecl:
			if _, ok := privates[decl.Pkg()]; !ok {
				privates[decl.Pkg()] = make([]*FuncData, 0)
			}
			privates[decl.Pkg()] = append(privates[decl.Pkg()], data)
		}
	}

	for _, slice := range publics {
		sort.SliceStable(slice, func(x, y int) bool {
			return slice[x].FuncName < slice[y].FuncName
		})
	}
	for _, slice := range privates {
		sort.SliceStable(slice, func(x, y int) bool {
			return slice[x].FuncName < slice[y].FuncName
		})
	}

	importMap := make(map[string]string, 0)
	imports := make([]string, 0)
	for _, binding := range decls {
		for _, path := range binding.Imports() {
			if path != "" {
				importMap[path] = path
			}
		}
	}
	for _, imp := range importMap {
		imports = append(imports, imp)
	}
	sort.SliceStable(imports, func(x, y int) bool {
		return imports[x] < imports[y]
	})

	path := strings.Split(library, "/")

	return &Data{
		Package:      path[len(path)-1],
		Imports:      imports,
		PrivateDecls: privates,
		PublicDecls:  publics,
	}, nil, nil
}

func loadProviderImpl(cache *parser.ObjectCache, library, name string) (*parser.Struct, error) {
	obj, ok := cache.Get(library, name)
	if !ok {
		return nil, errors.New(fmt.Sprintf("%s.%s does not exist.", library, name))
	}
	impl, ok := obj.Struct()
	if !ok {
		return nil, errors.New(fmt.Sprintf("%s.%s is not struct type", library, name))
	}
	return impl, nil
}

type Resolver struct {
	provider *parser.Struct

	fields []*FieldDecl
	decls  []*PrivateFuncDecl

	bindings map[*parser.Iface]*parser.Func

	cache   *parser.ObjectCache
	library string
}

func NewResolver(provider *parser.Struct, cache *parser.ObjectCache, library string) *Resolver {
	fields := make([]*FieldDecl, 0)

	for i := 0; i < provider.Type().NumFields(); i++ {
		f := provider.Type().Field(i)
		fields = append(fields, &FieldDecl{
			Name: f.Name(),
			Type: f.Type(),
		})
	}

	return &Resolver{
		provider: provider,
		fields:   fields,
		decls:    make([]*PrivateFuncDecl, 0),
		cache:    cache,
		library:  library,
	}
}

// Resolve feature searches for constructors within the ObjectCache that can resolve dependencies, and returns the resolved results as function declarations (FuncDecl).
// Here, a 'constructor' refers to a function that is not a method, has one or more arguments, and returns a single value.
//
// 'Resolution' means generating the necessary arguments for a constructor and making the function executable. 'Resolved result' refers to the declarations of the functions required to perform such processing.
//
// The resolution process occurs in the following steps:
//
// 1. For all interfaces in the ObjectCache, determine a single constructor that will be bound to each.
// 2. Seek derivations for these interfaces. This involves checking if constructors bound to interface types can be derived using other methods.
// 3. For all functions in the ObjectCache, determine their resolution results.
//
// Through this process, we can provide a simple and user-friendly interface with resolved dependencies.
func (r *Resolver) Resolve() ([]FuncDecl, []error) {
	resolved := make([]FuncDecl, 0)

	// Step 1: Build bindings for all interfaces in the ObjectCache.
	if errs := r.setupBindings(); len(errs) > 0 {
		return nil, errs
	}

	// Step 2: Derive constructors for all interfaces.
	r.deriveConstructorsForEachInterfaces()
	for _, decl := range r.decls {
		resolved = append(resolved, decl)
	}

	// Step 3: For all constructors, find their resolution results.
	// Note that the resolution results are the union of all methods found here and all derivations generated in Step 2.
	decls, errs := r.resolveEachConstructorsPresumingDerivationIsDone()
	if len(errs) > 0 {
		return nil, errs
	}
	for _, decl := range decls {
		resolved = append(resolved, decl)
	}

	return resolved, nil
}

func (r *Resolver) resolveEachConstructorsPresumingDerivationIsDone() ([]*PublicFuncDecl, []error) {
	resolved := make([]*PublicFuncDecl, 0)
	errs := make([]error, 0)
	for _, fn := range r.cache.Funcs() {
		if fn.ImportPath() == r.library {
			continue
		}
		if !fn.ShouldTryToResolve() {
			continue
		}
		params, err := r.findDerivationsForParams(fn)
		if err != nil {
			if fn.MustBeResolved() {
				errs = append(errs, fmt.Errorf(
					"unable to resolve %s.%s, which is marked as `must_resolve`: %s",
					fn.ImportPath(), fn.Name(), err.Error(),
				))
			}
			continue
		}

		decl := &PublicFuncDecl{
			fn:      fn,
			library: r.library,
			params:  params,
		}
		resolved = append(resolved, decl)
	}

	if len(errs) > 0 {
		return nil, errs
	}

	return resolved, nil
}

func (r *Resolver) deriveConstructorsForEachInterfaces() {
	derived := make(map[*parser.Iface]*PrivateFuncDecl)

	for loop := 0; ; loop++ {
		// If the number of loops exceeds 1000, it's likely that an infinite loop has occurred.
		if loop > 1000 {
			panic("Infinite loop found")
		}

		numDerived := 0
		for iface, fn := range r.bindings {
			// Check if the interface has already been derived. If so, skip it.
			if _, ok := derived[iface]; ok {
				continue
			}

			// Check if all arguments of the function are derived. If not, skip it.
			params, err := r.findDerivationsForParams(fn)
			if err != nil {
				continue
			}

			decl := &PrivateFuncDecl{
				library: r.library,
				fn:      fn,
				iface:   iface,
				params:  params,
			}
			numDerived += 1
			r.AddFunc(decl)
			derived[iface] = decl
		}

		if numDerived == 0 {
			return
		}
	}
}

func (w *Resolver) AddFunc(fn *PrivateFuncDecl) {
	w.decls = append(w.decls, fn)
}

// setupBindings は、 r の持つ ObjectCache 内のすべてのインターフェースに対する binding を構築します。
// bindings については、 Resolver 型内のコメントを参照してください。
func (r *Resolver) setupBindings() []error {
	r.bindings = make(map[*parser.Iface]*parser.Func, 0)
	errs := make([]error, 0)

	for _, iface := range r.cache.Ifaces() {
		// In the case of 'exclude', skip it.
		if iface.IsExcluded() {
			continue
		}
		// In the case of 'interface{}', skip it.
		if iface.Interface().Empty() {
			continue
		}
		// In the case of 'private', skip it.
		if !iface.Exported() {
			continue
		}

		// In the case where 'provider:resolve' is specified
		if iface.IsResolve() {
			pkg, name, err := iface.ResolvedPkgAndFuncName()
			if err != nil {
				errs = append(errs, err)
				continue
			}

			obj, ok := r.cache.Get(pkg, name)
			if !ok {
				errs = append(errs, errors.Errorf("%s does not found %s.%s", iface.Name(), pkg, name))
				continue
			}
			fn, ok := obj.Func()
			if !ok {
				errs = append(errs, errors.Errorf("%s.%s is not a function: %s", pkg, name, iface.Name()))
				continue
			}
			r.bindings[iface] = fn
		} else {
			typs := r.cache.Implementations(iface)
			if len(typs) == 0 {
				continue
			}

			// If multiple Types implementing 'iface' exist, it's necessary to either
			// narrow down the candidates with 'provider:resolve' or exclude unwanted
			// candidates with 'provider:exclude'. In this case, since neither applies,
			// it's not possible to uniquely identify the binding target, resulting in an error.
			if len(typs) != 1 {
				msg := fmt.Sprintf(
					"Unable to determine an implementation for %s.%s: "+
						"more than one parser implement this interface.\n"+
						"Use // provider:resolve to specify which constructor should be used, "+
						"or // provider:exclude if you want to ignore certain constructors for this type.\n\n"+
						"Possible implementations for this interface are:\n",
					iface.ImportPath(), iface.Name())
				for i, t := range typs {
					msg += fmt.Sprintf("\t%d: %s\n", i, parser.QualifiedTypeName(t))
				}
				errs = append(errs, errors.New(msg))
				continue
			}

			t := typs[0]

			fns := make([]*parser.Func, 0)
			for _, fn := range r.cache.Funcs() {
				if !fn.IsBindable() {
					continue
				}
				if parser.IsEmpty(t) {
					continue
				}
				if !parser.Identical(fn.Results().At(0).Type(), t) {
					continue
				}
				fns = append(fns, fn)
			}
			// skip if not found
			if len(fns) == 0 {
				log.Printf("skip(can not find resolver function): %v\n", typs[0])
				continue
			}

			// If multiple functions can be bound to 'iface', it's necessary to either
			// narrow down the candidates with 'provider:resolve' or exclude unwanted
			// candidates with 'provider:exclude'. In this case, since neither applies,
			// it's not possible to uniquely identify the binding target, resulting in an error.
			if len(fns) != 1 {
				msg := fmt.Sprintf(
					"Unable to determine constructors for %s.%s (which is resolved to %s): "+
						"More than one constructors are found for this interface\n"+
						"Use // provider:resolve to specify which constructor should be used, "+
						"or // provider:exclude if you want to ignore certain constructors for this type.\n\n"+
						"Possible constructors for this interface are:\n",
					iface.ImportPath(), iface.Name(), parser.QualifiedTypeName(t))
				for i, fn := range fns {
					msg += fmt.Sprintf("\t%d: %s\n", i, parser.QualifiedTypeName(fn.Type()))
				}
				errs = append(errs, errors.New(msg))
				continue
			}
			r.bindings[iface] = fns[0]
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}

func (r *Resolver) findDerivationsForParams(fn *parser.Func) ([]Derivation, error) {
	errs := make([]error, 0)
	params := make([]Derivation, 0)
	for i := 0; i < fn.Params().Len(); i++ {
		t := fn.Params().At(i).Type()
		f, err := r.findDerivation(t)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		params = append(params, f)
	}
	if len(errs) > 0 {
		errMsg := fmt.Sprintf("unable to derive parameters for the function %s.%s\n",
			fn.ImportPath(), fn.Name())
		for _, e := range errs {
			errMsg += fmt.Sprintf("\t%s\n", e.Error())
		}
		return nil, errors.New(errMsg)
	}
	return params, nil
}

func (w *Resolver) findDerivation(t parser.Type) (Derivation, error) {
	// unsupported to struct{}, interface{}
	if parser.IsEmpty(t) {
		return nil, fmt.Errorf("the type %s is empty", parser.TypeNamePrefixedByImportPath(t))
	}
	for _, f := range w.fields {
		// NOTE: If the field is an interface and multiple fields satisfy 't', it might pass an unexpected one.
		// Currently, no workaround comes to mind and it's not a concern for now, so let's set it aside for the moment.
		if parser.AssignableTo(f.Type, t) {
			return f, nil
		}
	}

	for _, fn := range w.decls {
		// Use Identical for functions since AssignableTo may return a different Type that satisfies the interface
		if parser.Identical(fn.ReturnType(), t) {
			return fn, nil
		}
	}

	return nil, fmt.Errorf("no derivations found for %s", parser.TypeNamePrefixedByImportPath(t))
}
