package cbmdos

import (
	"cbm"
	"cbm/disk"
)

// bamEntry is a BAM record for a single track.
type bamEntry struct {
	freeSectors int
	bitmap      []byte
}

// parseBAMEntries reads count BAM entries from data starting at offset.
// Each entry is entrySize bytes: 1 free-count byte followed by bitmap bytes.
func parseBAMEntries(data []byte, offset, count, entrySize int) []bamEntry {
	entries := make([]bamEntry, count)
	for i := range count {
		off := offset + i*entrySize
		bitmap := make([]byte, entrySize-1)
		copy(bitmap, data[off+1:off+entrySize])
		entries[i] = bamEntry{
			freeSectors: int(data[off]),
			bitmap:      bitmap,
		}
	}
	return entries
}

// readBAM reads the BAM and populates f.info.
func (f *FS) readBAM() error {
	switch f.Img.Format() {
	case disk.D81:
		return f.readBAMD81()
	case disk.D80, disk.D82:
		return f.readBAMD80()
	default:
		return f.readBAMD64()
	}
}

func (f *FS) readBAMD64() error {
	raw, err := f.Img.ReadSector(DirTrack(f.Img.Format()), 0)
	if err != nil {
		return err
	}

	f.info = cbm.MediaInfo{
		Format:  f.Img.Format().String(),
		Name:    cbm.DecodePETSCII(raw[0x90:0xA0]),
		ID:      cbm.DecodePETSCII(raw[0xA2:0xA4]),
		DOSType: cbm.DecodePETSCII(raw[0xA5:0xA7]),
	}

	tc := f.Img.Tracks()
	limit := min(tc, 35)
	entries := parseBAMEntries(raw, 0x04, limit, 4)

	free := 0
	for _, e := range entries {
		free += e.freeSectors
	}

	// Extended D64: tracks 36-40 at offsets 0xAC-0xBF (SpeedDOS format).
	if f.Img.Format() == disk.D64Ext {
		extCount := min(tc, 40) - 35
		for _, e := range parseBAMEntries(raw, 0xAC, extCount, 4) {
			free += e.freeSectors
		}
	}

	// D71: Side 2 (tracks 36-70).
	// Free sector counts at 18/0 offsets 0xDD-0xFF (1 byte each).
	if f.Img.Format() == disk.D71 {
		for t := 36; t <= 70; t++ {
			free += int(raw[0xDD+(t-36)])
		}
	}

	f.info.Free = free
	return nil
}

func (f *FS) readBAMD81() error {
	header, err := f.Img.ReadSector(40, 0)
	if err != nil {
		return err
	}

	f.info = cbm.MediaInfo{
		Format:  f.Img.Format().String(),
		Name:    cbm.DecodePETSCII(header[0x04:0x14]),
		ID:      cbm.DecodePETSCII(header[0x16:0x18]),
		DOSType: cbm.DecodePETSCII(header[0x19:0x1B]),
	}

	bam1, err := f.Img.ReadSector(40, 1)
	if err != nil {
		return err
	}
	bam2, err := f.Img.ReadSector(40, 2)
	if err != nil {
		return err
	}

	free := 0
	for _, e := range parseBAMEntries(bam1, 0x10, 40, 6) {
		free += e.freeSectors
	}
	for _, e := range parseBAMEntries(bam2, 0x10, 40, 6) {
		free += e.freeSectors
	}

	f.info.Free = free
	return nil
}

// readBAMD80 parses BAM for D80 (8050) and D82 (8250) images.
// The BAM header is at track 39, sector 0. BAM entries are 5 bytes each
// (1 free count + 4 bitmap bytes) in sectors 39/1-2 (D80) or 39/1-2 and
// 39/4-5 (D82).
func (f *FS) readBAMD80() error {
	raw, err := f.Img.ReadSector(39, 0)
	if err != nil {
		return err
	}

	f.info = cbm.MediaInfo{
		Format:  f.Img.Format().String(),
		Name:    cbm.DecodePETSCII(raw[0x06:0x16]),
		ID:      cbm.DecodePETSCII(raw[0x18:0x1A]),
		DOSType: cbm.DecodePETSCII(raw[0x1B:0x1D]),
	}

	type bamRange struct {
		sector, start, end int
	}
	ranges := []bamRange{
		{1, 1, 50},
		{2, 51, 77},
	}
	if f.Img.Format() == disk.D82 {
		ranges = append(ranges,
			bamRange{4, 78, 128},
			bamRange{5, 129, 154},
		)
	}

	free := 0
	for _, br := range ranges {
		bamData, err := f.Img.ReadSector(39, br.sector)
		if err != nil {
			return err
		}
		count := br.end - br.start + 1
		for _, e := range parseBAMEntries(bamData, 0x06, count, 5) {
			free += e.freeSectors
		}
	}

	f.info.Free = free
	return nil
}
