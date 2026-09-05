package metrics

import "testing"

func TestParseNormalizesKeysAndSeparators(t *testing.T) {
	input := "Family: Helvetica\n" +
		"UnitsPerEm=1000\n" +
		"cap_height 700px\n" +
		"descender: -200\n"

	fields, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	want := map[string]string{
		"font-family":  "Helvetica",
		"units-per-em": "1000",
		"cap-height":   "700",
		"descent":      "-200",
	}

	got := make(map[string]string, len(fields))
	for _, f := range fields {
		got[f.Key] = f.Value
	}

	for key, val := range want {
		if got[key] != val {
			t.Errorf("field %q = %q, want %q", key, got[key], val)
		}
	}
}

func TestParseRejectsDuplicateFields(t *testing.T) {
	_, err := Parse("ascent: 800\nAscender: 810\n")
	if err == nil {
		t.Fatal("expected an error for duplicate field, got nil")
	}
}

func TestParseRejectsUnsplittableLine(t *testing.T) {
	_, err := Parse("thisisnotakeyvaluepair\n")
	if err == nil {
		t.Fatal("expected an error for a line with no key/value split, got nil")
	}
}
