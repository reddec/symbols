package symbols

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScan(t *testing.T) {
	pack := "github.com/reddec/astools"

	proj, err := ProjectByPackage(pack)

	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, "github.com/reddec/astools", proj.Package.Import)
	t.Log(proj.Package.Import)
	f := proj.Package.Files[0]
	t.Log(f.Filename)

	t.Log(proj.FindPackageImport("ast", f))

	sym, err := proj.FindSymbol("Struct", f)
	assert.NoError(t, err, "find symbol")
	assert.True(t, sym.IsStruct(), "it's a struct")

	fields, err := sym.Fields(proj)
	assert.NoError(t, err, "struct fields")

	for _, field := range fields {
		t.Log(field.Name, field.Type)
	}
}

func TestVars(t *testing.T) {
	pack := "github.com/reddec/liana/cmd/liana"
	proj, err := ProjectByPackage(pack)

	if err != nil {
		t.Error(err)
		return
	}
	for _, name := range proj.Names() {
		sym := proj.Package.FindSymbol(name)
		if assert.NotNil(t, sym, "symbol "+name) {
			t.Log(name)
			t.Log(sym.VarType())
		}
	}
	fmt.Println(proj.Names())

}

func TestScanPackage(t *testing.T) {
	dir := "."
	proj, err := ProjectByDir(dir)
	assert.NoError(t, err)
	assert.Equal(t, "symbols", proj.Package.Package)
}

func TestAliases(t *testing.T) {
	dir := "./sample"
	proj, err := ProjectByDir(dir)
	assert.NoError(t, err)
	sym, err := proj.FindSymbol("empty.Header", proj.Package.Files[0])
	assert.NoError(t, err)
	assert.Equal(t, sym.Import.Import, "net/http")
}

type SampleIface interface {
	Greet(name string) (string, error)
}

func TestSymbol_Methods(t *testing.T) {
	proj, err := ProjectByDir(".")
	assert.NoError(t, err)
	assert.Equal(t, "symbols", proj.Package.Package)
	sym, err := proj.FindLocalSymbol("SampleIface")
	assert.NoError(t, err, "find interface")
	assert.True(t, sym.IsInterface())
	methods, err := sym.Methods(proj)
	assert.NoError(t, err, "methods")
	t.Log(methods)
}
