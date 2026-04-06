// Hex-dump a file from a Commodore 64 D64 disk image.
//
// Usage: d64xxd [-v] <image.d64> <filename>

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"d64/d64"
)

func main() {
	verbose := flag.Bool("v", false, "Verbose logging")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-v] <image.d64> <filename>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}

	f, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	disk, err := d64.Open(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	disk.Verbose = *verbose

	entries, err := disk.Directory()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading directory: %v\n", err)
		os.Exit(1)
	}

	name := strings.ToLower(flag.Arg(1))
	var match *d64.DirEntry
	for _, e := range entries {
		normalized := strings.ToLower(strings.TrimRight(e.Filename, "\u00a0 "))
		normalized = strings.ReplaceAll(normalized, "\u00a0", " ")
		if normalized == name {
			match = &e
			break
		}
	}
	if match == nil {
		fmt.Fprintf(os.Stderr, "File %q not found on disk\n", flag.Arg(1))
		os.Exit(1)
	}

	data, err := disk.ReadFile(*match)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	hexdump(data)
}

func hexdump(data []byte) {
	for off := 0; off < len(data); off += 16 {
		// Offset
		fmt.Printf("%08x: ", off)

		// Hex bytes in pairs
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

		// ASCII
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
