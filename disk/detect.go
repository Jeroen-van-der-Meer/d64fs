package disk

// detectGeometry returns the geometry matching the given image size,
// or nil if unrecognized.
func detectGeometry(size int) *geometry {
	// Try known formats: standard D64, 40-track D64, D71, D81.
	known := []*geometry{
		newD64Geometry(35),
		newD64Geometry(40),
		newD71Geometry(),
		newD81Geometry(),
		newD80Geometry(),
		newD82Geometry(),
	}
	for _, g := range known {
		if size == g.imageSize || size == g.imageSizeErr {
			return g
		}
	}

	// Fallback: D64 with nonstandard track counts (36-53).
	for tracks := 36; tracks <= 53; tracks++ {
		if tracks == 40 {
			continue // already tried above
		}
		g := newD64Geometry(tracks)
		if size == g.imageSize || size == g.imageSizeErr {
			return g
		}
	}

	return nil
}
