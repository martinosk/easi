package api

import (
	"testing"

	"easi/backend/internal/onepagers/application/queries"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildOnePagerDTO_SkipsZeroValueField(t *testing.T) {
	onePager := &queries.OnePager{
		SubjectType: "application",
		SubjectID:   "app-1",
		SubjectName: "Payments",
		Fields: []queries.Field{
			{BuiltIn: &queries.BuiltInField{ID: "description", Label: "Description"}},
			{},
		},
	}

	dto := BuildOnePagerDTO(onePager, testLinks())

	require.Len(t, dto.Fields, 1)
	assert.Equal(t, "builtIn", dto.Fields[0].Kind)
}
