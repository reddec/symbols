package main

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"github.com/jessevdk/go-flags"
	"github.com/reddec/symbols"
	"github.com/reddec/symbols/coder"
	"os"
)

func main() {
	parser := flags.NewParser(nil, flags.Default)
	parser.AddCommand("mutate", "mutate struct", "mutate struct and generate mappers for them", &mutateStruct{})
	parser.AddCommand("methods", "list methods", "list all found methods in all packages", &methods{})
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
	Drop         []string `long:"drop" env:"DROP" description:"Drop fields"`
	Value        bool     `long:"value" env:"VALUE" description:"Map items passed by value"`
	ScanLimit    int      `long:"scan-limit" env:"SCAN_LIMIT" description:"Maximum amount of packages to scan. -1 - all" default:"-1"`
}

func (m *mutateStruct) Execute([]string) error {
	proj, err := symbols.ProjectByDir(".", m.ScanLimit)
	if err != nil {
		return err
	}
	sym, err := proj.FindLocalSymbol(m.SourceStruct)
	if err != nil {
		return err
	}

	// imitate drop by excluding field from source
	sym, err = coder.MutateStruct(sym, m.Drop)
	if err != nil {
		return err
	}

	mutated, err := coder.MutateStruct(sym, append(m.Exclude, m.Drop...))
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

type methods struct {
	ScanLimit int `long:"scan-limit" env:"SCAN_LIMIT" description:"Maximum amount of packages to scan. -1 - all" default:"-1"`
}

func (m *methods) Execute([]string) error {
	proj, err := symbols.ProjectByDir(".", m.ScanLimit)
	if err != nil {
		return err
	}
	for _, imp := range proj.Imports {
		err := imp.Symbols(func(sym *symbols.Symbol) error {
			fn, err := sym.Function()
			if err != nil {
				return nil
			}
			fmt.Println(fn.Name)
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
