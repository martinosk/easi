package valueobjects

import (
	"math"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTextValue_TrimsAndAccepts(t *testing.T) {
	value, err := NewTextValue("  Runs on shared Kubernetes cluster  ")
	require.NoError(t, err)
	assert.Equal(t, "Runs on shared Kubernetes cluster", value.Value())
	assert.Equal(t, "text", value.FieldTypeValue())
}

func TestNewTextValue_RejectsWhitespaceOnly(t *testing.T) {
	_, err := NewTextValue("   ")
	assert.ErrorIs(t, err, ErrTextValueEmpty)
}

func TestNewTextValue_RejectsTooLong(t *testing.T) {
	_, err := NewTextValue(strings.Repeat("a", MaxTextValueLength+1))
	assert.ErrorIs(t, err, ErrTextValueTooLong)
}

func TestNewNumberValue_AcceptsFiniteDecimal(t *testing.T) {
	value, err := NewNumberValue(42.5)
	require.NoError(t, err)
	assert.InDelta(t, 42.5, value.Value(), 0)
	assert.Equal(t, "number", value.FieldTypeValue())
}

func TestNewNumberValue_RejectsNonFinite(t *testing.T) {
	for _, raw := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		_, err := NewNumberValue(raw)
		assert.ErrorIs(t, err, ErrNumberValueNotFinite)
	}
}

func TestNewDateValue_AcceptsISODate(t *testing.T) {
	value, err := NewDateValue("2026-03-01")
	require.NoError(t, err)
	assert.Equal(t, "2026-03-01", value.Value())
	assert.Equal(t, "date", value.FieldTypeValue())
}

func TestNewDateValue_RejectsNonISODates(t *testing.T) {
	for _, raw := range []string{"March 1st", "2026-13-40", "01-03-2026", "", "2026-03-01T10:00:00Z"} {
		t.Run(raw, func(t *testing.T) {
			_, err := NewDateValue(raw)
			assert.ErrorIs(t, err, ErrDateValueInvalid)
		})
	}
}

func TestNewLinkValue_AcceptsLabelAndAbsoluteURL(t *testing.T) {
	value, err := NewLinkValue("MSA", "https://contracts.example.com")
	require.NoError(t, err)
	assert.Equal(t, "MSA", value.Label())
	assert.Equal(t, "https://contracts.example.com", value.URL().Value())
	assert.Equal(t, "link", value.FieldTypeValue())
}

func TestNewLinkValue_RejectsInvalidInput(t *testing.T) {
	cases := []struct {
		name  string
		label string
		url   string
	}{
		{"empty label", "   ", "https://contracts.example.com"},
		{"too long label", strings.Repeat("a", MaxLinkLabelLength+1), "https://contracts.example.com"},
		{"ftp url", "MSA", "ftp://x"},
		{"relative url", "MSA", "/contracts/msa"},
		{"empty url", "MSA", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewLinkValue(tc.label, tc.url)
			assert.Error(t, err)
		})
	}
}

func TestNewSelectionValue_AcceptsOptionID(t *testing.T) {
	optionID := uuid.New().String()
	value, err := NewSelectionValue(optionID)
	require.NoError(t, err)
	assert.Equal(t, optionID, value.OptionID().Value())
	assert.Equal(t, "selection", value.FieldTypeValue())
}

func TestNewSelectionValue_RejectsInvalidOptionID(t *testing.T) {
	_, err := NewSelectionValue("not-a-uuid")
	assert.ErrorIs(t, err, ErrInvalidOptionID)
}

func TestNewContactPerson_AcceptsNameEmailAndOptionalCompany(t *testing.T) {
	value, err := NewContactPerson(ContactPersonParams{Name: "A. Larsen", Email: "al@ext.example", Company: "Ext ApS"})
	require.NoError(t, err)
	assert.Equal(t, "A. Larsen", value.Name())
	assert.Equal(t, "al@ext.example", value.Email().Value())
	assert.Equal(t, "Ext ApS", value.Company())
	assert.Equal(t, "contact-person", value.FieldTypeValue())

	noCompany, err := NewContactPerson(ContactPersonParams{Name: "A. Larsen", Email: "al@ext.example", Company: ""})
	require.NoError(t, err)
	assert.Empty(t, noCompany.Company())
}

func TestNewContactPerson_RejectsInvalidInput(t *testing.T) {
	cases := []struct {
		name    string
		person  string
		email   string
		company string
		wantErr error
	}{
		{"empty name", "  ", "al@ext.example", "", ErrContactNameEmpty},
		{"too long name", strings.Repeat("a", MaxContactNameLength+1), "al@ext.example", "", ErrContactNameTooLong},
		{"invalid email", "A. Larsen", "not-an-email", "", ErrUserEmailInvalid},
		{"empty email", "A. Larsen", "", "", ErrUserEmailEmpty},
		{"too long company", "A. Larsen", "al@ext.example", strings.Repeat("a", MaxContactCompanyLength+1), ErrContactCompanyTooLong},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewContactPerson(ContactPersonParams{Name: tc.person, Email: tc.email, Company: tc.company})
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestFieldValue_Equals(t *testing.T) {
	textA, err := NewTextValue("same")
	require.NoError(t, err)
	textB, err := NewTextValue("same")
	require.NoError(t, err)
	textC, err := NewTextValue("different")
	require.NoError(t, err)
	number, err := NewNumberValue(1)
	require.NoError(t, err)

	assert.True(t, textA.Equals(textB))
	assert.False(t, textA.Equals(textC))
	assert.False(t, textA.Equals(number))

	linkA, err := NewLinkValue("MSA", "https://contracts.example.com")
	require.NoError(t, err)
	linkB, err := NewLinkValue("MSA", "https://contracts.example.com")
	require.NoError(t, err)
	linkC, err := NewLinkValue("DPA", "https://contracts.example.com")
	require.NoError(t, err)
	assert.True(t, linkA.Equals(linkB))
	assert.False(t, linkA.Equals(linkC))

	contactA, err := NewContactPerson(ContactPersonParams{Name: "A. Larsen", Email: "al@ext.example", Company: "Ext ApS"})
	require.NoError(t, err)
	contactB, err := NewContactPerson(ContactPersonParams{Name: "A. Larsen", Email: "al@ext.example", Company: "Ext ApS"})
	require.NoError(t, err)
	contactC, err := NewContactPerson(ContactPersonParams{Name: "A. Larsen", Email: "al@ext.example", Company: ""})
	require.NoError(t, err)
	assert.True(t, contactA.Equals(contactB))
	assert.False(t, contactA.Equals(contactC))
}

func TestDisplayText_RendersEachKind(t *testing.T) {
	optionID := uuid.New().String()
	selection, err := NewSelectionValue(optionID)
	require.NoError(t, err)

	cases := []struct {
		value FieldValue
		want  string
	}{
		{mustText(t, "plain"), "plain"},
		{mustNumber(t, 42.5), "42.5"},
		{mustDate(t, "2026-03-01"), "2026-03-01"},
		{mustLink(t, "MSA", "https://contracts.example.com"), "MSA"},
		{selection, optionID},
		{mustContact(t, "A. Larsen", "al@ext.example", "Ext ApS"), "A. Larsen"},
	}
	for _, tc := range cases {
		t.Run(tc.value.FieldTypeValue(), func(t *testing.T) {
			assert.Equal(t, tc.want, DisplayText(tc.value))
		})
	}
}

func mustText(t *testing.T, raw string) FieldValue {
	t.Helper()
	v, err := NewTextValue(raw)
	require.NoError(t, err)
	return v
}

func mustNumber(t *testing.T, raw float64) FieldValue {
	t.Helper()
	v, err := NewNumberValue(raw)
	require.NoError(t, err)
	return v
}

func mustDate(t *testing.T, raw string) FieldValue {
	t.Helper()
	v, err := NewDateValue(raw)
	require.NoError(t, err)
	return v
}

func mustLink(t *testing.T, label, url string) FieldValue {
	t.Helper()
	v, err := NewLinkValue(label, url)
	require.NoError(t, err)
	return v
}

func mustContact(t *testing.T, name, email, company string) FieldValue {
	t.Helper()
	v, err := NewContactPerson(ContactPersonParams{Name: name, Email: email, Company: company})
	require.NoError(t, err)
	return v
}
