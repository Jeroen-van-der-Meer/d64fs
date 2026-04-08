package disk

func newD81Geometry() *geometry {
	g := &geometry{
		format:     D81,
		trackCount: 80,
	}
	g.spt = make([]int, 81)
	for t := 1; t <= 80; t++ {
		g.spt[t] = 40
	}
	g.computeOffsets()
	return g
}
