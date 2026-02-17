package valueobjects

type EncryptedAPIKey struct {
	value string
}

func NewEncryptedAPIKey(value string) EncryptedAPIKey {
	return EncryptedAPIKey{value: value}
}

func (k EncryptedAPIKey) Value() string  { return k.value }
func (k EncryptedAPIKey) IsEmpty() bool  { return k.value == "" }
