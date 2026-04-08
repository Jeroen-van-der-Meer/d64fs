package disk

// d64Sectors returns sectors per track for 1541/1571 density zones.
// Tracks 1-17 have 21 sectors, 18-24 have 19, 25-30 have 18, 31+ have 17.
func d64Sectors(track int) int {
	switch {
	case track <= 17:
		return 21
	case track <= 24:
		return 19
	case track <= 30:
		return 18
	default:
		return 17
	}
}

func newD64Geometry(trackCount int) *geometry {
	f := D64
	if trackCount > 35 {
		f = D64Ext
	}
	g := &geometry{
		format:     f,
		trackCount: trackCount,
	}
	g.spt = make([]int, trackCount+1)
	for t := 1; t <= trackCount; t++ {
		g.spt[t] = d64Sectors(t)
	}
	g.computeOffsets()
	return g
}

func newD71Geometry() *geometry {
	g := &geometry{
		format:     D71,
		trackCount: 70,
	}
	g.spt = make([]int, 71)
	for t := 1; t <= 35; t++ {
		g.spt[t] = d64Sectors(t)
		g.spt[t+35] = d64Sectors(t) // side 2 mirrors side 1
	}
	g.computeOffsets()
	return g
}
