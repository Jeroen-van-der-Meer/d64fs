package disk

// geometry holds computed disk layout for format detection.
type geometry struct {
	format       Format
	trackCount   int
	totalSectors int
	imageSize    int
	imageSizeErr int // size with per-sector error bytes
	spt          []int
	offsets      []int
}

// computeOffsets fills in offsets, totalSectors, imageSize, and imageSizeErr.
func (g *geometry) computeOffsets() {
	g.offsets = make([]int, g.trackCount+1)
	offset := 0
	for t := 1; t <= g.trackCount; t++ {
		g.offsets[t] = offset
		offset += g.spt[t] * SectorSize
	}
	g.totalSectors = offset / SectorSize
	g.imageSize = offset
	g.imageSizeErr = g.totalSectors * (SectorSize + 1)
}
