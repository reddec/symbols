package symbols

import (
	"github.com/pkg/errors"
	"strings"
)

type Resolver interface {
	FindSymbol(qualifiedName string, sourceFile *File) (*Symbol, error)
}

type Project struct {
	Imports Imports
	Package *Import
}

func ProjectByPackage(packageImport string) (*Project, error) {
	imps, err := ScanPackage(packageImport)
	if err != nil {
		return nil, err
	}
	imp := imps.ByImport(packageImport)
	if imp == nil {
		panic("import scanned but not found")
	}
	return &Project{
		Imports: imps,
		Package: imp,
	}, nil
}

func (prj *Project) FindPackageImport(packageNameOrAlias string, file *File) (*Import, error) {
	// find by alias
	imps := prj.Imports.ByFile(file)
	for _, imp := range file.Ast.Imports {
		alias := ""
		if imp.Name != nil {
			alias = imp.Name.Name
		}
		if alias == packageNameOrAlias {
			return imps.ByImport(imp.Path.Value), nil
		}
	}
	// find by package
	imp := prj.Imports.ByFile(file).ByPackageName(packageNameOrAlias)
	if imp != nil {
		return imp, nil
	}
	return nil, errors.Errorf("failed to resolve import by package or alias %v", packageNameOrAlias)
}

func (prj *Project) FindSymbol(qualifiedName string, sourceFile *File) (*Symbol, error) {
	// unref
	if builtinTypes[qualifiedName] {
		return &Symbol{Name: qualifiedName, File: sourceFile, Node: nil, Import: nil, BuiltIn: true}, nil
	}
	qualifiedName = strings.Replace(qualifiedName, "*", "", -1)
	parts := strings.Split(qualifiedName, ".")
	var lookupImport *Import
	if len(parts) == 1 {
		// in current package
		lookupImport = prj.Package
	} else {
		imp, err := prj.FindPackageImport(parts[0], sourceFile)
		if err != nil {
			return nil, err
		}
		lookupImport = imp
	}
	name := parts[len(parts)-1]
	sym := lookupImport.FindSymbol(name)
	if sym == nil {
		return nil, errors.Errorf("symbol %v not found in %v", name, lookupImport.Import)
	}
	return sym, nil
}

func (prj *Project) Names() []string {
	var ans []string
	for _, v := range prj.Package.Files {
		ans = append(ans, v.Symbols()...)
	}
	return ans
}
