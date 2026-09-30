# fontmetrics

A font's vertical metrics decide how much space a line of text takes
up and where the baseline sits. The catch is that a single TrueType
or OpenType file stores three different, overlapping sets of these
numbers:

- `hhea.ascender` / `hhea.descender` / `hhea.lineGap`
- `OS/2.sTypoAscender` / `sTypoDescender` / `sTypoLineGap`
- `OS/2.usWinAscent` / `usWinDescent`

Font tools disagree about which set to trust, so the same font can
render with visibly different line spacing in a browser, in Word, and
in a design tool, depending on which values that particular renderer
picked. When line-height looks wrong and you suspect the font itself,
you need to see all three sets side by side. `fontmetrics` reads a
font file directly and prints them.

## Usage

```
$ go build -o fontmetrics .
$ ./fontmetrics ./Inter-Regular.ttf
units per em: 2048
glyphs:       3895

hhea (used by most text-layout engines):
  ascender:    1984  (0.969 em)
  descender:   -494  (-0.241 em)
  line gap:       0  (0.000 em)

OS/2 typo metrics (what browsers prefer, if USE_TYPO_METRICS is set):
  typoAscender:    1901  (0.928 em)
  typoDescender:   -483  (-0.236 em)
  typoLineGap:      141  (0.069 em)

OS/2 win metrics (used for clipping on Windows; often larger than the others):
  winAscent:    2075  (1.013 em)
  winDescent:    494  (0.241 em)

  capHeight: 1462  (0.714 em)
  xHeight:   1096  (0.535 em)
```

Pass `-json` to get the same numbers as a JSON object (design units
only, with the `os2` key left out when the font has no OS/2 table):

```
$ ./fontmetrics -json ./Inter-Regular.ttf
```

Text values are printed both in raw font design units and as a fraction of
the em, since the design-unit numbers are meaningless without knowing
`unitsPerEm`.

## Scope

This reads `.ttf` and `.otf` files by parsing the SFNT table
directory directly - no font libraries involved. It currently
supports the tables every font is expected to have (`head`, `hhea`,
`maxp`) plus `OS/2` when present. Font collections (`.ttc`) and
variable font instances are not handled yet.

## License

MIT, see LICENSE.
