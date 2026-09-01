package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHostingClassification_ValidValues(t *testing.T) {
	for _, value := range []string{
		HostingOnPremises,
		HostingCloud,
		HostingSaaS,
		HostingThirdPartyHosted,
		HostingUnknown,
	} {
		hosting, err := NewHostingClassification(value)
		require.NoError(t, err)
		assert.Equal(t, value, hosting.String())
	}
}

func TestNewHostingClassification_InvalidValue(t *testing.T) {
	_, err := NewHostingClassification("mainframe")
	assert.ErrorIs(t, err, ErrInvalidHostingClassification)
}

func TestUnknownHostingClassification(t *testing.T) {
	hosting := UnknownHostingClassification()
	assert.Equal(t, HostingUnknown, hosting.String())
}

func TestHostingClassification_Equals(t *testing.T) {
	a, _ := NewHostingClassification(HostingSaaS)
	b, _ := NewHostingClassification(HostingSaaS)
	c, _ := NewHostingClassification(HostingCloud)

	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(c))
}
