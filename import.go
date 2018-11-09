package symbols

import (
	"go/ast"
)

type Type struct {
	Location string
	Import   string
	Name     string
	Ast      ast.Expr
}

type Imports []Import

func (imps Imports) ByImport(packageImport string) *Import {
	for _, imp := range imps {
		if imp.Import == packageImport {
			return &imp
		}
	}
	return nil
}

func (imps Imports) ByPackageName(packageName string) *Import {
	for _, imp := range imps {
		if imp.Package == packageName {
			return &imp
		}
	}
	return nil
}

func (imps Imports) ByFile(f *File) Imports {
	var ans Imports
	var mp = make(map[string]bool)
	for _, n := range f.Imports() {
		mp[n] = true
	}
	for _, v := range imps {
		if mp[v.Import] {
			ans = append(ans, v)
		}
	}
	return ans
}

func (imp *Import) FindSymbol(name string) *Symbol {
	for _, f := range imp.Files {
		node := f.FindSymbol(name)
		if node != nil {
			return &Symbol{Import: imp, File: f, Node: node, Name: name}
		}
	}
	return nil
}

func (f *File) FindSymbol(targetName string) ast.Node {
	var stack []ast.Node
	for i := len(f.Ast.Decls) - 1; i >= 0; i-- {
		stack = append(stack, f.Ast.Decls[i])
	}

	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		switch v := node.(type) {
		case *ast.TypeSpec:
			if v.Name.Name == targetName {
				return v
			}
		case *ast.GenDecl:
			for _, spec := range v.Specs {
				stack = append(stack, spec)
			}
		case *ast.FuncDecl:
			if v.Name.Name == targetName {
				return v
			}
		case *ast.Ident:
			if v.Name == targetName {
				return v
			}
		case *ast.ValueSpec:
			for _, name := range v.Names {
				if name.Name == targetName {
					return name
				}
			}
		}
	}
	return nil
}

func (f *File) Symbols() []string {
	var stack []ast.Node
	for i := len(f.Ast.Decls) - 1; i >= 0; i-- {
		stack = append(stack, f.Ast.Decls[i])
	}
	var ans []string

	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		switch v := node.(type) {
		case *ast.GenDecl:
			for _, spec := range v.Specs {
				stack = append(stack, spec)
			}
		case *ast.TypeSpec:
			ans = append(ans, v.Name.Name)
		case *ast.FuncDecl:
			ans = append(ans, v.Name.Name)
		case *ast.Ident:
			ans = append(ans, v.Name)
		case *ast.ValueSpec:
			for _, name := range v.Names {
				ans = append(ans, name.Name)
			}
		}
	}
	return ans
}
