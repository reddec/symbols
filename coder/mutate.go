// Mutate structs and interfaces
package coder

import (
	"github.com/pkg/errors"
	"github.com/reddec/symbols"
	"go/ast"
	"golang.org/x/tools/go/ast/astutil"
)

// Mutate struct
func MutateStruct(symStruct *symbols.Symbol, excludeFields []string) (*symbols.Symbol, error) {
	ok := symStruct.IsStruct()
	if !ok {
		return nil, errors.New("is not struct")
	}
	excluded := toSet(excludeFields)
	oldRoot := symStruct.Node.(*ast.TypeSpec)
	st := oldRoot.Type.(*ast.StructType)

	cp := make([]*ast.Field, len(st.Fields.List))
	copy(cp, st.Fields.List)

	newRoot := *oldRoot
	newTpSpec := *st
	newRoot.Type = &newTpSpec

	newTpSpec.Fields = &ast.FieldList{List: cp}

	node := astutil.Apply(&newRoot, func(cursor *astutil.Cursor) bool {
		f, ok := cursor.Node().(*ast.Field)
		if !ok {
			return true
		}
		if len(f.Names) == 1 {
			name := f.Names[0].Name
			if excluded[name] {
				cursor.Delete()
			}
		}
		return true
	}, func(cursor *astutil.Cursor) bool {
		return true
	})
	return symStruct.WithNode(node), nil
}

func toSet(opt []string) map[string]bool {
	ans := make(map[string]bool)
	for _, o := range opt {
		ans[o] = true
	}
	return ans
}
