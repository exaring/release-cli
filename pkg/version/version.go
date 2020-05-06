package version

import (
	"fmt"
	"regexp"
	"strconv"
)

const (
	// Major is the position in the slice of the major value.
	Major = iota
	// Minor is the position in the slice of the minor value.
	Minor
	// Patch is the position in the slice of the patch value.
	Patch
	// Pre is the position in the slice of the pre value.
	Pre

	// RegExPatternVersionString is the RegEX to parse the version string and detected the major, minor, patch and pre version.
	RegExPatternVersionString = `((\d+)\.(\d+)\.(\d+))(?:-RC([\dA-Za-z\-]+(?:\.[\dA-Za-z\-]+)*))?`
)

// Version is the abstraction of the version.
type Version []uint

// New creates an new instance of the version.
func New(v string) (Version, error) {
	r := regexp.
		MustCompile(RegExPatternVersionString).
		FindStringSubmatch(v)

	if len(r) == 0 {
		version := make(Version, 4)
		return version, nil
	}

	var version Version
	for i := 2; i < len(r); i++ {
		number, err := strconv.ParseInt(r[i], 10, 64)
		if err != nil {
			number = 0
		}

		version = append(version, uint(number))
	}

	return version, nil
}

// Increase increase the semantic version by the values of major, minor, patch and pre.
func (v Version) Increase(major, minor, patch, pre bool) {
	switch {
	case v.IsReleaseCandidate() && (major || minor || patch):
		v[Pre] = 0
		return
	case major:
		v.increaseVersion(Major)
	case minor:
		v.increaseVersion(Minor)
	case patch:
		v.increaseVersion(Patch)
	case pre:
		if !v.IsReleaseCandidate() {
			v.increaseVersion(Patch)
		}
	default:
		if v.IsReleaseCandidate() {
			v[Pre] = 0
			return
		}
		v.increaseVersion(Patch)
		return
	}

	v.IncreasePre(pre)
}

// IsReleaseCandidate validates the version for a release candidate.
func (v Version) IsReleaseCandidate() bool {
	return v[Pre] > 0
}

// IncreasePre increase the release candidate value of the version.
func (v *Version) IncreasePre(pre bool) {
	if pre {
		v.increaseVersion(Pre)
	}
}

func (v Version) increaseVersion(barrier int) {
	for i := 0; i < len(v); i++ {
		switch {
		case i < barrier:
			continue
		case i == barrier:
			v[i] += 1
		default:
			v[i] = 0
		}
	}
}

// Byte returns the value of the version as byte type.
func (v Version) Byte() (version uint) {
	for _, pos := range v {
		version = (version << 8) + uint(pos)
	}
	return
}

// String returns the version as string.
func (v Version) String() (version string) {
	if version = fmt.Sprintf("v%v.%v.%v", v[Major], v[Minor], v[Patch]); v.IsReleaseCandidate() {
		version += fmt.Sprintf("-RC.%v", v[Pre])
	}
	return
}
