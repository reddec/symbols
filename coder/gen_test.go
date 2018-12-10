package coder

import (
	"bytes"
	"github.com/dave/jennifer/jen"
	"github.com/reddec/symbols"
	"github.com/stretchr/testify/assert"
	"testing"
)

const sample = `package main

import "bytes"

type Hello struct {
	A      int
	B      *string
	C      []*float64
	Buffer *bytes.Buffer ` + "`json:\"xxx\"`" + `
}
`

type Hello struct {
	A      int
	B      *string
	C      []*float64
	Buffer *bytes.Buffer `json:"xxx"`
}

func TestGenerateStruct(t *testing.T) {
	out := jen.NewFile("main")
	sym, err := symbols.ProjectByDir(".")
	assert.NoError(t, err, "parse")
	st, err := sym.FindSymbol("Hello", sym.Package.FindFile("gen_test.go"))
	assert.NoError(t, err, "find struct")

	generated, err := GenerateStruct(st, sym)
	assert.NoError(t, err, "generate")
	out.Add(generated)
	buf := &bytes.Buffer{}
	err = out.Render(buf)
	assert.NoError(t, err, "render")

	assert.Equal(t, sample, buf.String(), "compare generated")
}

const sample2 = `package main

type Hello struct {
	A int
	B *string
	C []*float64
}
`

func TestMutateStruct(t *testing.T) {
	out := jen.NewFile("main")
	sym, err := symbols.ProjectByDir(".")
	assert.NoError(t, err, "parse")
	st, err := sym.FindSymbol("Hello", sym.Package.FindFile("gen_test.go"))
	assert.NoError(t, err, "find struct")
	mutated, err := MutateStruct(st, []string{"Buffer"})
	assert.NoError(t, err, "mutate")
	generated, err := GenerateStruct(mutated, sym)
	assert.NoError(t, err, "generate")
	out.Add(generated)
	buf := &bytes.Buffer{}
	err = out.Render(buf)
	assert.NoError(t, err, "render")

	assert.Equal(t, sample2, buf.String(), "compare generated")
}
