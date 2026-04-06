package d64

// CBM DOS file type stored in a directory entry.
type FileType byte

const (
	FileTypeDEL FileType = 0x00
	FileTypeSEQ FileType = 0x01
	FileTypePRG FileType = 0x02
	FileTypeUSR FileType = 0x03
	FileTypeREL FileType = 0x04
)

// Return the 3-letter file type name as shown in a C64 directory listing.
func (ft FileType) String() string {
	switch ft {
	case FileTypeDEL:
		return "DEL"
	case FileTypeSEQ:
		return "SEQ"
	case FileTypePRG:
		return "PRG"
	case FileTypeUSR:
		return "USR"
	case FileTypeREL:
		return "REL"
	default:
		return "???"
	}
}

// Single file entry in the disk directory.
type DirEntry struct {
	FileType      FileType
	Closed        bool // True if the file was properly closed
	Locked        bool // True if the file is locked
	Filename      string
	StartTrack    int
	StartSector   int
	SizeInSectors int
}

const dirEntrySize = 32

// Track 18 has 19 sectors of which the first is the BAM.
var maxDirSectors = sectorsPerTrack[DirTrack] - 1

// Reads all file entries from the disk's directory by following the sector
// chain starting at track 18, sector 1.
func (d *Disk) Directory() ([]DirEntry, error) {
	var entries []DirEntry

	track, sector := DirTrack, DirSector

	for i := 0; i < maxDirSectors; i++ {
		raw := d.ReadSector(track, sector)

		// Each sector holds 8 directory entries of 32 bytes each.
		for off := 0; off < SectorSize; off += dirEntrySize {
			entry := parseDirEntry(raw[off:(off + dirEntrySize)])
			if entry != nil {
				entries = append(entries, *entry)
			}
		}

		// Bytes 0–1 of the sector are the chain pointer to the next directory
		// sector.
		nextTrack := int(raw[0])
		nextSector := int(raw[1])
		// A track value of 0 marks the end of the chain.
		if nextTrack == 0 {
			break
		}
		track, sector = nextTrack, nextSector
	}

	return entries, nil
}

// Parse a 32-byte directory entry.
//
// Layout:
//
// 0x00,0x01: Next dir track/sector
// 0x02:      Type byte.
// 0x03:      Track of first data sector.
// 0x04:      Sector of first data sector.
// 0x05-0x14: Filename.
// 0x15,0x16: First side-sector track/sector (REL files only).
// 0x17:      REL record length.
// 0x18-0x1D: Unused.
// 0x1E,0x1F: File size in sectors (little-endian).
func parseDirEntry(raw []byte) *DirEntry {
	typeByte := raw[0x02]
	if typeByte == 0x00 {
		return nil // scratched/empty entry
	}

	return &DirEntry{
		FileType:      FileType(typeByte & 0x07),
		Closed:        typeByte&0x80 != 0,
		Locked:        typeByte&0x40 != 0,
		Filename:      DecodePETSCII(raw[0x05:0x15]),
		StartTrack:    int(raw[0x03]),
		StartSector:   int(raw[0x04]),
		SizeInSectors: int(raw[0x1E]) | int(raw[0x1F])<<8,
	}
}
