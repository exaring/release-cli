package version

import (
	"testing"
)

func TestNew(t *testing.T) {
	tt := []struct {
		value    string
		expected []uint
	}{
		{"", []uint{0, 0, 0, 0}},
		{"1", Version{0, 0, 0, 0}},
		{"1.1", Version{0, 0, 0, 0}},
		{".", Version{0, 0, 0, 0}},
		{"1.", Version{0, 0, 0, 0}},
		{"1.1.", Version{0, 0, 0, 0}},
		{"x.x", Version{0, 0, 0, 0}},
		{"1.x.1", Version{0, 0, 0, 0}},
		{"1.2.3-RC", Version{1, 2, 3, 0}},
		{"1.2.3-RCy", Version{1, 2, 3, 0}},
		{"1.1.1", Version{1, 1, 1, 0}},
		{"1.2.3-RC4", Version{1, 2, 3, 4}},
		{"refs/tags/1.0.0", Version{1, 0, 0, 0}},
		{"x/y/z/1.2.3-RC4", Version{1, 2, 3, 4}},
		{"x/y/z/v1.2.3-RC4", Version{1, 2, 3, 4}},
	}

	for _, tc := range tt {
		t.Run("", func(t *testing.T) {
			v, err := New(tc.value)
			if err != nil {
				t.Error("Couldn't create new version instance")
			}

			for index, entry := range tc.expected {
				if entry != v[index] {
					t.Errorf("The elements are not equal: \n value: \t %v \n actual: \t %v \n expected: \t %v",
						tc.value, v, tc.expected)
				}
			}
		})
	}
}

func TestVersion_Increase(t *testing.T) {
	tt := []struct {
		value                    string
		major, minor, patch, pre bool
		expected                 []uint
	}{
		{"0.0.0", false, false, false, true, Version{0, 0, 1, 1}},
		{"0.0.0", false, false, true, false, Version{0, 0, 1, 0}},
		{"0.0.0", false, true, false, false, Version{0, 1, 0, 0}},
		{"0.0.0", true, false, false, false, Version{1, 0, 0, 0}},
		{"1.1.1", false, false, true, true, Version{1, 1, 2, 1}},
		{"1.1.1-RC1", false, false, false, false, Version{1, 1, 1, 0}},
		{"1.1.1-RC1", false, false, true, false, Version{1, 1, 1, 0}},
		{"1.1.1-RC1", false, true, false, false, Version{1, 1, 1, 0}},
		{"1.1.1-RC1", false, true, true, false, Version{1, 1, 1, 0}},
		{"1.1.1-RC1", true, false, true, false, Version{1, 1, 1, 0}},
		{"1.1.1-RC1", false, true, false, true, Version{1, 1, 1, 0}},
		{"1.1.1-RC1", true, false, false, false, Version{1, 1, 1, 0}},
		{"1.1.1-RC1", true, false, true, false, Version{1, 1, 1, 0}},
		{"1.1.1-RC1", true, false, false, true, Version{1, 1, 1, 0}},
		{"1.1.1", false, false, false, false, Version{1, 1, 2, 0}},
		{"1.1.1", false, false, true, false, Version{1, 1, 2, 0}},
		{"1.1.1", false, true, false, false, Version{1, 2, 0, 0}},
		{"1.1.1", true, false, false, false, Version{2, 0, 0, 0}},
		{"1.1.1", true, true, false, false, Version{2, 0, 0, 0}},
		{"1.1.1", true, false, true, false, Version{2, 0, 0, 0}},
		{"1.1.1", true, false, false, true, Version{2, 0, 0, 1}},
		{"1.1.1", true, true, false, true, Version{2, 0, 0, 1}},
		{"1.1.1", true, false, true, false, Version{2, 0, 0, 0}},
		{"1.1.1", true, true, true, false, Version{2, 0, 0, 0}},
	}

	for _, tc := range tt {
		t.Run("", func(t *testing.T) {
			v, err := New(tc.value)
			if err != nil {
				t.Error("Couldn't create new version instance")
			}

			v.Increase(tc.major, tc.minor, tc.patch, tc.pre)

			for index, entry := range tc.expected {
				if entry != v[index] {
					t.Errorf("The elements are not equal: \n value: \t %v \n actual: \t %v \n expected: \t %v",
						tc.value, v, tc.expected)
				}
			}
		})
	}

}

func TestVersion_String(t *testing.T) {
	tt := []struct {
		expectedValue string
	}{
		{"v0.0.0"},
		{"v1.1.1-RC1"},
	}

	for _, tc := range tt {
		t.Run("", func(t *testing.T) {
			v, err := New(tc.expectedValue)
			if err != nil {
				t.Error("Couldn't create new version instance")
			}

			if tc.expectedValue != v.String() {
				t.Errorf("The elements are not equal: \n actual: \t %v \n expected: \t %v",
					v, tc.expectedValue)
			}
		})
	}
}
