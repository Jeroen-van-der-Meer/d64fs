// Unified CLI for Commodore storage format inspection.
//
// Usage:
//
//	cbm info <file>              Format, disk/tape name, ID, free sectors
//	cbm ls <file>                Directory listing
//	cbm cat <file> <name>        Write file contents to stdout
//	cbm xxd <file> <name>        Hex dump of file contents
//	cbm extract <file> [dir]     Extract all files to directory
package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"cbm"
	_ "cbm/cbmdos"
	//_ "cbm/prg"
	//_ "cbm/t64"
	//_ "cbm/tap"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case "info":
		requireArgs(3)
		cmdInfo(os.Args[2])
	case "ls":
		requireArgs(3)
		cmdLs(os.Args[2])
	case "cat":
		requireArgs(4)
		cmdCat(os.Args[2], os.Args[3])
	case "xxd":
		requireArgs(4)
		cmdXxd(os.Args[2], os.Args[3])
	case "extract":
		dir := "."
		if len(os.Args) >= 4 {
			dir = os.Args[3]
		}
		requireArgs(3)
		cmdExtract(os.Args[2], dir)
	default:
		usage()
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: cbm <command> [args...]

Commands:
  info <file>              Show format, disk/tape name, ID, free sectors
  ls <file>                List directory
  cat <file> <name>        Print file contents to stdout
  xxd <file> <name>        Hex dump a file
  extract <file> [dir]     Extract all files to directory (default: .)
`)
	os.Exit(1)
}

func requireArgs(min int) {
	if len(os.Args) < min {
		usage()
	}
}

func openArchive(path string) cbm.Archive {
	f, err := os.Open(path)
	if err != nil {
		fatal(err)
	}
	defer f.Close()

	a, err := cbm.Open(f)
	if err != nil {
		fatal(err)
	}
	return a
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}

func cmdInfo(path string) {
	a := openArchive(path)
	info := a.Info()

	fmt.Printf("Format: %s\n", info.Format)
	if info.Name != "" {
		fmt.Printf("Name:   %s\n", info.Name)
	}
	if info.ID != "" {
		fmt.Printf("ID:     %s\n", info.ID)
	}
	if info.DOSType != "" {
		fmt.Printf("DOS:    %s\n", info.DOSType)
	}
	if info.Free >= 0 {
		fmt.Printf("Free:   %d sectors\n", info.Free)
	}

	entries, err := a.Entries()
	if err != nil {
		fatal(err)
	}
	fmt.Printf("Files:  %d\n", len(entries))
}

func cmdLs(path string) {
	a := openArchive(path)
	info := a.Info()

	fmt.Printf("Disk:  %s\n", info.Name)
	if info.ID != "" {
		fmt.Printf("ID:    %s\n", info.ID)
	}
	if info.DOSType != "" {
		fmt.Printf("DOS:   %s\n", info.DOSType)
	}
	if info.Free >= 0 {
		fmt.Printf("Free:  %d sectors\n", info.Free)
	}
	fmt.Println()

	entries, err := a.Entries()
	if err != nil {
		fatal(err)
	}

	if len(entries) == 0 {
		fmt.Println("(no files)")
		return
	}

	fmt.Printf("%-16s  %4s  %-3s  %s\n", "NAME            ", "SIZE", "TYP", "FLGS")
	fmt.Printf("%-16s  %4s  %-3s  %s\n", "----------------", "----", "---", "----")
	for _, e := range entries {
		size := e.Size / 254 // approximate sector count
		if e.Size < 0 {
			size = 0
		}
		var flags string
		if e.Locked {
			flags += "<"
		}
		if !e.Closed {
			flags += "*"
		}
		fmt.Printf("%-16s  %4d  %-3s  %s\n", e.Name, size, e.Type, flags)
	}
}

func cmdCat(path, name string) {
	a := openArchive(path)
	name = strings.ToLower(name)

	f, err := a.Open(name)
	if err != nil {
		fatal(err)
	}
	defer f.Close()

	data, err := readAll(f)
	if err != nil {
		fatal(err)
	}
	os.Stdout.Write(data)
}

func cmdXxd(path, name string) {
	a := openArchive(path)
	name = strings.ToLower(name)

	f, err := a.Open(name)
	if err != nil {
		fatal(err)
	}
	defer f.Close()

	data, err := readAll(f)
	if err != nil {
		fatal(err)
	}
	hexdump(data)
}

func cmdExtract(path, dir string) {
	a := openArchive(path)

	entries, err := a.Entries()
	if err != nil {
		fatal(err)
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		fatal(err)
	}

	for _, e := range entries {
		name := strings.ToLower(strings.TrimRight(e.Name, "\u00a0 "))
		name = strings.ReplaceAll(name, "\u00a0", " ")
		// Sanitize filename
		name = strings.ReplaceAll(name, "/", "_")
		if name == "" || name == "." || name == ".." {
			name = "unnamed"
		}

		f, err := a.Open(strings.ToLower(e.Name))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %s: %v\n", e.Name, err)
			continue
		}
		data, err := readAll(f)
		f.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %s: %v\n", e.Name, err)
			continue
		}

		outPath := filepath.Join(dir, name)
		if err := os.WriteFile(outPath, data, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %s: %v\n", outPath, err)
			continue
		}
		fmt.Printf("  %s (%d bytes)\n", name, len(data))
	}
}

func readAll(f fs.File) ([]byte, error) {
	var data []byte
	buf := make([]byte, 4096)
	for {
		n, err := f.Read(buf)
		data = append(data, buf[:n]...)
		if err != nil {
			if err == io.EOF {
				break
			}
			return data, err
		}
	}
	return data, nil
}

func hexdump(data []byte) {
	for off := 0; off < len(data); off += 16 {
		fmt.Printf("%08x: ", off)

		end := off + 16
		if end > len(data) {
			end = len(data)
		}
		for i := off; i < off+16; i++ {
			if i < end {
				fmt.Printf("%02x", data[i])
			} else {
				fmt.Print("  ")
			}
			if i%2 == 1 {
				fmt.Print(" ")
			}
		}

		fmt.Print(" ")
		for i := off; i < off+16; i++ {
			if i < end {
				b := data[i]
				if b >= 0x20 && b <= 0x7e {
					fmt.Printf("%c", b)
				} else {
					fmt.Print(".")
				}
			}
		}
		fmt.Println()
	}
}
