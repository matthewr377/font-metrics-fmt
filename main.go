// fontfmt normalizes font metrics files that come from mismatched
// tools: some spit out "CapHeight: 700", others "cap_height=700px" or
// "cap height 700". It reads one or more files, or stdin when none are
// given, and prints each normalized to the same key/value shape.
package main

import (
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

func run(args []string, stdin io.Reader, stdout io.Writer) error {
	if len(args) == 0 {
		return formatOne("(stdin)", stdin, stdout, false)
	}

	for i, path := range args {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		err = formatOne(path, f, stdout, len(args) > 1)
		f.Close()
		if err != nil {
			return err
		}
		if len(args) > 1 && i < len(args)-1 {
			fmt.Fprintln(stdout)
		}
	}
	return nil
}

func formatOne(label string, r io.Reader, w io.Writer, withHeader bool) error {
	raw, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}

	fields, err := metrics.Parse(string(raw))
	if err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}

	if withHeader {
		fmt.Fprintf(w, "== %s ==\n", label)
	}
	fmt.Fprint(w, metrics.Format(fields))
	return nil
}
