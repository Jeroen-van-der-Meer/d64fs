package cbm_test

import (
	"os"
	"path/filepath"
	"testing"

	"cbm"
	_ "cbm/cbmdos"
)

func TestAllImages(t *testing.T) {
	var paths []string
	for _, ext := range []string{"*.d64", "*.d71", "*.d81"} {
		matches, err := filepath.Glob("testdata/" + ext)
		if err != nil {
			t.Fatal(err)
		}
		paths = append(paths, matches...)
	}
	if len(paths) == 0 {
		t.Fatal("No disk images found in testdata/")
	}

	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			f, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			a, err := cbm.Open(f)
			if err != nil {
				t.Fatalf("Open: %v", err)
			}

			info := a.Info()
			t.Logf("Format: %s  Name: %s  ID: %s  DOS: %s  Free: %d",
				info.Format, info.Name, info.ID, info.DOSType, info.Free)

			entries, err := a.Entries()
			if err != nil {
				t.Fatalf("Entries: %v", err)
			}
			t.Logf("%d files", len(entries))

			for _, e := range entries {
				f, err := a.Open(e.Name)
				if err != nil {
					t.Logf("Open(%q): %v", e.Name, err)
					continue
				}
				fi, err := f.Stat()
				if err != nil {
					t.Errorf("Stat(%q): %v", e.Name, err)
				} else {
					t.Logf("  %s (%d bytes)", e.Name, fi.Size())
				}
				f.Close()
			}
		})
	}
}
