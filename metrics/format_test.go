package metrics

import (
	"strings"
	"testing"
)

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

func TestIsKnown(t *testing.T) {
	if !IsKnown("cap-height") {
		t.Error("cap-height should be a known canonical field")
	}
	if IsKnown("italic-angle") {
		t.Error("italic-angle has no alias and should not be known")
	}
}

func TestParseHandlesRealAFMFile(t *testing.T) {
	input := "StartFontMetrics 4.1\n" +
		"Comment Generated for testing\n" +
		"FamilyName Helvetica\n" +
		"Ascender 718\n" +
		"Descender -207\n" +
		"CapHeight 718\n" +
		"XHeight 523\n" +
		"UnderlinePosition -100\n" +
		"UnderlineThickness 50\n" +
		"StartCharMetrics 3\n" +
		"C 32 ; WX 278 ; N space ; B 0 0 0 0 ;\n" +
		"C 33 ; WX 278 ; N exclam ; B 90 0 187 718 ;\n" +
		"C 34 ; WX 355 ; N quotedbl ; B 70 463 285 718 ;\n" +
		"EndCharMetrics\n" +
		"StartKernData\n" +
		"StartKernPairs 1\n" +
		"KPX A V -70\n" +
		"EndKernPairs\n" +
		"EndKernData\n" +
		"EndFontMetrics\n"

	fields, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	want := map[string]string{
		"font-family":         "Helvetica",
		"ascent":              "718",
		"descent":             "-207",
		"cap-height":          "718",
		"x-height":            "523",
		"underline-position":  "-100",
		"underline-thickness": "50",
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

	for _, unwanted := range []string{"c", "kpx"} {
		if _, ok := got[unwanted]; ok {
			t.Errorf("expected char/kern metrics to be skipped, but found field %q", unwanted)
		}
	}
}

func TestParseAllSplitsOnBlankLines(t *testing.T) {
	input := "Family: Arial\nAscender: 905\n\nFamily: Georgia\nAscender: 900\n"

	instances, err := ParseAll(input)
	if err != nil {
		t.Fatalf("ParseAll returned error: %v", err)
	}
	if len(instances) != 2 {
		t.Fatalf("got %d instances, want 2", len(instances))
	}

	want := []string{"Arial", "Georgia"}
	for i, fields := range instances {
		got := ""
		for _, f := range fields {
			if f.Key == "font-family" {
				got = f.Value
			}
		}
		if got != want[i] {
			t.Errorf("instance %d font-family = %q, want %q", i, got, want[i])
		}
	}
}

func TestParseAllSplitsConcatenatedAFMFiles(t *testing.T) {
	afm := func(family string, ascent string) string {
		return "StartFontMetrics 4.1\n" +
			"FamilyName " + family + "\n" +
			"Ascender " + ascent + "\n" +
			"EndFontMetrics\n"
	}
	input := afm("Helvetica", "718") + afm("Courier", "629")

	instances, err := ParseAll(input)
	if err != nil {
		t.Fatalf("ParseAll returned error: %v", err)
	}
	if len(instances) != 2 {
		t.Fatalf("got %d instances, want 2", len(instances))
	}

	wantFamily := []string{"Helvetica", "Courier"}
	wantAscent := []string{"718", "629"}
	for i, fields := range instances {
		got := make(map[string]string, len(fields))
		for _, f := range fields {
			got[f.Key] = f.Value
		}
		if got["font-family"] != wantFamily[i] {
			t.Errorf("instance %d font-family = %q, want %q", i, got["font-family"], wantFamily[i])
		}
		if got["ascent"] != wantAscent[i] {
			t.Errorf("instance %d ascent = %q, want %q", i, got["ascent"], wantAscent[i])
		}
	}
}

func TestParseAllReportsLineNumbersRelativeToWholeInput(t *testing.T) {
	input := "Family: Arial\n\nthisisnotakeyvaluepair\n"

	_, err := ParseAll(input)
	if err == nil {
		t.Fatal("expected an error for a line with no key/value split, got nil")
	}
	if got, want := err.Error(), "line 3:"; !strings.HasPrefix(got, want) {
		t.Errorf("error = %q, want it to start with %q", got, want)
	}
}
