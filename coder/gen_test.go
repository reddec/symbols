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
	A      int           // comment
	B      *string       // another comment
	C      []*float64    // also
	Buffer *bytes.Buffer ` + "`json:\"xxx\"`" + ` // text
}
`

type Hello struct {
	A      int           // comment
	B      *string       // another comment
	C      []*float64    // also
	Buffer *bytes.Buffer `json:"xxx"` // text
}

func TestGenerateStruct(t *testing.T) {
	out := jen.NewFile("main")
	sym, err := symbols.ProjectByDir(".", symbols.All)
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
	A int        // comment
	B *string    // another comment
	C []*float64 // also
}
`

func TestMutateStruct(t *testing.T) {
	out := jen.NewFile("main")
	sym, err := symbols.ProjectByDir(".", symbols.All)
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
	// should not change original one
	out = jen.NewFile("main")
	generated, err = GenerateStruct(st, sym)
	assert.NoError(t, err, "generate")
	out.Add(generated)
	buf = &bytes.Buffer{}
	err = out.Render(buf)
	assert.NoError(t, err, "render")

	assert.Equal(t, sample, buf.String(), "compare generated")
}

type A struct {
	Request string
}

type UserA struct {
	UserID  int64
	Request string
}

const sample3 = `package main

import coder "github.com/reddec/symbols/coder"

func MapA(srcA *coder.A, userID int64) *coder.UserA {
	var destUserA coder.UserA
	destUserA.UserID = userID
	destUserA.Request = srcA.Request
	return &destUserA
}
`

func TestGenerateStructMapper(t *testing.T) {
	out := jen.NewFile("main")
	sym, err := symbols.ProjectByDir(".", symbols.All)
	assert.NoError(t, err, "parse")
	stA, err := sym.FindSymbol("A", sym.Package.FindFile("gen_test.go"))
	assert.NoError(t, err, "find struct A")
	userA, err := sym.FindSymbol("UserA", sym.Package.FindFile("gen_test.go"))
	assert.NoError(t, err, "find struct UserA")

	generated, err := GenerateStructMapper(stA, userA, sym, "MapA", true)
	assert.NoError(t, err, "generate")
	out.Add(generated)
	buf := &bytes.Buffer{}
	err = out.Render(buf)
	assert.NoError(t, err, "render")

	assert.Equal(t, sample3, buf.String(), "compare generated")
}
