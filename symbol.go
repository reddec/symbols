package symbols

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	"go/ast"
	"reflect"
)

type Symbol struct {
	Import  *Import
	File    *File
	Node    ast.Node
	Name    string
	BuiltIn bool
}

func (sym *Symbol) String() string {
	if sym.BuiltIn {
		return sym.Name
	}
	var val interface{}
	if sym.IsLiteral() {
		val, _ = sym.Literal()
	}
	if val == nil {
		return fmt.Sprint(sym.Import.Package, "{", sym.Import.Import, "}", sym.Name)
	}
	return fmt.Sprint(sym.Import.Package, "{", sym.Import.Import, "}", sym.Name, "=", val)
}

func (sym *Symbol) IsFunction() bool {
	_, ok := sym.Node.(*ast.FuncDecl)
	return ok
}

func (sym *Symbol) IsArray() bool {
	_, ok := sym.Node.(*ast.ArrayType)
	return ok
}

func (sym *Symbol) IsPointer() bool {
	_, ok := sym.Node.(*ast.StarExpr)
	return ok
}

func (sym *Symbol) IsStruct() bool {
	if !sym.IsType() {
		return false
	}
	_, ok := (sym.Node.(*ast.TypeSpec)).Type.(*ast.StructType)
	return ok
}

func (sym *Symbol) IsVariable() bool {
	v, ok := sym.Node.(*ast.Ident)
	if !ok {
		return false
	}
	if v.Obj.Kind != ast.Var {
		return false
	}
	return true
}

func (sym *Symbol) IsConstant() bool {
	v, ok := sym.Node.(*ast.Ident)
	if !ok {
		return false
	}
	if v.Obj.Kind != ast.Con {
		return false
	}
	return true
}

func (sym *Symbol) IsCall() bool {
	_, ok := sym.Node.(*ast.CallExpr)
	return ok
}

func (sym *Symbol) IsLiteral() bool {
	_, ok := sym.Node.(*ast.BasicLit)
	return ok
}

func (sym *Symbol) Literal() (string, error) {
	v, ok := sym.Node.(*ast.BasicLit)
	if !ok {
		return "", errors.Errorf("is not a literal")
	}
	return v.Value, nil
}

func (sym *Symbol) VarType() (*Symbol, error) {
	v, ok := sym.Node.(*ast.Ident)
	if !ok {
		return nil, errors.New("is not a var")
	}
	switch v.Obj.Kind {
	case ast.Var:
		switch o := v.Obj.Decl.(type) {
		case *ast.ValueSpec:
			spew.Dump(o)
			return &Symbol{
				Node:   o.Values[0],
				Name:   o.Names[0].Name,
				File:   sym.File,
				Import: sym.Import,
			}, nil
		default:
			return nil, errors.Errorf("unknown var type %v for symbol %v", unref(reflect.ValueOf(o).Type()).Name(), sym.Name)
		}
	}
	return nil, errors.Errorf("unknown var kind %v", v.Obj.Kind)
}

func unref(v reflect.Type) reflect.Type {
	if v.Kind() == reflect.Ptr {
		return unref(v.Elem())
	}
	return v
}

func (sym *Symbol) IsType() bool {
	_, ok := sym.Node.(*ast.TypeSpec)
	return ok
}

type Field struct {
	Name string
	Type *Symbol
}

func (sym *Symbol) Fields(resolver Resolver) ([]*Field, error) {
	st, ok := (sym.Node.(*ast.TypeSpec)).Type.(*ast.StructType)
	if !ok {
		return nil, errors.New("is not a struct")
	}
	var ans []*Field
	for _, p := range st.Fields.List {
		if len(p.Names) == 1 {
			sm, err := resolver.FindSymbol(realTypeQN(p.Type), sym.File)
			if err != nil {
				return nil, errors.Wrapf(err, "get real type of %v", p.Names[0])
			}
			ans = append(ans, &Field{
				Name: p.Names[0].Name,
				Type: sm,
			})
		}
	}
	return ans, nil
}

func realTypeQN(t ast.Node) string {
	if v, ok := t.(*ast.StarExpr); ok {
		return realTypeQN(v.X)
	}
	if v, ok := t.(*ast.ArrayType); ok {
		return realTypeQN(v.Elt)
	}
	if v, ok := t.(*ast.SelectorExpr); ok {
		return realTypeQN(v.X) + "." + v.Sel.Name
	}
	v := t.(*ast.Ident)
	return v.Name
}
