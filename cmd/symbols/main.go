package main

import (
	"github.com/dave/jennifer/jen"
	"github.com/jessevdk/go-flags"
	"github.com/reddec/symbols"
	"github.com/reddec/symbols/coder"
	"os"
)

func main() {
	parser := flags.NewParser(nil, flags.Default)
	parser.AddCommand("mutate", "mutate struct", "mutate struct and generate mappers for them", &mutateStruct{})
	_, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}
}

type mutateStruct struct {
	SourceStruct string   `long:"source" env:"SOURCE_STRUCT" description:"Name of source struct" required:"true"`
	Target       string   `long:"target" env:"TARGET" description:"Name of target struct" required:"true"`
	Map          string   `long:"map" env:"MAP" description:"Name of function to map from source to target"`
	Unmap        string   `long:"unmap" env:"UNMAP" description:"Name of function to map form target to source"`
	SelfMap      string   `long:"self-map" env:"SELF_MAP" description:"Name of function to map from source to target (self)"`
	SelfUnmap    string   `long:"self-unmap" env:"SELF_UNMAP" description:"Name of function to map form target to source"`
	Exclude      []string `long:"exclude" env:"EXCLUDE" description:"Exclude fields"`
	Value        bool     `long:"value" env:"VALUE" description:"Map items passed by value"`
}

func (m *mutateStruct) Execute([]string) error {
	proj, err := symbols.ProjectByDir(".")
	if err != nil {
		return err
	}
	sym, err := proj.FindLocalSymbol(m.SourceStruct)
	if err != nil {
		return err
	}
	mutated, err := coder.MutateStruct(sym, m.Exclude)
	if err != nil {
		return err
	}
	mutated.Name = m.Target
	mutated.Import = proj.Package

	out := jen.NewFilePathName(proj.Package.Import, proj.Package.Package)

	generated, err := coder.GenerateStruct(mutated, proj)
	if err != nil {
		return err
	}

	out.Add(generated)

	if m.Map != "" {
		// generate map to destination
		generated, err = coder.GenerateStructMapper(sym, mutated, proj, m.Map, !m.Value)
		if err != nil {
			return err
		}
		out.Add(generated)
	}
	if m.Unmap != "" {
		// generate map to source
		generated, err = coder.GenerateStructMapper(mutated, sym, proj, m.Unmap, !m.Value)
		if err != nil {
			return err
		}
		out.Add(generated)
	}
	if m.SelfMap != "" {
		// generate map to destination
		generated, err = coder.GenerateSelfStructMapper(sym, mutated, proj, m.SelfMap, !m.Value)
		if err != nil {
			return err
		}
		out.Add(generated)
	}
	if m.SelfUnmap != "" {
		// generate map to source
		generated, err = coder.GenerateSelfStructMapper(mutated, sym, proj, m.SelfUnmap, !m.Value)
		if err != nil {
			return err
		}
		out.Add(generated)
	}

	return out.Render(os.Stdout)
}
