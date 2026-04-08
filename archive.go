// Package cbm provides read-only access to Commodore storage formats.
//
// Supported formats include disk images (D64, D71, D81), tape archives (T64),
// and raw program files (PRG). Format detection is automatic.
//
// The central abstraction is the Archive interface, which every format
// implements. Import format packages for their side effects to register
// them with Open:
//
//	import (
//		"cbm"
//		_ "cbm/cbmdos"
//	)
//
//	a, err := cbm.Open(f)
package cbm

import "io/fs"

// Archive is a read-only Commodore storage medium containing files.
type Archive interface {
	fs.FS
	Info() MediaInfo
	Entries() ([]Entry, error)
}

// MediaInfo describes the storage medium.
type MediaInfo struct {
	Format  string // "D64", "D71", "D81", "T64", "TAP", "PRG", etc.
	Name    string // Disk/tape name (empty for PRG)
	ID      string // Disk ID (empty for tape/PRG)
	DOSType string // e.g. "2A" for CBM DOS (empty if not applicable)
	Free    int    // Free sectors/blocks; -1 if not applicable
}

// Entry is a single file in an archive.
type Entry struct {
	Name     string
	Size     int    // Size in bytes; -1 if unknown before reading
	Type     string // "PRG", "SEQ", "USR", "REL", "DEL"
	Closed   bool
	Locked   bool
	LoadAddr uint16 // Load address (T64/PRG); 0 if unknown
}
