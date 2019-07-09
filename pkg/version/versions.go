package version

type Versions []Version

func (versions Versions) Len() int {
	return len(versions)
}

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

func (versions Versions) Swap(i, j int) {
	versions[i], versions[j] = versions[j], versions[i]
}
