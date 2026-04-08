package cbmdos

import (
	"fmt"
	"os"

	"cbm"
	"cbm/disk"
)

// FileType is a CBM DOS file type stored in a directory entry.
type FileType byte

const (
	FileTypeDEL FileType = 0x00
	FileTypeSEQ FileType = 0x01
	FileTypePRG FileType = 0x02
	FileTypeUSR FileType = 0x03
	FileTypeREL FileType = 0x04
)

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

// dirEntry is an internal directory entry with CBM DOS-specific fields.
type dirEntry struct {
	fileType      FileType
	closed        bool
	locked        bool
	filename      string
	startTrack    int
	startSector   int
	sizeInSectors int
}

const dirEntrySize = 32

func (f *FS) readDirectory() ([]dirEntry, error) {
	var entries []dirEntry

	track := DirTrack(f.Img.Format())
	sector := dirSector(f.Img.Format())

	for i := 0; i < f.Img.TotalSectors(); i++ {
		raw, err := f.Img.ReadSector(track, sector)
		if err != nil {
			return nil, err
		}

		for off := 0; off < disk.SectorSize; off += dirEntrySize {
			entry := parseDirEntry(raw[off : off+dirEntrySize])
			if entry != nil {
				entries = append(entries, *entry)
			}
		}

		nextTrack := int(raw[0])
		nextSector := int(raw[1])
		if nextTrack == 0 {
			break
		}
		if !f.Img.Valid(nextTrack, nextSector) {
			fmt.Fprintf(os.Stderr, "Warning: directory chain points to invalid sector (%d, %d), truncating\n",
				nextTrack, nextSector)
			break
		}
		track, sector = nextTrack, nextSector
	}

	return entries, nil
}

func parseDirEntry(raw []byte) *dirEntry {
	typeByte := raw[0x02]
	if typeByte == 0x00 {
		return nil // scratched/empty entry
	}
	return &dirEntry{
		fileType:      FileType(typeByte & 0x07),
		closed:        typeByte&0x80 != 0,
		locked:        typeByte&0x40 != 0,
		filename:      cbm.DecodePETSCII(raw[0x05:0x15]),
		startTrack:    int(raw[0x03]),
		startSector:   int(raw[0x04]),
		sizeInSectors: int(raw[0x1E]) | int(raw[0x1F])<<8,
	}
}
