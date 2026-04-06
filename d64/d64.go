package d64

// Package d64: Read-only parser for Commodore 64 D64 disk images.
//
// A D64 file is a raw byte dump of a 1541 floppy disk: 35 tracks, 683 sectors
// of 256 bytes each, totaling 174,848 bytes. This package parses the CBM DOS
// filesystem stored within.

import (
	"fmt"
	"io"
	"os"
)

const (
	SectorCount = 683
	SectorSize  = 256
	TrackCount  = 35

	// Standard D64 image sizes.
	ImageSize           = SectorCount * SectorSize       // 174848
	ImageSizeWithErrors = SectorCount * (SectorSize + 1) // 175531

	// Track 18 holds the BAM (sector 0) and directory (sectors 1+).
	DirTrack  = 18
	BAMSector = 0
	DirSector = 1
)

// Map each track (1-35) to the number of sectors it contains.
var sectorsPerTrack = [TrackCount + 1]int{
	0, // 0 (tracks are 1-based)
	21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, // 1-17
	19, 19, 19, 19, 19, 19, 19, // 18-24
	18, 18, 18, 18, 18, 18, // 25-30
	17, 17, 17, 17, 17, // 31-35
}

// Store the byte offset of the first sector of each track.
var trackOffset [TrackCount + 1]int

func init() {
	for t := 1; t <= TrackCount; t++ {
		trackOffset[t] = trackOffset[t-1] + sectorsPerTrack[t-1]*SectorSize
	}
}

// Byte offset into a D64 image for a given (track, sector).
func sectorOffset(track, sector int) int {
	return trackOffset[track] + sector*SectorSize
}

// Opened D64 disk image.
type Disk struct {
	Verbose bool
	data    []byte
}

// warn prints a message to stderr when Verbose is enabled.
func (d *Disk) warn(format string, args ...any) {
	if d.Verbose {
		fmt.Fprintf(os.Stderr, "d64: "+format+"\n", args...)
	}
}

// Read an entire D64 image.
func Open(r io.Reader) (*Disk, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("Failed to read D64 image: %w", err)
	}
	if len(data) != ImageSize && len(data) != ImageSizeWithErrors {
		return nil, fmt.Errorf("Unexpected D64 image size: %d bytes (expected %d or %d)",
			len(data), ImageSize, ImageSizeWithErrors)
	}
	return &Disk{data: data}, nil
}

// Return the 256 bytes at the given (track, sector).
func (d *Disk) ReadSector(track, sector int) []byte {
	off := sectorOffset(track, sector)
	buf := make([]byte, SectorSize)
	copy(buf, d.data[off:(off+SectorSize)])
	return buf
}
