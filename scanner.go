package symbols

import (
	"fmt"
	"github.com/pkg/errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

type File struct {
	Filename string
	Import   string
	Ast      *ast.File
}

func (f *File) Imports() []string {
	var imports []string
	for _, imp := range f.Ast.Imports {
		importName, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			panic(err)
		}
		imports = append(imports, importName)
	}
	return imports
}

type Import struct {
	Import    string
	Package   string
	Directory string
	Files     []*File
}

func Scan(dir string) (Imports, error) {
	imps, _, err := selfScan(dir)
	return imps, err
}

func selfScan(dir string) (Imports, string, error) {
	var path = lookupFolders(dir)
	selfPackage, err := findPackageByDir(dir, path...)
	if err != nil {
		return nil, "", err
	}

	imps, err := scanPackageWithLookups(selfPackage, path)
	return imps, selfPackage, err
}

func ScanPackage(importPath string) (Imports, error) {
	return scanPackageWithLookups(importPath, lookupFolders("."))

}

func scanPackageWithLookups(importName string, path []string) (Imports, error) {
	var imports = make(map[string]Import)
	var packagesToScan = []string{importName}

	for len(packagesToScan) > 0 {
		toScan := packagesToScan[len(packagesToScan)-1]
		packagesToScan = packagesToScan[:len(packagesToScan)-1]
		if _, scanned := imports[toScan]; scanned {
			continue
		}
		imp, importSet, err := scanImport(toScan, path...)
		if err != nil {
			_, isN := err.(*ImportNotFoundErr)
			if isN {
				continue // just ignore broken import (mostly from GOROOT/src)
			}
			return nil, err
		}
		imports[imp.Import] = imp
		for _, file := range imp.Files {
			file.Import = imp.Import
		}
		for _, importPath := range importSet {
			if importPath == "C" { // special import
				continue
			}

			if _, ok := imports[importPath]; !ok {
				packagesToScan = append(packagesToScan, importPath)
			}
		}
	}

	// map result
	var imps []Import
	for _, imp := range imports {
		imps = append(imps, imp)
	}
	return imps, nil
}

func lookupFolders(root string) []string {
	var path []string
	goRoot := filepath.Join(runtime.GOROOT(), "src")
	goPath := filepath.Join(os.Getenv("GOPATH"), "src")
	if root != "" {
		vendorDir := findVendorDir(root)
		if vendorDir != "" {
			path = append(path, vendorDir)
		}
	}
	return append(path, goPath, goRoot)
}

// non-recursive import scan in path..
func scanImport(importPath string, pathes ...string) (Import, []string, error) {
	for _, path := range pathes {
		location := filepath.Join(path, strings.Replace(importPath, "/", string(filepath.Separator), -1))
		imp, imports, err := scanDirectory(location, importPath)
		if err != nil {
			continue
		}
		return imp, imports, err
	}
	e := ImportNotFoundErr(fmt.Sprintf("import %v not found in %v", importPath, strings.Join(pathes, ", ")))
	return Import{}, nil, &e
}

type ImportNotFoundErr string

func (in *ImportNotFoundErr) Error() string { return string(*in) }

func scanDirectory(directory, assumingImportName string) (Import, []string, error) {
	var imp Import
	var importSet = make(map[string]struct{})
	imp.Import = assumingImportName
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return imp, nil, err
	}
	var ok bool
	for _, fileStat := range files {
		if fileStat.IsDir() || filepath.Ext(fileStat.Name()) != ".go" {
			continue
		}
		ok = true

		imp.Directory = directory
		fileName := filepath.Join(directory, fileStat.Name())
		imports, f, err := scanFile(fileName)
		if imp.Package == "" || strings.HasSuffix(imp.Package, "_test") {
			imp.Package = f.Ast.Name.Name
		}
		if err != nil {
			return imp, nil, errors.Wrapf(err, "scan file %v for import %v", fileName, assumingImportName)
		}
		for _, impPath := range imports {
			importSet[impPath] = struct{}{}
		}
		imp.Files = append(imp.Files, f)
	}
	if !ok {
		return imp, nil, errors.Errorf("no source files in %v", directory)
	}
	var allImports []string
	for impPath := range importSet {
		allImports = append(allImports, impPath)
	}
	return imp, allImports, nil
}

func scanFile(filename string) ([]string, *File, error) {
	tokens := token.NewFileSet()
	file, err := parser.ParseFile(tokens, filename, nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}
	var imports []string
	for _, imp := range file.Imports {
		importName, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			panic(err)
		}
		imports = append(imports, importName)
	}

	return imports, &File{Ast: file, Filename: filename}, nil
}

func findPackageByDir(fileName string, lookups ...string) (string, error) {
	abs, err := filepath.Abs(fileName)
	if err != nil {
		return "", err
	}
	for _, lookup := range lookups {
		aLookup, err := filepath.Abs(lookup)
		if err != nil {
			return "", err
		}
		if pkg, err := filepath.Rel(aLookup, abs); err == nil && !strings.Contains(pkg, "..") {
			return strings.Replace(pkg, string(filepath.Separator), "/", -1), nil
		}
	}
	return "", errors.Errorf("failed to detect package for file %v against %v", fileName, strings.Join(lookups, ", "))
}

func findVendorDir(dir string) string {
	vendorDir := filepath.Join(dir, "vendor")
	st, err := os.Stat(vendorDir)
	if !os.IsNotExist(err) && err != nil {
		return ""
	} else if err == nil && st.IsDir() {
		pth, err := filepath.Abs(vendorDir)
		if err != nil {
			panic(err)
		}
		return pth
	}
	up, err := filepath.Abs(filepath.Join(dir, ".."))
	if err != nil {
		panic(err)
	}
	if up == dir {
		return ""
	}
	return findVendorDir(up)
}

var builtinTypes = map[string]bool{
	"string":      true,
	"bool":        true,
	"byte":        true,
	"uint":        true,
	"uint8":       true,
	"uint16":      true,
	"uint32":      true,
	"uint64":      true,
	"int":         true,
	"int8":        true,
	"int16":       true,
	"int32":       true,
	"int64":       true,
	"float32":     true,
	"float64":     true,
	"interface{}": true,
	"struct{}":    true, //temporary hack
	"error":       true,
}
