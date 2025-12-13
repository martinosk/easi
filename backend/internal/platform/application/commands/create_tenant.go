package commands

type CreateTenant struct {
	ID              string
	Name            string
	Domains         []string
	DiscoveryURL    string
	ClientID        string
	AuthMethod      string
	Scopes          string
	FirstAdminEmail string
}

func (c CreateTenant) CommandName() string {
	return "CreateTenant"
}
