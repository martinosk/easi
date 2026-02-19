package valueobjects

type TokenCount struct {
	value int
}

func NewTokenCount(value int) TokenCount {
	if value < 0 {
		value = 0
	}
	return TokenCount{value: value}
}

func ReconstructTokenCount(value int) TokenCount {
	return TokenCount{value: value}
}

func (t TokenCount) Value() int   { return t.value }
func (t TokenCount) Pointer() *int { v := t.value; return &v }
