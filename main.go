// fontfmt normalizes font metrics files that come from mismatched
// tools: some spit out "CapHeight: 700", others "cap_height=700px" or
// "cap height 700". It reads one or more files, or stdin when none are
// given, and prints each normalized to the same key/value shape.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/matthewr377/fontfmt/metrics"
)

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "fontfmt:", err)
		os.Exit(1)
	}
}

// namedFields ties a set of parsed fields back to the input it came
// from, so output covering more than one font (multiple files, or
// multiple instances found within one file) can tell them apart via
// text headers or the "file" key in JSON.
type namedFields struct {
	Label  string          `json:"file"`
	Fields []metrics.Field `json:"fields"`
}

func run(args []string, stdin io.Reader, stdout io.Writer) error {
	fs := flag.NewFlagSet("fontfmt", flag.ContinueOnError)
	jsonOut := fs.Bool("json", false, "print fields as JSON instead of aligned text")
	check := fs.Bool("check", false, "fail if any field isn't a recognized canonical name")
	if err := fs.Parse(args); err != nil {
		return err
	}
	paths := fs.Args()

	var results []namedFields
	if len(paths) == 0 {
		instances, err := readInstances("(stdin)", stdin)
		if err != nil {
			return err
		}
		results = append(results, labelInstances("(stdin)", instances)...)
	} else {
		for _, path := range paths {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			instances, err := readInstances(path, f)
			f.Close()
			if err != nil {
				return err
			}
			results = append(results, labelInstances(path, instances)...)
		}
	}

	if *check {
		if err := checkKnown(results); err != nil {
			return err
		}
	}

	if *jsonOut {
		return writeJSON(results, stdout)
	}
	return writeText(results, stdout)
}

// checkKnown reports an error naming every field whose key isn't one of
// the canonical names metrics.Parse recognizes via alias, so --check
// catches typos or unfamiliar metrics instead of silently passing them
// through.
func checkKnown(results []namedFields) error {
	for _, r := range results {
		var unknown []string
		for _, f := range r.Fields {
			if !metrics.IsKnown(f.Key) {
				unknown = append(unknown, f.Key)
			}
		}
		if len(unknown) > 0 {
			return fmt.Errorf("%s: unknown field(s): %s", r.Label, strings.Join(unknown, ", "))
		}
	}
	return nil
}

// labelInstances turns the font instances parsed out of one input into
// namedFields entries. The label is only suffixed with "#n" when the
// input actually held more than one instance, so the common case of one
// font per file keeps the plain label it always had.
func labelInstances(label string, instances [][]metrics.Field) []namedFields {
	if len(instances) <= 1 {
		var fields []metrics.Field
		if len(instances) == 1 {
			fields = instances[0]
		}
		return []namedFields{{Label: label, Fields: fields}}
	}

	named := make([]namedFields, len(instances))
	for i, fields := range instances {
		named[i] = namedFields{Label: fmt.Sprintf("%s#%d", label, i+1), Fields: fields}
	}
	return named
}

func readInstances(label string, r io.Reader) ([][]metrics.Field, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", label, err)
	}

	instances, err := metrics.ParseAll(string(raw))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", label, err)
	}
	return instances, nil
}

func writeText(results []namedFields, w io.Writer) error {
	for i, r := range results {
		if len(results) > 1 {
			fmt.Fprintf(w, "== %s ==\n", r.Label)
		}
		fmt.Fprint(w, metrics.Format(r.Fields))
		if len(results) > 1 && i < len(results)-1 {
			fmt.Fprintln(w)
		}
	}
	return nil
}

// writeJSON emits a single field array for one input, or an array of
// {file, fields} objects when there's more than one, so the shape of
// the output doesn't carry a pointless wrapper in the common case.
func writeJSON(results []namedFields, w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if len(results) == 1 {
		return enc.Encode(results[0].Fields)
	}
	return enc.Encode(results)
}
