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

	"github.com/matthewr377/fontfmt/metrics"
)

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "fontfmt:", err)
		os.Exit(1)
	}
}

// namedFields ties a set of parsed fields back to the input it came
// from, so multi-file output (text headers, or the "file" key in JSON)
// can tell them apart.
type namedFields struct {
	Label  string          `json:"file"`
	Fields []metrics.Field `json:"fields"`
}

func run(args []string, stdin io.Reader, stdout io.Writer) error {
	fs := flag.NewFlagSet("fontfmt", flag.ContinueOnError)
	jsonOut := fs.Bool("json", false, "print fields as JSON instead of aligned text")
	if err := fs.Parse(args); err != nil {
		return err
	}
	paths := fs.Args()

	var results []namedFields
	if len(paths) == 0 {
		fields, err := readFields("(stdin)", stdin)
		if err != nil {
			return err
		}
		results = append(results, namedFields{Label: "(stdin)", Fields: fields})
	} else {
		for _, path := range paths {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			fields, err := readFields(path, f)
			f.Close()
			if err != nil {
				return err
			}
			results = append(results, namedFields{Label: path, Fields: fields})
		}
	}

	if *jsonOut {
		return writeJSON(results, stdout)
	}
	return writeText(results, stdout)
}

func readFields(label string, r io.Reader) ([]metrics.Field, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", label, err)
	}

	fields, err := metrics.Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", label, err)
	}
	return fields, nil
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
