package main

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite golden files from current output")

// TestGolden feeds each testdata/<name>.in through the CLI and compares
// the result to testdata/<name>.<mode>.golden. Run with -update after an
// intentional output change, then review the diff before committing.
func TestGolden(t *testing.T) {
	cases := []struct {
		input string
		mode  string
		args  []string
	}{
		{"arial", "text", nil},
		{"arial", "json", []string{"-json"}},
		{"sample.afm", "text", nil},
		{"two-fonts", "text", nil},
	}

	for _, tc := range cases {
		t.Run(tc.input+"/"+tc.mode, func(t *testing.T) {
			in, err := os.ReadFile(filepath.Join("testdata", tc.input+".in"))
			if err != nil {
				t.Fatal(err)
			}

			var stdout bytes.Buffer
			if err := run(tc.args, strings.NewReader(string(in)), &stdout); err != nil {
				t.Fatalf("run returned error: %v", err)
			}

			goldenPath := filepath.Join("testdata", tc.input+"."+tc.mode+".golden")
			if *update {
				if err := os.WriteFile(goldenPath, stdout.Bytes(), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}

			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatal(err)
			}
			if got := stdout.String(); got != string(want) {
				t.Errorf("output differs from %s\n got:\n%s\nwant:\n%s", goldenPath, got, want)
			}
		})
	}
}

func TestGoldenMultipleFiles(t *testing.T) {
	var stdout bytes.Buffer
	args := []string{
		filepath.Join("testdata", "arial.in"),
		filepath.Join("testdata", "sample.afm.in"),
	}
	if err := run(args, strings.NewReader(""), &stdout); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	got := stdout.String()
	for _, header := range []string{"== " + args[0] + " ==\n", "== " + args[1] + " ==\n"} {
		if !strings.Contains(got, header) {
			t.Errorf("output is missing header %q:\n%s", header, got)
		}
	}
}
