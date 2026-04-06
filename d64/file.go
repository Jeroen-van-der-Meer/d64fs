package d64

// Read the full contents of a file by following its sector chain.
func (d *Disk) ReadFile(entry DirEntry) ([]byte, error) {
	if entry.StartTrack == 0 {
		return nil, nil // empty file
	}

	var data []byte
	visited := make(map[[2]int]bool)

	track, sector := entry.StartTrack, entry.StartSector
	sectors := 0

	for i := 0; i < SectorCount; i++ {
		key := [2]int{track, sector}
		if visited[key] {
			d.warn("%s: cycle in sector chain at (%d, %d), truncating after %d sectors",
				entry.Filename, track, sector, sectors)
			break
		}
		visited[key] = true

		raw := d.ReadSector(track, sector)
		sectors++

		nextTrack := int(raw[0])
		nextSector := int(raw[1])

		if nextTrack > 0 {
			data = append(data, raw[2:]...)
			track, sector = nextTrack, nextSector
		} else {
			// If nextTrack == 0, we've reached the last sector, and nextSector
			// indicates the number of bytes remaining.
			data = append(data, raw[2:nextSector]...)
			break
		}
	}

	if sectors != entry.SizeInSectors {
		d.warn("%s: sector count mismatch (directory says %d, chain has %d)",
			entry.Filename, entry.SizeInSectors, sectors)
	}

	return data, nil
}
