// Command fontmetrics dumps the vertical metrics embedded in a
// TrueType or OpenType font file: the values renderers use to work
// out line height, baseline position, and cap/x-height guides.
//
// Fonts carry three overlapping sets of ascent/descent numbers
// (hhea, OS/2 typo, OS/2 win) and it's common for them to disagree,
// which is why the same font can render with different line spacing
// in different apps. This tool just prints what's actually in the
// file so you can compare them yourself.
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: fontmetrics <font-file>")
		os.Exit(2)
	}

	path := os.Args[1]
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fontmetrics: %v\n", err)
		os.Exit(1)
	}

	f, err := ParseFont(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fontmetrics: %s: %v\n", path, err)
		os.Exit(1)
	}

	printReport(f)
}

func printReport(f *Font) {
	em := float64(f.UnitsPerEm)

	fmt.Printf("units per em: %d\n", f.UnitsPerEm)
	fmt.Printf("glyphs:       %d\n", f.NumGlyphs)
	fmt.Println()

	fmt.Println("hhea (used by most text-layout engines):")
	fmt.Printf("  ascender:  %6d  (%.3f em)\n", f.Ascent, float64(f.Ascent)/em)
	fmt.Printf("  descender: %6d  (%.3f em)\n", f.Descent, float64(f.Descent)/em)
	fmt.Printf("  line gap:  %6d  (%.3f em)\n", f.LineGap, float64(f.LineGap)/em)
	fmt.Println()

	if !f.HasOS2 {
		fmt.Println("OS/2 table not present; browsers and Windows apps will fall back to hhea.")
		return
	}

	fmt.Println("OS/2 typo metrics (what browsers prefer, if USE_TYPO_METRICS is set):")
	fmt.Printf("  typoAscender:  %6d  (%.3f em)\n", f.TypoAscender, float64(f.TypoAscender)/em)
	fmt.Printf("  typoDescender: %6d  (%.3f em)\n", f.TypoDescender, float64(f.TypoDescender)/em)
	fmt.Printf("  typoLineGap:   %6d  (%.3f em)\n", f.TypoLineGap, float64(f.TypoLineGap)/em)
	fmt.Println()

	fmt.Println("OS/2 win metrics (used for clipping on Windows; often larger than the others):")
	fmt.Printf("  winAscent:  %6d  (%.3f em)\n", f.WinAscent, float64(f.WinAscent)/em)
	fmt.Printf("  winDescent: %6d  (%.3f em)\n", f.WinDescent, float64(f.WinDescent)/em)

	if f.HasCapHeight {
		fmt.Println()
		fmt.Printf("  capHeight: %6d  (%.3f em)\n", f.CapHeight, float64(f.CapHeight)/em)
		fmt.Printf("  xHeight:   %6d  (%.3f em)\n", f.XHeight, float64(f.XHeight)/em)
	}
}
