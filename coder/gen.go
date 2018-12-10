package coder

import (
	"github.com/dave/jennifer/jen"
	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
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

func GenerateStructMapper(source, target *symbols.Symbol, resolver symbols.Resolver, funcName string, ref bool) (jen.Code, error) {
	exists, tFields, unknownField, err := prepareMapStruct(source, target, resolver)
	if err != nil {
		return nil, err
	}
	var (
		srcName    = "src" + source.Name
		targetName = "dest" + target.Name
	)

	var mod = jen.Empty()
	if ref {
		mod = jen.Op("*")
	}

	return jen.Func().Id(funcName).ParamsFunc(func(params *jen.Group) {
		params.Id(srcName).Add(mod).Add(generateType(source))
		for _, f := range unknownField {
			params.Id(strcase.ToLowerCamel(f.Name)).Add(generateType(f.Type))
		}
	}).Add(mod).Add(generateType(target)).Add(mapStruct(targetName, srcName, exists, tFields, target, ref)), nil
}
func GenerateSelfStructMapper(source, target *symbols.Symbol, resolver symbols.Resolver, funcName string, ref bool) (jen.Code, error) {
	exists, tFields, unknownField, err := prepareMapStruct(source, target, resolver)
	if err != nil {
		return nil, err
	}
	var (
		srcName    = "src" + source.Name
		targetName = "dest" + target.Name
	)

	var mod = jen.Empty()
	if ref {
		mod = jen.Op("*")
	}

	return jen.Func().Parens(jen.Id(srcName).Add(mod).Add(generateType(source))).Id(funcName).ParamsFunc(func(params *jen.Group) {
		for _, f := range unknownField {
			params.Id(strcase.ToLowerCamel(f.Name)).Add(generateType(f.Type))
		}
	}).Add(mod).Add(generateType(target)).Add(mapStruct(targetName, srcName, exists, tFields, target, ref)), nil
}

func prepareMapStruct(source, target *symbols.Symbol, resolver symbols.Resolver) (map[string]*symbols.Field, []*symbols.Field, []*symbols.Field, error) {
	sFields, err := source.Fields(resolver)
	if err != nil {
		return nil, nil, nil, err
	}
	tFields, err := target.Fields(resolver)
	if err != nil {
		return nil, nil, nil, err
	}
	var exists = map[string]*symbols.Field{}
	for _, f := range sFields {
		exists[f.Name] = f
	}
	var unknownField []*symbols.Field
	for _, f := range tFields {
		if sf, ok := exists[f.Name]; !ok {
			unknownField = append(unknownField, f)
		} else if !sf.Type.Equal(f.Type) {
			return nil, nil, nil, errors.Errorf("field %v has different type in source and target struct", f.Name)
		}
	}
	return exists, tFields, unknownField, nil
}

func mapStruct(targetName string, srcName string, exists map[string]*symbols.Field, tFields []*symbols.Field, target *symbols.Symbol, ref bool) jen.Code {
	return jen.BlockFunc(func(group *jen.Group) {
		group.Var().Id(targetName).Add(generateType(target))
		for _, f := range tFields {
			if exists[f.Name] != nil {
				group.Id(targetName).Dot(f.Name).Op("=").Id(srcName).Dot(f.Name)
			} else {
				group.Id(targetName).Dot(f.Name).Op("=").Id(strcase.ToLowerCamel(f.Name))
			}
		}
		if ref {
			group.Return(jen.Op("&").Id(targetName))
		} else {
			group.Return(jen.Id(targetName))
		}
	})

}
