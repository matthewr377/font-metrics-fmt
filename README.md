# fontfmt

Font metrics show up in whatever shape the tool that produced them
decided to use. One export writes `CapHeight: 700`, another writes
`cap_height=700px`, another writes `cap height 700` with no separator
at all. If you're diffing metrics across font versions or feeding them
into something downstream, that inconsistency is annoying to deal with
by hand.

`fontfmt` reads that kind of loose key/value text and prints it back
out with consistent field names, a consistent separator, and units
stripped from numbers (everything is already in font design units, so
a trailing `px` or `pt` carries no information).

## Example

Input (`arial.metrics`):

```
Family: Arial
UnitsPerEm=1000
Ascender: 905
descender -212
cap_height 716px
x-height 519
```

```
$ fontfmt arial.metrics
font-family:   Arial
units-per-em:  1000
ascent:        905
descent:       -212
cap-height:    716
x-height:      519
```

## Usage

```
fontfmt [-json] [-check] [file ...]
```

With no arguments, it reads from stdin:

```
$ cat arial.metrics | fontfmt
```

With more than one file, each block is printed under a `== path ==`
header so you can tell them apart.

A single file or stdin stream can also hold more than one font: a blank
line separates one plain key/value block from the next, and several AFM
files can just be concatenated, since each opens with its own
`StartFontMetrics`. When an input holds more than one instance, its
blocks are numbered `== path#1 ==`, `== path#2 ==`, and so on.

Pass `-json` to get the same fields as JSON instead of aligned text.
For a single input that's a flat array of `{"key", "value"}` objects,
in the same canonical order as the text output:

```
$ fontfmt -json arial.metrics
[
  {
    "key": "font-family",
    "value": "Arial"
  },
  ...
]
```

With more than one file, or more than one instance found in a single
input, the output is an array of `{"file", "fields"}` objects instead,
one per font, so they stay distinguishable.

Lines starting with `#` are treated as comments and skipped. A line
that can't be split into a key and a value, or a field that's given
twice under two different aliases, is reported as an error with the
line number.

Pass `-check` to fail (with a non-zero exit and a message naming the
fields) if the input contains anything outside the recognized field
list below. Without it, unrecognized fields are passed through as-is
instead of being dropped, which is fine for a quick look but not for
catching a typo or a metric the tool doesn't know about yet.

Real [AFM](https://learn.microsoft.com/en-us/typography/opentype/spec/afm)
files work too: `fontfmt some.afm` picks the font-level fields (`FamilyName`,
`Ascender`, `CapHeight`, and so on) out of the header and ignores the
`Comment` lines, the `StartFontMetrics`/`EndFontMetrics` wrapper, and the
`StartCharMetrics`/`StartKernData`/`StartComposites` blocks, since those hold
per-glyph and per-pair data rather than metrics for the font as a whole.

## Recognized fields

`font-family`, `units-per-em`, `ascent`, `descent`, `line-gap`,
`cap-height`, `x-height`, `underline-position`, `underline-thickness`.

Common aliases for these (`family`, `upm`, `ascender`, `descender`,
`CapHeight`, ...) are recognized regardless of case or whether the
words are separated by a hyphen, underscore, or space. Anything not on
that list is passed through, just with whitespace collapsed, so
normalizing doesn't silently drop data the tool doesn't know about
yet.

## Install

```
go install github.com/matthewr377/fontfmt@latest
```

Or clone and `go build`. No dependencies outside the standard library.

## License

MIT, see [LICENSE](LICENSE).
