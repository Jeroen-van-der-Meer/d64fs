// Package disk provides sector-level access to Commodore disk images.
//
// This package handles container geometry only — it knows nothing about
// filesystems. Filesystem interpretation (CBM DOS, CP/M, etc.) is handled
// by separate packages that consume the Image interface.
package disk

import "fmt"

const SectorSize = 256

// Format identifies the disk image format.
type Format int

const (
	D64    Format = iota // Standard 35-track 1541
	D64Ext               // Extended D64 (36-53 tracks)
	D71                  // Double-sided 1571
	D81                  // 1581 3.5"
	D80                  // 8050 single-sided 8"
	D82                  // 8250 double-sided 8"
)

func (f Format) String() string {
	switch f {
	case D64:
		return "D64"
	case D64Ext:
		return "D64 (extended)"
	case D71:
		return "D71"
	case D81:
		return "D81"
	case D80:
		return "D80"
	case D82:
		return "D82"
	default:
		return "Unknown"
	}
}

// Image provides read-only sector-level access to a disk image.
type Image interface {
	Format() Format
	Tracks() int
	SectorsIn(track int) int
	TotalSectors() int
	ReadSector(track, sector int) ([]byte, error)
	Valid(track, sector int) bool
}

// image is the concrete implementation backed by a byte slice.
type image struct {
	format       Format
	data         []byte
	trackCount   int
	totalSectors int
	spt          []int // sectors per track, indexed by track (0 unused)
	offsets      []int // byte offset of each track's first sector
}

func (img *image) Format() Format    { return img.format }
func (img *image) Tracks() int       { return img.trackCount }
func (img *image) TotalSectors() int { return img.totalSectors }

func (img *image) SectorsIn(track int) int {
	if track < 1 || track > img.trackCount {
		return 0
	}
	return img.spt[track]
}

func (img *image) Valid(track, sector int) bool {
	return track >= 1 && track <= img.trackCount &&
		sector >= 0 && sector < img.spt[track]
}

func (img *image) ReadSector(track, sector int) ([]byte, error) {
	if !img.Valid(track, sector) {
		return nil, fmt.Errorf("invalid sector (%d, %d)", track, sector)
	}
	off := img.offsets[track] + sector*SectorSize
	buf := make([]byte, SectorSize)
	copy(buf, img.data[off:off+SectorSize])
	return buf, nil
}

// Open detects the disk format from the raw image data and returns an Image.
func Open(data []byte) (Image, error) {
	g := detectGeometry(len(data))
	if g == nil {
		return nil, fmt.Errorf("unrecognized disk image size: %d bytes", len(data))
	}
	return &image{
		format:       g.format,
		data:         data,
		trackCount:   g.trackCount,
		totalSectors: g.totalSectors,
		spt:          g.spt,
		offsets:      g.offsets,
	}, nil
}
