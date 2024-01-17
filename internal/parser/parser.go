package parser

import (
	"context"
	"go/ast"
	"os"
	"path/filepath"
	"regexp"

	"github.com/yuemori/blueprinter/internal/logger"
	"golang.org/x/tools/go/packages"
)

// Parse is a wrapper of packages.Load.
func Parse(ctx context.Context, dir string, env, globs, ignores []string) (*ObjectCache, []error) {
	patterns, err := collectPackagePatterns(dir, globs, ignores)
	if err != nil {
		return nil, err
	}

	pkgs, errs := load(ctx, dir, env, patterns)
	if len(errs) != 0 {
		return nil, errs
	}

	return buildCache(pkgs), nil
}

func collectPackagePatterns(dir string, globs, ignores []string) ([]string, []error) {
	dirsFound := make(map[string]bool)

	logger.Debug("Globs:", globs)
	logger.Debug("Ignores:", ignores)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if _, ok := dirsFound[path]; ok {
			return nil
		}

		if len(ignores) != 0 {
			matched, err := matches(path, ignores)
			if err != nil {
				return err
			}

			if matched {
				logger.Debug("Skip(ignore):", path)

				return nil
			}
		}

		if len(globs) != 0 {
			matched, err := matches(path, globs)
			if err != nil {
				return err
			}

			if !matched {
				logger.Debug("Skip(not matched):", path)

				return nil
			}
		}

		// TODO: Skip source code files that are specified with the //go:build !skip_blueprinter directive
		if filepath.Ext(path) == ".go" {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return err
			}

			dir := filepath.Dir(absPath)

			if _, found := dirsFound[dir]; !found {
				logger.Debug("Go Directory Found:", dir)
				dirsFound[dir] = true
			}
		}

		return nil
	})
	if err != nil {
		return nil, []error{err}
	}

	patterns := make([]string, 0, len(dirsFound))
	for dir := range dirsFound {
		patterns = append(patterns, dir)
	}

	return patterns, nil
}

// see: https://github.com/google/wire/blob/523d8fbe880bb310a188d472bccc0cef939c45b8/internal/wire/parse.go#L352
func load(ctx context.Context, wd string, env, patterns []string) ([]*packages.Package, []error) {
	cfg := &packages.Config{
		Context:    ctx,
		Mode:       packages.LoadAllSyntax,
		Dir:        wd,
		Env:        env,
		BuildFlags: []string{"-tags", "skip_blueprinter"},
	}
	escaped := make([]string, len(patterns))
	for i := range patterns {
		escaped[i] = "pattern=" + patterns[i]
	}

	pkgs, err := packages.Load(cfg, escaped...)
	if err != nil {
		return nil, []error{err}
	}
	var errs []error
	for _, p := range pkgs {
		for _, e := range p.Errors {
			errs = append(errs, e)
		}
	}
	if len(errs) > 0 {
		return nil, errs
	}
	return pkgs, nil
}

func buildCache(pkgs []*packages.Package) *ObjectCache {
	cache := newObjectCache()

	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				switch g := decl.(type) {
				case *ast.FuncDecl:
					if !g.Name.IsExported() {
						continue
					}
					obj := pkg.TypesInfo.ObjectOf(g.Name)
					if obj == nil {
						continue
					}
					cache.Add(newObject(obj, g.Doc))
				case *ast.GenDecl:
					for _, spec := range g.Specs {
						t, ok := spec.(*ast.TypeSpec)
						if !ok {
							continue
						}
						// skip private
						if !t.Name.IsExported() {
							continue
						}
						obj := pkg.TypesInfo.ObjectOf(t.Name)
						if obj == nil {
							continue
						}
						cache.Add(newObject(obj, g.Doc))
					}
				}
			}
		}
	}

	return cache
}

func compilePatterns(patterns []string) ([]*regexp.Regexp, error) {
	var regexps []*regexp.Regexp
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		regexps = append(regexps, re)
	}
	return regexps, nil
}

func matches(path string, patterns []string) (bool, error) {
	for _, p := range patterns {
		match, err := filepath.Match(p, path)
		if err != nil {
			return false, err
		}

		if match {
			return true, nil
		}
	}
	return false, nil
}
