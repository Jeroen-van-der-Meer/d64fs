package disk

// D80: Commodore 8050, 77 tracks, single-sided 8" floppy.
// D82: Commodore 8250, 154 tracks (77 per side), double-sided.
//
// Density zones (per side):
//   Tracks  1-39: 29 sectors
//   Tracks 40-53: 27 sectors
//   Tracks 54-64: 25 sectors
//   Tracks 65-77: 23 sectors

func d80Sectors(track int) int {
	switch {
	case track <= 39:
		return 29
	case track <= 53:
		return 27
	case track <= 64:
		return 25
	default:
		return 23
	}
}

func newD80Geometry() *geometry {
	g := &geometry{
		format:     D80,
		trackCount: 77,
	}
	g.spt = make([]int, 78)
	for t := 1; t <= 77; t++ {
		g.spt[t] = d80Sectors(t)
	}
	g.computeOffsets()
	return g
}

func newD82Geometry() *geometry {
	g := &geometry{
		format:     D82,
		trackCount: 154,
	}
	g.spt = make([]int, 155)
	for t := 1; t <= 77; t++ {
		g.spt[t] = d80Sectors(t)
		g.spt[t+77] = d80Sectors(t) // side 2 mirrors side 1
	}
	g.computeOffsets()
	return g
}
