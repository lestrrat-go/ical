package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/pkg/errors"
)

func main() {
	if err := _main(); err != nil {
		log.Printf("%s", err)
		os.Exit(1)
	}
}

type definition struct {
	Name                         string   `json:"name"`
	Type                         string   `json:"type"`
	MandatoryUniqueProperties    []string `json:"mandatory_unique_properties"`
	OptionalRepeatableProperties []string `json:"optional_repeatable_properties"`
	OptionalUniqueProperties     []string `json:"optional_unique_properties"`
	SkipConstructor              bool     `json:"skip_constructor"`
}

func fieldName(s string) string {
	var upNext bool
	var buf bytes.Buffer
	for _, r := range s {
		if upNext {
			r = unicode.ToUpper(r)
			upNext = false
		} else if r == '-' {
			upNext = true
			continue
		}

		buf.WriteRune(r)
	}

	return buf.String()
}

func _main() error {
	if len(os.Args) < 2 {
		return errors.New(`usage: gentypes [definition-file]`)
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		return errors.Wrapf(err, `failed to open file %s`, os.Args[1])
	}
	defer f.Close()

	// Read from the definition file
	var defs []*definition
	if err := json.NewDecoder(f).Decode(&defs); err != nil {
		return errors.Wrap(err, `failed to read from json file`)
	}

	for _, def := range defs {
		if err := writeType(def); err != nil {
			return errors.Wrapf(err, `failed to generate type %s`, def.Name)
		}
	}
	return nil
}

func writeType(def *definition) error {
	dst := &bytes.Buffer{}

	fmt.Fprintf(dst, "package ical")
	fmt.Fprintf(dst, "\n\n// THIS FILE IS AUTO-GENERATED BY internal/cmd/gentypes/gentypes.go")
	fmt.Fprintf(dst, "\n// DO NOT EDIT. ALL CHANGES WILL BE LOST")

	writeImports := func(dst io.Writer, imports []string) {
		for _, lib := range imports {
			fmt.Fprintf(dst, "\n%s", strconv.Quote(lib))
		}
	}
	fmt.Fprintf(dst, "\n\nimport (")
	writeImports(dst, []string{"bytes", "strings"})
	fmt.Fprintf(dst, "\n")
	writeImports(dst, []string{"github.com/pkg/errors"})
	fmt.Fprintf(dst, "\n)")

	fmt.Fprintf(dst, "\n\ntype %s struct {", def.Name)
	fmt.Fprintf(dst, "\nentries EntryList")
	fmt.Fprintf(dst, "\nprops *PropertySet")
	fmt.Fprintf(dst, "\n}")

	if !def.SkipConstructor {
		fmt.Fprintf(dst, "\n\nfunc New%s() *%s {", def.Name, def.Name)
		fmt.Fprintf(dst, "\nreturn &%s{", def.Name)
		fmt.Fprintf(dst, "\nprops: NewPropertySet(),")
		fmt.Fprintf(dst, "\n}")
		fmt.Fprintf(dst, "\n}")
	}

	fmt.Fprintf(dst, "\n\nfunc (v *%s) String() string {", def.Name)
	fmt.Fprintf(dst, "\nvar buf bytes.Buffer")
	fmt.Fprintf(dst, "\nNewEncoder(&buf).Encode(v)")
	fmt.Fprintf(dst, "\nreturn buf.String()")
	fmt.Fprintf(dst, "\n}")

	fmt.Fprintf(dst, "\n\nfunc (v %s) Type() string {", def.Name)
	fmt.Fprintf(dst, "\nreturn %s", strconv.Quote(def.Type))
	fmt.Fprintf(dst, "\n}")

	fmt.Fprintf(dst, "\n\nfunc (v *%s) AddEntry(e Entry) error {", def.Name)
	fmt.Fprintf(dst, "\nv.entries.Append(e)")
	fmt.Fprintf(dst, "\nreturn nil")
	fmt.Fprintf(dst, "\n}")

	fmt.Fprintf(dst, "\n\nfunc (v *%s) Entries() <-chan Entry {", def.Name)
	fmt.Fprintf(dst, "\nreturn v.entries.Iterator()")
	fmt.Fprintf(dst, "\n}")

	fmt.Fprintf(dst, "\n\nfunc (v *%s) GetProperty(name string) (*Property, bool) {", def.Name)
	fmt.Fprintf(dst, "\nreturn v.props.GetFirst(name)")
	fmt.Fprintf(dst, "\n}")

	fmt.Fprintf(dst, "\n\nfunc (v *%s) Properties() <-chan *Property {", def.Name)
	fmt.Fprintf(dst, "\nreturn v.props.Iterator()")
	fmt.Fprintf(dst, "\n}")

	fmt.Fprintf(dst, "\n\nfunc (v *%s) AddProperty(key, value string, options ...PropertyOption) error {", def.Name)
	fmt.Fprintf(dst, "\nvar params Parameters")
	fmt.Fprintf(dst, "\nvar force bool")
	fmt.Fprintf(dst, "\nfor _, option := range options {")
	fmt.Fprintf(dst, "\nswitch option.Name() {")
	fmt.Fprintf(dst, "\ncase \"Parameters\":")
	fmt.Fprintf(dst, "\nparams = option.Get().(Parameters)")
	fmt.Fprintf(dst, "\ncase \"Force\":")
	fmt.Fprintf(dst, "\nforce = option.Get().(bool)")
	fmt.Fprintf(dst, "\n}")
	fmt.Fprintf(dst, "\n}")

	fmt.Fprintf(dst, "\n")

	fmt.Fprintf(dst, "\n\nswitch key = strings.ToLower(key); key {")
	props := append(def.MandatoryUniqueProperties, def.OptionalUniqueProperties...)
	if lprops := len(props); lprops > 0 {
		fmt.Fprintf(dst, "\ncase ")
		for i, prop := range props {
			fmt.Fprintf(dst, "%s", strconv.Quote(prop))
			if i < lprops-1 {
				fmt.Fprintf(dst, ", ")
			}
		}
		fmt.Fprintf(dst, ":\nv.props.Set(NewProperty(key, value, params))")
	}

	props = def.OptionalRepeatableProperties
	if lprops := len(props); lprops > 0 {
		fmt.Fprintf(dst, "\ncase ")
		for i, prop := range props {
			fmt.Fprintf(dst, "%s", strconv.Quote(prop))
			if i < lprops-1 {
				fmt.Fprintf(dst, ", ")
			}
		}
		fmt.Fprintf(dst, ":\nv.props.Append(NewProperty(key, value, params))")
	}

	fmt.Fprintf(dst, "\ndefault:")
	fmt.Fprintf(dst, "\nif strings.HasPrefix(key, \"x-\") || force {")
	fmt.Fprintf(dst, "\nv.props.Append(NewProperty(key, value, params))")
	fmt.Fprintf(dst, "\n} else {")
	fmt.Fprintf(dst, "\nreturn errors.Errorf(`invalid property %%s`, key)")
	fmt.Fprintf(dst, "\n} /* end if */")
	fmt.Fprintf(dst, "\n}")
	fmt.Fprintf(dst, "\nreturn nil")
	fmt.Fprintf(dst, "\n}")

	formatted, err := format.Source(dst.Bytes())
	if err != nil {
		os.Stderr.Write(dst.Bytes())
		return errors.Wrap(err, `failed to format source code`)
	}

	filename := strings.ToLower(def.Name) + "_gen.go"
	f, err := os.Create(filename)
	if err != nil {
		return errors.Wrapf(err, `failed to open %s for writing`, filename)
	}
	defer f.Close()

	f.Write(formatted)

	return nil
}
