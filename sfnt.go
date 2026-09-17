package main

import (
	"encoding/binary"
	"fmt"
)

// Font holds the metrics we care about, pulled out of a handful of
// required and optional SFNT tables. Everything is in font design
// units; divide by UnitsPerEm to get em-relative values.
type Font struct {
	UnitsPerEm  uint16
	NumGlyphs   uint16
	NumHMetrics uint16

	// From hhea. Always present in a valid font.
	Ascent    int16
	Descent   int16
	LineGap   int16

	// From OS/2, when present. HasOS2 is false for the rare font
	// (mostly old CFF fonts) that omits the table entirely.
	HasOS2        bool
	TypoAscender  int16
	TypoDescender int16
	TypoLineGap   int16
	WinAscent     uint16
	WinDescent    uint16

	// Also from OS/2, only populated when the table version is >= 2.
	HasCapHeight bool
	CapHeight    int16
	XHeight      int16
}

const (
	sfntVersionTrueType = 0x00010000
	sfntVersionOpenType = 0x4F54544F // "OTTO"
	sfntVersionCollection = 0x74746366 // "ttcf"
)

// ParseFont reads the SFNT table directory out of data and extracts
// the metrics tables. data must be the full contents of a .ttf or
// .otf file.
func ParseFont(data []byte) (*Font, error) {
	if len(data) < 12 {
		return nil, fmt.Errorf("file too small to be a font (%d bytes)", len(data))
	}

	version := binary.BigEndian.Uint32(data[0:4])
	switch version {
	case sfntVersionTrueType, sfntVersionOpenType:
		// supported
	case sfntVersionCollection:
		return nil, fmt.Errorf("font collections (.ttc/.otc) are not supported yet")
	default:
		return nil, fmt.Errorf("not a TrueType/OpenType font (unrecognized sfnt version 0x%08x)", version)
	}

	numTables := int(binary.BigEndian.Uint16(data[4:6]))
	tables := make(map[string][2]uint32, numTables) // tag -> [offset, length]

	const dirEntryStart = 12
	const dirEntrySize = 16
	need := dirEntryStart + numTables*dirEntrySize
	if len(data) < need {
		return nil, fmt.Errorf("truncated table directory: need %d bytes, have %d", need, len(data))
	}

	for i := 0; i < numTables; i++ {
		rec := data[dirEntryStart+i*dirEntrySize : dirEntryStart+(i+1)*dirEntrySize]
		tag := string(rec[0:4])
		offset := binary.BigEndian.Uint32(rec[8:12])
		length := binary.BigEndian.Uint32(rec[12:16])
		tables[tag] = [2]uint32{offset, length}
	}

	getTable := func(tag string) ([]byte, error) {
		bounds, ok := tables[tag]
		if !ok {
			return nil, fmt.Errorf("font has no %q table", tag)
		}
		offset, length := bounds[0], bounds[1]
		end := uint64(offset) + uint64(length)
		if end > uint64(len(data)) {
			return nil, fmt.Errorf("%q table extends past end of file", tag)
		}
		return data[offset:end], nil
	}

	head, err := getTable("head")
	if err != nil {
		return nil, err
	}
	if len(head) < 54 {
		return nil, fmt.Errorf("head table is truncated (%d bytes)", len(head))
	}

	hhea, err := getTable("hhea")
	if err != nil {
		return nil, err
	}
	if len(hhea) < 36 {
		return nil, fmt.Errorf("hhea table is truncated (%d bytes)", len(hhea))
	}

	maxp, err := getTable("maxp")
	if err != nil {
		return nil, err
	}
	if len(maxp) < 6 {
		return nil, fmt.Errorf("maxp table is truncated (%d bytes)", len(maxp))
	}

	f := &Font{
		UnitsPerEm:  binary.BigEndian.Uint16(head[18:20]),
		NumGlyphs:   binary.BigEndian.Uint16(maxp[4:6]),
		Ascent:      int16(binary.BigEndian.Uint16(hhea[4:6])),
		Descent:     int16(binary.BigEndian.Uint16(hhea[6:8])),
		LineGap:     int16(binary.BigEndian.Uint16(hhea[8:10])),
		NumHMetrics: binary.BigEndian.Uint16(hhea[34:36]),
	}

	if os2, err := getTable("OS/2"); err == nil && len(os2) >= 78 {
		f.HasOS2 = true
		f.TypoAscender = int16(binary.BigEndian.Uint16(os2[68:70]))
		f.TypoDescender = int16(binary.BigEndian.Uint16(os2[70:72]))
		f.TypoLineGap = int16(binary.BigEndian.Uint16(os2[72:74]))
		f.WinAscent = binary.BigEndian.Uint16(os2[74:76])
		f.WinDescent = binary.BigEndian.Uint16(os2[76:78])

		os2Version := binary.BigEndian.Uint16(os2[0:2])
		if os2Version >= 2 && len(os2) >= 90 {
			f.HasCapHeight = true
			f.XHeight = int16(binary.BigEndian.Uint16(os2[86:88]))
			f.CapHeight = int16(binary.BigEndian.Uint16(os2[88:90]))
		}
	}

	return f, nil
}
