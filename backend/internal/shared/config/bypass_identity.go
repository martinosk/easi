package config

type SyntheticIdentity struct {
	UserID     string
	Email      string
	Name       string
	Role       string
	TenantID   string
	TenantName string
}

func BypassIdentity() SyntheticIdentity {
	return SyntheticIdentity{
		UserID:     "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12",
		Email:      "admin@acme.com",
		Name:       "Admin User (Bypass)",
		Role:       "admin",
		TenantID:   "acme",
		TenantName: "ACME Corporation",
	}
}
