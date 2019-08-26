package version

// Versions is the collection of version.
type Versions []Version

// Len is the number of elements in the collection.
func (versions Versions) Len() int {
	return len(versions)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (versions Versions) Less(i, j int) bool {
	vA, vB := versions[i].Byte(), versions[j].Byte()
	if vA>>8 == vB>>8 {
		switch {
		case versions[i][Pre] == 0:
			return false
		case versions[j][Pre] == 0:
			return true
		}
	}

	return vA <= vB
}

// Swap swaps the elements with indexes i and j.
func (versions Versions) Swap(i, j int) {
	versions[i], versions[j] = versions[j], versions[i]
}
