package go_apario_identifier

import (
	`fmt`
	`strconv`
	`strings`
)

func ParseVersion(v string) *Version {
	version := &Version{}
	if strings.HasPrefix(v, "v") {
		v = strings.ReplaceAll(v, `v`, ``)
		p := strings.Split(v, `.`)
		if len(p) == 3 {
			major, mmIntErr := strconv.Atoi(p[0])
			if mmIntErr == nil {
				version.Major = major
			}
			minor, mIntErr := strconv.Atoi(p[1])
			if mIntErr == nil {
				version.Minor = minor
			}
			patch, pIntErr := strconv.Atoi(p[2])
			if pIntErr == nil {
				version.Patch = patch
			} else {
				version.Patch = 1
			}
		}
	}
	return version
}

type Version struct {
	Major int `json:"ma"`
	Minor int `json:"mi"`
	Patch int `json:"pa"`
}

func (v *Version) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v *Version) BumpMajor() bool {
	v.Major += 1
	v.Minor = 0
	v.Patch = 0
	return true
}

func (v *Version) BumpMinor() bool {
	v.Minor += 1
	v.Patch = 0
	return true
}

func (v *Version) BumpPatch() bool {
	v.Patch += 1
	return true
}
