package symbols

import (
	"github.com/pkg/errors"
	"go/ast"
)

func ArrayItem(node ast.Node) ast.Node {
	v := node.(*ast.ArrayType)
	return v.Elt
}

func IsIdent(node ast.Node) bool {
	_, ok := node.(*ast.Ident)
	return ok
}

func IsFunction(node ast.Node) bool {
	_, ok := node.(*ast.FuncDecl)
	return ok
}

func IsArray(node ast.Node) bool {
	_, ok := node.(*ast.ArrayType)
	return ok
}

func IsPointer(node ast.Node) bool {
	_, ok := node.(*ast.StarExpr)
	return ok
}

func IsStruct(node ast.Node) bool {
	if !IsType(node) {
		return false
	}
	_, ok := (node.(*ast.TypeSpec)).Type.(*ast.StructType)
	return ok
}

func IsStructDefinition(node ast.Node) bool {
	if !IsType(node) {
		return false
	}
	_, ok := node.(*ast.StructType)
	return ok
}

func IsMap(node ast.Node) bool {
	_, ok := node.(*ast.MapType)
	return ok
}

func IsInterface(node ast.Node) bool {
	if !IsType(node) {
		return false
	}
	_, ok := (node.(*ast.TypeSpec)).Type.(*ast.InterfaceType)
	return ok
}

func IsVariable(node ast.Node) bool {
	v, ok := node.(*ast.Ident)
	if !ok {
		return false
	}
	if v.Obj.Kind != ast.Var {
		return false
	}
	return true
}

func IsConstant(node ast.Node) bool {
	v, ok := node.(*ast.Ident)
	if !ok {
		return false
	}
	if v.Obj.Kind != ast.Con {
		return false
	}
	return true
}

func IsCall(node ast.Node) bool {
	_, ok := node.(*ast.CallExpr)
	return ok
}

func IsLiteral(node ast.Node) bool {
	_, ok := node.(*ast.BasicLit)
	return ok
}

func Literal(node ast.Node) (string, error) {
	v, ok := node.(*ast.BasicLit)
	if !ok {
		return "", errors.Errorf("is not a literal")
	}
	return v.Value, nil
}

func IsType(node ast.Node) bool {
	_, ok := node.(*ast.TypeSpec)
	return ok
}
