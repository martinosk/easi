package contracts

import "time"

type ApplicationComponentCreatedPayload struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

type ApplicationComponentUpdatedPayload struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type ApplicationComponentDeletedPayload struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	DeletedAt time.Time `json:"deletedAt"`
}

type AcquiredEntityCreatedPayload struct {
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	AcquisitionDate   *time.Time `json:"acquisitionDate,omitempty"`
	IntegrationStatus string     `json:"integrationStatus"`
	Notes             string     `json:"notes"`
	CreatedAt         time.Time  `json:"createdAt"`
}

type AcquiredEntityUpdatedPayload struct {
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	AcquisitionDate   *time.Time `json:"acquisitionDate,omitempty"`
	IntegrationStatus string     `json:"integrationStatus"`
	Notes             string     `json:"notes"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

type VendorCreatedPayload struct {
	ID                    string    `json:"id"`
	Name                  string    `json:"name"`
	ImplementationPartner string    `json:"implementationPartner"`
	Notes                 string    `json:"notes"`
	CreatedAt             time.Time `json:"createdAt"`
}

type VendorUpdatedPayload struct {
	ID                    string    `json:"id"`
	Name                  string    `json:"name"`
	ImplementationPartner string    `json:"implementationPartner"`
	Notes                 string    `json:"notes"`
	UpdatedAt             time.Time `json:"updatedAt"`
}

type InternalTeamCreatedPayload struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Department    string    `json:"department"`
	ContactPerson string    `json:"contactPerson"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"createdAt"`
}

type InternalTeamUpdatedPayload struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Department    string    `json:"department"`
	ContactPerson string    `json:"contactPerson"`
	Notes         string    `json:"notes"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
