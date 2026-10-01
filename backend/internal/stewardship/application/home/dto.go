package home

import "easi/backend/internal/shared/types"

type Home struct {
	Scope     Scope      `json:"scope"`
	Portfolio *Portfolio `json:"portfolio,omitempty"`
	MyWork    *MyWork    `json:"myWork,omitempty"`
}

type Scope struct {
	Kind                   ScopeKind          `json:"kind"`
	Stewardships           []ScopeStewardship `json:"stewardships"`
	ArchitectedDomains     []DomainRef        `json:"architectedDomains,omitempty"`
	ArchitectedDomainCount int                `json:"architectedDomainCount"`
	EAOwnedCapabilities    int                `json:"eaOwnedCapabilities"`
	OwnedApplications      int                `json:"ownedApplications"`
	EditGrants             int                `json:"editGrants"`
}

type DomainRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ScopeStewardship struct {
	Domain       *DomainRef `json:"domain,omitempty"`
	Concern      string     `json:"concern"`
	ConcernLabel string     `json:"concernLabel"`
}

type Portfolio struct {
	Capabilities *CapabilityTile `json:"capabilities,omitempty"`
	Applications *CountTile      `json:"applications,omitempty"`
	Domains      *DomainTile     `json:"domains,omitempty"`
	Time         *TimeTile       `json:"time,omitempty"`
}

type CapabilityTile struct {
	Total    int             `json:"total"`
	ByStatus StatusBreakdown `json:"byStatus"`
}

type StatusBreakdown struct {
	Active     int `json:"active"`
	Planned    int `json:"planned"`
	Deprecated int `json:"deprecated"`
}

type CountTile struct {
	Total int `json:"total"`
}

type DomainTile struct {
	Total int      `json:"total"`
	Names []string `json:"names,omitempty"`
}

type TimeTile struct {
	Total  int         `json:"total"`
	Shares *TimeShares `json:"shares,omitempty"`
}

type TimeShares struct {
	Invest      int `json:"invest"`
	Tolerate    int `json:"tolerate"`
	Migrate     int `json:"migrate"`
	Eliminate   int `json:"eliminate"`
	NotAssessed int `json:"notAssessed"`
}

type MyWork struct {
	Total int        `json:"total"`
	Items []WorkItem `json:"items"`
}

type WorkItem struct {
	SubjectType    SubjectType `json:"subjectType"`
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Level          string      `json:"level,omitempty"`
	Relation       Relation    `json:"relation"`
	GrantExpiresOn string      `json:"grantExpiresOn,omitempty"`
	DominantGrade  string      `json:"dominantGrade,omitempty"`
	Links          types.Links `json:"_links,omitempty"`
}

func (i WorkItem) subject() Subject {
	return Subject{Type: i.SubjectType, ID: i.ID}
}
