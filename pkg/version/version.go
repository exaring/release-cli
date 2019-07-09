package version

import (
	"fmt"
	"regexp"
	"strconv"
)

const (
	Major = iota
	Minor
	Patch
	Pre

	RegExPatternVersionString = `((\d+)\.(\d+)\.(\d+))(?:-RC([\dA-Za-z\-]+(?:\.[\dA-Za-z\-]+)*))?`
)

type Version []uint8

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
		version = append(version, func(o string) uint8 {
			number, err := strconv.ParseInt(o, 10, 8)
			if err != nil {
				return 0
			}
			return uint8(number)
		}(r[i]))
	}

	return version, nil
}

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

func (v Version) IsReleaseCandidate() bool {
	return v[Pre] > 0
}

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

func (v Version) Byte() (version uint) {
	for _, pos := range v {
		version = (version << 8) + uint(pos)
	}
	return
}

func (v Version) String() (version string) {
	if version = fmt.Sprintf("v%v.%v.%v", v[Major], v[Minor], v[Patch]); v.IsReleaseCandidate() {
		version += fmt.Sprintf("-RC%v", v[Pre])
	}
	return
}
