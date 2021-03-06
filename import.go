package symbols

import (
	"go/ast"
	"path/filepath"
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
			return &Symbol{Import: imp, File: f, Node: node.Raw, ParentNode: node.Parent, Name: name}
		}
	}
	return nil
}

func (imp *Import) FindFile(name string) *File {
	for _, f := range imp.Files {
		if filepath.Base(f.Filename) == name {
			return f
		}
	}
	return nil
}

func (imp *Import) Symbols(walk func(sym *Symbol) error) error {
	for _, f := range imp.Files {
		err := f.Symbols(func(node ast.Node, f *File, name string) error {
			return walk(&Symbol{Import: imp, File: f, Node: node, Name: name})
		})
		if err != nil {
			return err
		}
	}
	return nil
}

type Node struct {
	Raw    ast.Node
	Parent ast.Node
}

func (f *File) FindSymbol(targetName string) *Node {
	var stack []Node
	for i := len(f.Ast.Decls) - 1; i >= 0; i-- {
		stack = append(stack, Node{Raw: f.Ast.Decls[i], Parent: f.Ast})
	}

	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		switch v := node.Raw.(type) {
		case *ast.TypeSpec:
			if v.Name.Name == targetName {
				return &node
			}
		case *ast.GenDecl:
			for _, spec := range v.Specs {
				stack = append(stack, Node{Raw: spec, Parent: node.Raw})
			}
		case *ast.FuncDecl:
			if v.Name.Name == targetName {
				return &node
			}
		case *ast.Ident:
			if v.Name == targetName {
				return &node
			}
		case *ast.ValueSpec:
			for _, name := range v.Names {
				if name.Name == targetName {
					return &Node{Raw: name, Parent: v}
				}
			}
		}
	}
	return nil
}

func (f *File) Symbols(walk func(node ast.Node, f *File, name string) error) error {
	var stack []ast.Node
	for i := len(f.Ast.Decls) - 1; i >= 0; i-- {
		stack = append(stack, f.Ast.Decls[i])
	}
	var err error
	for len(stack) > 0 && err == nil {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		switch v := node.(type) {
		case *ast.TypeSpec:
			err = walk(v, f, v.Name.Name)
		case *ast.GenDecl:
			for _, spec := range v.Specs {
				stack = append(stack, spec)
			}
		case *ast.FuncDecl:
			err = walk(v, f, v.Name.Name)
		case *ast.Ident:
			err = walk(v, f, v.Name)
		case *ast.ValueSpec:
			for _, name := range v.Names {
				err = walk(v, f, name.Name)
			}

		}
	}
	return err
}

func (f *File) SymbolsNames() []string {
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
