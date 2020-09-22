package patch

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWhiteList(t *testing.T) {
	info := Info{
		Input:  []string{"Field1", "Field2", "Field3"},
		Output: []string{"field_1", "field_2", "field_3"},
	}
	info.Whitelist([]string{"Field1", "Field3", "Field4"})
	// Field4 should be removed from Input/Output because unknown from the structure
	// Field2 should be removed from Input/Output because not in the whitelist
	assert.Equal(t, []string{"Field1", "Field3"}, info.Input)
	assert.Equal(t, []string{"field_1", "field_3"}, info.Output)
}
