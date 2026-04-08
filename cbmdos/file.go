package cbmdos

import (
	"fmt"
	"os"
)

func (f *FS) readFile(entry dirEntry) ([]byte, error) {
	if entry.startTrack == 0 {
		return nil, nil // empty file
	}
	if !f.Img.Valid(entry.startTrack, entry.startSector) {
		return nil, fmt.Errorf("%s: start sector (%d, %d) out of range",
			entry.filename, entry.startTrack, entry.startSector)
	}

	var data []byte
	visited := make(map[[2]int]bool)
	track, sector := entry.startTrack, entry.startSector
	sectors := 0

	for i := 0; i < f.Img.TotalSectors(); i++ {
		key := [2]int{track, sector}
		if visited[key] {
			fmt.Fprintf(os.Stderr, "Warning: %s: cycle in sector chain at (%d, %d), truncating after %d sectors\n",
				entry.filename, track, sector, sectors)
			break
		}
		visited[key] = true

		raw, err := f.Img.ReadSector(track, sector)
		if err != nil {
			return nil, err
		}
		sectors++

		nextTrack := int(raw[0])
		nextSector := int(raw[1])

		if nextTrack > 0 {
			if !f.Img.Valid(nextTrack, nextSector) {
				fmt.Fprintf(os.Stderr, "Warning: %s: chain points to invalid sector (%d, %d), truncating after %d sectors\n",
					entry.filename, nextTrack, nextSector, sectors)
				data = append(data, raw[2:]...)
				break
			}
			data = append(data, raw[2:]...)
			track, sector = nextTrack, nextSector
		} else {
			// nextTrack == 0: last sector. nextSector = bytes used in this sector.
			data = append(data, raw[2:nextSector]...)
			break
		}
	}

	if sectors != entry.sizeInSectors {
		fmt.Fprintf(os.Stderr, "Warning: %s: sector count mismatch (directory says %d, chain has %d)\n",
			entry.filename, entry.sizeInSectors, sectors)
	}

	return data, nil
}
