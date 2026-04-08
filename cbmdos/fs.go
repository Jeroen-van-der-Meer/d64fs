// Package cbmdos implements the CBM DOS filesystem found on Commodore disk
// images (D64, D71, D81). It consumes a disk.Image for sector access and
// implements the cbm.Archive interface.
//
// Import this package for its side effect of registering CBM DOS with cbm.Open:
//
//	import _ "cbm/cbmdos"
package cbmdos

import (
	"fmt"
	"io"
	"io/fs"
	"strings"
	"time"

	"cbm"
	"cbm/disk"
)

func init() {
	cbm.RegisterFormat("cbmdos", sniff, open)
}

func sniff(data []byte) bool {
	img, err := disk.Open(data)
	if err != nil {
		return false
	}
	return validateDOS(img) == nil
}

func open(data []byte) (cbm.Archive, error) {
	img, err := disk.Open(data)
	if err != nil {
		return nil, err
	}
	return NewFS(img)
}

// validateDOS checks the DOS version byte at the expected BAM location.
func validateDOS(img disk.Image) error {
	var track int
	var want byte
	switch img.Format() {
	case disk.D64, disk.D64Ext, disk.D71:
		track = 18
		want = 0x41 // 'A'
	case disk.D81:
		track = 40
		want = 0x44 // 'D'
	case disk.D80, disk.D82:
		track = 39
		want = 0x43 // 'C'
	default:
		return fmt.Errorf("unsupported disk format: %s", img.Format())
	}
	raw, err := img.ReadSector(track, 0)
	if err != nil {
		return err
	}
	if got := raw[0x02]; got != want {
		return fmt.Errorf("invalid DOS version byte: 0x%02x (expected 0x%02x for %s)",
			got, want, img.Format())
	}
	return nil
}

// DirTrack returns the directory/BAM track for the given format.
func DirTrack(f disk.Format) int {
	switch f {
	case disk.D81:
		return 40
	case disk.D80, disk.D82:
		return 39
	default:
		return 18
	}
}

// dirSector returns the first directory sector for the given format.
func dirSector(f disk.Format) int {
	switch f {
	case disk.D81:
		return 3
	default:
		return 1
	}
}

// FS is a CBM DOS filesystem backed by a disk image.
type FS struct {
	Img     disk.Image
	entries []dirEntry
	info    cbm.MediaInfo
}

// NewFS creates a CBM DOS filesystem from a disk image.
func NewFS(img disk.Image) (*FS, error) {
	if err := validateDOS(img); err != nil {
		return nil, err
	}
	f := &FS{Img: img}
	if err := f.readBAM(); err != nil {
		return nil, fmt.Errorf("reading BAM: %w", err)
	}
	entries, err := f.readDirectory()
	if err != nil {
		return nil, fmt.Errorf("reading directory: %w", err)
	}
	f.entries = entries
	return f, nil
}

// Info returns media information.
func (f *FS) Info() cbm.MediaInfo { return f.info }

// Entries returns the directory listing.
func (f *FS) Entries() ([]cbm.Entry, error) {
	out := make([]cbm.Entry, len(f.entries))
	for i, e := range f.entries {
		out[i] = cbm.Entry{
			Name:   normalizeFilename(e.filename),
			Size:   e.sizeInSectors * 254,
			Type:   e.fileType.String(),
			Closed: e.closed,
			Locked: e.locked,
		}
	}
	return out, nil
}

// --- fs.FS implementation ---

// Open implements fs.FS. It accepts "." for the root directory or a filename
// (case-insensitive, no leading slash).
func (f *FS) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	if name == "." {
		return &rootDir{fsys: f}, nil
	}
	// CBM DOS has a flat directory — no subdirectories.
	if strings.Contains(name, "/") {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	for _, e := range f.entries {
		if normalizeFilename(e.filename) == name {
			data, err := f.readFile(e)
			if err != nil {
				return nil, err
			}
			return &file{entry: e, data: data}, nil
		}
	}
	return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
}

func normalizeFilename(name string) string {
	name = strings.ToLower(name)
	name = strings.TrimRight(name, "\u00a0 ")
	return strings.ReplaceAll(name, "\u00a0", " ")
}

// --- file: implements fs.File ---

type file struct {
	entry  dirEntry
	data   []byte
	offset int
	closed bool
}

func (f *file) Stat() (fs.FileInfo, error) {
	return &fileInfo{entry: f.entry, size: int64(len(f.data))}, nil
}

func (f *file) Read(buf []byte) (int, error) {
	if f.closed {
		return 0, fs.ErrClosed
	}
	if f.offset >= len(f.data) {
		return 0, io.EOF
	}
	n := copy(buf, f.data[f.offset:])
	f.offset += n
	return n, nil
}

func (f *file) Close() error {
	f.closed = true
	return nil
}

// --- rootDir: implements fs.ReadDirFile ---

type rootDir struct {
	fsys   *FS
	offset int
	closed bool
}

func (d *rootDir) Stat() (fs.FileInfo, error) {
	return &fileInfo{isDir: true}, nil
}

func (d *rootDir) Read([]byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: ".", Err: fs.ErrInvalid}
}

func (d *rootDir) Close() error {
	d.closed = true
	return nil
}

func (d *rootDir) ReadDir(n int) ([]fs.DirEntry, error) {
	if d.closed {
		return nil, fs.ErrClosed
	}
	remaining := d.fsys.entries[d.offset:]
	if n <= 0 {
		d.offset = len(d.fsys.entries)
		result := make([]fs.DirEntry, len(remaining))
		for i, e := range remaining {
			result[i] = &fileInfo{entry: e, size: int64(e.sizeInSectors) * 254}
		}
		return result, nil
	}
	if len(remaining) == 0 {
		return nil, io.EOF
	}
	if n > len(remaining) {
		n = len(remaining)
	}
	d.offset += n
	result := make([]fs.DirEntry, n)
	for i, e := range remaining[:n] {
		result[i] = &fileInfo{entry: e, size: int64(e.sizeInSectors) * 254}
	}
	if d.offset >= len(d.fsys.entries) {
		return result, io.EOF
	}
	return result, nil
}

// --- fileInfo: implements both fs.FileInfo and fs.DirEntry ---

type fileInfo struct {
	entry dirEntry
	size  int64
	isDir bool
}

func (i *fileInfo) Name() string {
	if i.isDir {
		return "."
	}
	return normalizeFilename(i.entry.filename)
}
func (i *fileInfo) Size() int64 { return i.size }
func (i *fileInfo) Mode() fs.FileMode {
	if i.isDir {
		return fs.ModeDir | 0o555
	}
	return 0o444
}
func (i *fileInfo) ModTime() time.Time       { return time.Time{} }
func (i *fileInfo) IsDir() bool              { return i.isDir }
func (i *fileInfo) Sys() any                 { return nil }
func (i *fileInfo) Type() fs.FileMode        { return i.Mode().Type() }
func (i *fileInfo) Info() (fs.FileInfo, error) { return i, nil }

// Compile-time interface checks.
var (
	_ cbm.Archive    = (*FS)(nil)
	_ fs.FS          = (*FS)(nil)
	_ fs.File        = (*file)(nil)
	_ fs.ReadDirFile = (*rootDir)(nil)
	_ fs.DirEntry    = (*fileInfo)(nil)
	_ fs.FileInfo    = (*fileInfo)(nil)
)
