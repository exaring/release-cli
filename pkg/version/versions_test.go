package version

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersions_Less(t *testing.T) {
	tt := []struct {
		versions     []string
		expectedLess bool
	}{
		{[]string{"0.0.1", "0.0.2"}, true},
		{[]string{"0.0.2", "0.0.1"}, false},
		{[]string{"0.1.0", "0.2.0"}, true},
		{[]string{"0.2.0", "0.1.0"}, false},
		{[]string{"1.0.0", "2.0.0"}, true},
		{[]string{"2.0.0", "1.0.0"}, false},
		{[]string{"1.1.0", "1.2.0"}, true},
		{[]string{"1.2.0", "1.1.0"}, false},
		{[]string{"1.1.0", "2.1.0"}, true},
		{[]string{"2.1.0", "1.1.0"}, false},
		{[]string{"1.1.1", "2.1.1"}, true},
		{[]string{"2.1.1", "1.1.1"}, false},
		{[]string{"1.1.1", "1.1.1"}, false},
		{[]string{"2.1.1-RC.1", "1.1.1-RC.2"}, false},
		{[]string{"0.1.0-RC.2", "0.2.0"}, true},
		{[]string{"0.1.0-RC.2", "0.1.0"}, true},
		{[]string{"0.4.0", "0.1.0-RC.1"}, false},
		{[]string{"0.1.0-RC.1", "0.5.0-RC.2"}, true},
		{[]string{"1.1.1-RC.1", "1.1.1-RC.10"}, true},
		{[]string{"1.1.1-RC.1", "1.1.1"}, true},
		{[]string{"1.1.1", "1.1.1-RC.1"}, false},
		{[]string{"1.1.1-RC.1", "1.1.1-RC.10"}, true},
		{[]string{"1.1.1-RC.10", "1.1.1-RC.1"}, false},
		{[]string{"1.1.1-RC.1", "1.1.2"}, true},
		{[]string{"1.1.2", "1.1.1-RC.1"}, false},
		{[]string{"1.0.0", "0.1.1-RC.1"}, false},
		{[]string{"0.1.1-RC.1", "1.0.0"}, true},
	}

	for _, tc := range tt {
		t.Run(fmt.Sprintf("%v", tc.versions), func(t *testing.T) {
			var versions = make(Versions, len(tc.versions))
			for i, o := range tc.versions {
				v, err := New(o)
				assert.NoError(t, err)
				versions[i] = v
			}
			assert.Equal(t, tc.expectedLess, versions.Less(0, 1))
		})
	}
}
