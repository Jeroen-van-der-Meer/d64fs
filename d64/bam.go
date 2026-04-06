package d64

// BAM: Block availability map stored at track 18, sector 0.
//
// The BAM tracks which sectors are free/used and stores the disk name and ID.
// Layout of track 18, sector 0:
//
// 0x00,0x01: Track/sector of first directory sector (usually 18/1)
// 0x02:      DOS version ('A' for standard CBM DOS 2.6)
// 0x03:      Unused
// 0x04-0x8F: BAM entries, 4 bytes per track (tracks 1-35 = 140 bytes)
// 0x90-0x9F: Disk name (16 bytes, padded with 0xA0)
// 0xA0,0xA1: Unused
// 0xA2,0xA3: Disk ID (2 bytes)
// 0xA4:      Unused
// 0xA5,0xA6: DOS type (usually "2A")

// Block availability map.
type BAM struct {
	DiskName string
	DiskID   string
	DOSType  string
	Entries  [TrackCount + 1]BAMEntry
}

// 4-byte BAM record for a single track.
type BAMEntry struct {
	FreeSectors int
	Bitmap      [3]byte // Sector allocation bitmap.
}

// Parse the BAM from the disk image.
func (d *Disk) BAM() (*BAM, error) {
	raw := d.ReadSector(DirTrack, BAMSector)

	b := &BAM{}

	// Parse BAM entries for each track: 4 bytes per track starting at offset 0x04.
	for t := 1; t <= TrackCount; t++ {
		off := 0x04 + (t-1)*4
		b.Entries[t] = BAMEntry{
			FreeSectors: int(raw[off]),
			Bitmap:      [3]byte{raw[off+1], raw[off+2], raw[off+3]},
		}
	}

	// Disk name: 16 bytes at offset 0x90, padded with 0xA0.
	b.DiskName = DecodePETSCII(raw[0x90:0xA0])

	// Disk ID: 2 bytes at offset 0xA2.
	b.DiskID = DecodePETSCII(raw[0xA2:0xA4])

	// DOS type: 2 bytes at offset 0xA5.
	b.DOSType = DecodePETSCII(raw[0xA5:0xA7])

	return b, nil
}

// Check whether the given sector on the given track is free.
func (b *BAM) IsFree(track, sector int) bool {
	// The bitmap is 3 bytes (24 bits). Bit N corresponds to sector N.
	byteIndex := sector / 8
	bitIndex := sector % 8
	return b.Entries[track].Bitmap[byteIndex]&(1<<bitIndex) != 0
}

// Total number of free sectors on the disk.
func (b *BAM) FreeSectors() int {
	total := 0
	for t := 1; t <= TrackCount; t++ {
		total += b.Entries[t].FreeSectors
	}
	return total
}
