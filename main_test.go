package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunJSONSingleInput(t *testing.T) {
	stdin := strings.NewReader("Family: Arial\nAscender: 905\n")
	var stdout bytes.Buffer

	if err := run([]string{"-json"}, stdin, &stdout); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	got := stdout.String()
	want := `[
  {
    "key": "font-family",
    "value": "Arial"
  },
  {
    "key": "ascent",
    "value": "905"
  }
]
`
	if got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestRunTextDefault(t *testing.T) {
	stdin := strings.NewReader("Family: Arial\n")
	var stdout bytes.Buffer

	if err := run(nil, stdin, &stdout); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if got, want := stdout.String(), "font-family:  Arial\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestRunRejectsUnknownFlag(t *testing.T) {
	var stdout bytes.Buffer
	if err := run([]string{"-nope"}, strings.NewReader(""), &stdout); err == nil {
		t.Fatal("expected an error for an unknown flag, got nil")
	}
}

func TestRunCheckPassesOnKnownFields(t *testing.T) {
	stdin := strings.NewReader("Family: Arial\nAscender: 905\n")
	var stdout bytes.Buffer

	if err := run([]string{"-check"}, stdin, &stdout); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
}

func TestRunCheckFailsOnUnknownField(t *testing.T) {
	stdin := strings.NewReader("Family: Arial\nItalic Angle: -12\n")
	var stdout bytes.Buffer

	err := run([]string{"-check"}, stdin, &stdout)
	if err == nil {
		t.Fatal("expected an error for an unrecognized field, got nil")
	}
	if got := err.Error(); !strings.Contains(got, "italic-angle") {
		t.Errorf("error = %q, want it to name the unknown field %q", got, "italic-angle")
	}
}
