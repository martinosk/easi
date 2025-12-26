package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion_Increment(t *testing.T) {
	ver, _ := NewVersion(1)

	newVer := ver.Increment()

	assert.Equal(t, 2, newVer.Value())
}
