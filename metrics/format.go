// Package metrics normalizes font metrics that show up in inconsistent
// text formats: mixed key casing, ":" vs "=" vs bare whitespace as the
// separator, stray units on numbers, and aliased field names.
package metrics

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// fieldOrder is the canonical output order. Fields not in this list are
// still kept, just sorted alphabetically after the known ones.
var fieldOrder = []string{
	"font-family",
	"units-per-em",
	"ascent",
	"descent",
	"line-gap",
	"cap-height",
	"x-height",
	"underline-position",
	"underline-thickness",
}

var fieldRank = func() map[string]int {
	r := make(map[string]int, len(fieldOrder))
	for i, f := range fieldOrder {
		r[f] = i
	}
	return r
}()

// aliases maps a loosely-normalized spelling (lowercased, letters and
// digits only) to the canonical field name. This is how "CapHeight",
// "cap_height" and "cap height" all end up as "cap-height".
var aliases = map[string]string{
	"fontfamily":         "font-family",
	"family":             "font-family",
	"name":               "font-family",
	"unitsperem":         "units-per-em",
	"upm":                "units-per-em",
	"ascent":             "ascent",
	"ascender":           "ascent",
	"descent":            "descent",
	"descender":          "descent",
	"linegap":            "line-gap",
	"capheight":          "cap-height",
	"xheight":            "x-height",
	"underlineposition":  "underline-position",
	"underlinethickness": "underline-thickness",
}

// numericUnit strips a trailing unit suffix from a value like "800px" or
// "1.5 pt", leaving just the number. The unit itself carries no
// information here because every field is already expressed in font
// design units.
var numericUnit = regexp.MustCompile(`^(-?[0-9]+(?:\.[0-9]+)?)\s*(px|pt|u|units?)$`)

// Field is one normalized metric entry.
type Field struct {
	Key   string // canonical, e.g. "cap-height"
	Value string // cleaned value, unit suffix removed
}

// Parse reads messy "key <sep> value" lines and returns them normalized.
// Blank lines and lines starting with "#" are ignored. Lines that can't
// be split into a key and a value are returned as an error, named after
// the 1-based line number so a caller can report where the input broke.
func Parse(input string) ([]Field, error) {
	var fields []Field
	seen := make(map[string]bool)

	for i, raw := range strings.Split(input, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := splitKeyValue(line)
		if !ok {
			return nil, fmt.Errorf("line %d: can't find a key/value split in %q", i+1, raw)
		}

		canonical := normalizeKey(key)
		if seen[canonical] {
			return nil, fmt.Errorf("line %d: %q duplicates an earlier field", i+1, canonical)
		}
		seen[canonical] = true

		fields = append(fields, Field{
			Key:   canonical,
			Value: normalizeValue(canonical, value),
		})
	}

	sort.SliceStable(fields, func(i, j int) bool {
		ri, iKnown := fieldRank[fields[i].Key]
		rj, jKnown := fieldRank[fields[j].Key]
		switch {
		case iKnown && jKnown:
			return ri < rj
		case iKnown:
			return true
		case jKnown:
			return false
		default:
			return fields[i].Key < fields[j].Key
		}
	})

	return fields, nil
}

// splitKeyValue tries ":" or "=" first, since those are unambiguous, and
// falls back to the first run of whitespace when the input just has a
// key and a value side by side.
func splitKeyValue(line string) (key, value string, ok bool) {
	if idx := strings.IndexAny(line, ":="); idx > 0 {
		return strings.TrimSpace(line[:idx]), strings.TrimSpace(line[idx+1:]), true
	}
	if idx := strings.IndexFunc(line, func(r rune) bool { return r == ' ' || r == '\t' }); idx > 0 {
		return strings.TrimSpace(line[:idx]), strings.TrimSpace(line[idx+1:]), true
	}
	return "", "", false
}

func normalizeKey(raw string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(raw) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	folded := b.String()
	if canonical, ok := aliases[folded]; ok {
		return canonical
	}
	// No known alias: fall back to a hyphenated form of whatever was given
	// so the output still looks consistent instead of echoing raw input.
	return strings.Join(strings.Fields(strings.ToLower(raw)), "-")
}

func normalizeValue(key, raw string) string {
	raw = strings.Trim(raw, `"'`)
	raw = strings.TrimSpace(raw)

	if key == "font-family" {
		return raw
	}

	if m := numericUnit.FindStringSubmatch(strings.ToLower(raw)); m != nil {
		raw = m[1]
	}

	if f, err := strconv.ParseFloat(raw, 64); err == nil {
		if f == float64(int64(f)) {
			return strconv.FormatInt(int64(f), 10)
		}
		return strconv.FormatFloat(f, 'f', -1, 64)
	}

	return raw
}

// Format renders fields as aligned "key: value" lines, one per field, in
// canonical order.
func Format(fields []Field) string {
	width := 0
	for _, f := range fields {
		if len(f.Key) > width {
			width = len(f.Key)
		}
	}

	var b strings.Builder
	for _, f := range fields {
		fmt.Fprintf(&b, "%-*s  %s\n", width+1, f.Key+":", f.Value)
	}
	return b.String()
}
