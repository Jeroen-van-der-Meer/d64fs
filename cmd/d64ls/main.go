// List the directory of a Commodore 64 D64 disk image.
//
// Usage: d64ls <image.d64>

package main

import (
	"fmt"
	"os"

	"d64/d64"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <image.d64>\n", os.Args[0])
		os.Exit(1)
	}

	f, err := os.Open(os.Args[1])
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

	bam, err := disk.BAM()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading BAM: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Disk:  %s\n", bam.DiskName)
	fmt.Printf("ID:    %s\n", bam.DiskID)
	fmt.Printf("DOS:   %s\n", bam.DOSType)
	fmt.Printf("Free:  %d sectors\n", bam.FreeSectors())
	fmt.Println()

	entries, err := disk.Directory()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading directory: %v\n", err)
		os.Exit(1)
	}

	if len(entries) == 0 {
		fmt.Println("(no files)")
		return
	}

	fmt.Printf("%-16s  %4s  %-3s  %s\n", "NAME            ", "SIZE", "TYP", "FLGS")
	fmt.Printf("%-16s  %4s  %-3s  %s\n", "----------------", "----", "---", "----")
	for _, e := range entries {
		var flags string
		if e.Locked {
			flags += "<"
		}
		if !e.Closed {
			flags += "*"
		}
		fmt.Printf("%-16s  %4d  %-3s  %s\n", e.Filename, e.SizeInSectors, e.FileType, flags)
	}
}
