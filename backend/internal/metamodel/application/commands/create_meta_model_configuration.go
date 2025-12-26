package commands

type CreateMetaModelConfiguration struct {
	TenantID  string
	CreatedBy string
	ID        string
}

func (c CreateMetaModelConfiguration) CommandName() string {
	return "CreateMetaModelConfiguration"
}
