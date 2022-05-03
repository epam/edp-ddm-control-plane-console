package registry

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortByVersion(t *testing.T) {
	in := []string{"1.5.6-SNAP-200", "1.5.6-SNAP-210", "1.5.6-SNAP-208", "1.5.6-SNAP-100", "1.5.6-SNAP-300"}
	sort.Sort(SortByVersion(in))
	assert.Equal(t, in,
		[]string{"1.5.6-SNAP-100", "1.5.6-SNAP-200", "1.5.6-SNAP-208", "1.5.6-SNAP-210", "1.5.6-SNAP-300"})
}
