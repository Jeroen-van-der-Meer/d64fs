package d64

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAllImages(t *testing.T) {
	paths, err := filepath.Glob("testdata/*.d64")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("No D64 images found in testdata/")
	}

	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			f, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			d, err := Open(f)
			if err != nil {
				t.Fatalf("Open: %v", err)
			}

			entries, err := d.Directory()
			if err != nil {
				t.Fatalf("Directory: %v", err)
			}
			t.Logf("%d files", len(entries))

			for _, e := range entries {
				data, err := d.ReadFile(e)
				if err != nil {
					t.Errorf("ReadFile(%q): %v", e.Filename, err)
					continue
				}
				t.Logf("  %s (%d bytes)", e.Filename, len(data))
			}

			// fs.FS interface.
			//err = fs.WalkDir(d, ".", func(path string, _ fs.DirEntry, err error) error {
			//	return err
			//})
			//if err != nil {
			//	t.Errorf("WalkDir: %v", err)
			//}
		})
	}
}
