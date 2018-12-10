package coder

import (
	"github.com/dave/jennifer/jen"
	"github.com/reddec/symbols"
	"go/ast"
)

func generatePrefix(expr ast.Expr) jen.Code {
	if v, ok := expr.(*ast.StarExpr); ok {
		return jen.Op("*").Add(generatePrefix(v.X))
	}
	if v, ok := expr.(*ast.ArrayType); ok {
		return jen.Index().Add(generatePrefix(v.Elt))
	}
	return jen.Empty()
}

func generateType(tp *symbols.Symbol) jen.Code {
	if tp.BuiltIn {
		return jen.Id(tp.Name)
	}
	return jen.Qual(tp.Import.Import, tp.Name)
}

func GenerateStruct(sym *symbols.Symbol, resolver symbols.Resolver) (jen.Code, error) {
	fields, err := sym.Fields(resolver)
	if err != nil {
		return nil, err
	}
	return jen.Type().Id(sym.Name).StructFunc(func(st *jen.Group) {
		for _, field := range fields {
			st.Id(field.Name).Add(generatePrefix(field.RawType)).Add(generateType(field.Type)).Tag(field.Tags)
		}
	}), nil
}
