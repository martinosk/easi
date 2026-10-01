package home

import "time"

type OwnedCapability struct {
	ID    string
	Name  string
	Level string
}

type ApplicationOwnership struct {
	ApplicationID string
	Name          string
	State         string
}

type EditGrant struct {
	Subject   Subject
	Name      string
	Level     string
	ExpiresAt time.Time
}

type CallerFacts struct {
	Stewardships        []StewardshipAnchor
	ArchitectedDomains  []Domain
	EAOwnedCapabilities []OwnedCapability
	Ownerships          []ApplicationOwnership
	Grants              []EditGrant
}

func (f CallerFacts) Anchors() Anchors {
	subjects := make([]SubjectAnchor, 0, len(f.EAOwnedCapabilities)+len(f.Ownerships)+len(f.Grants))
	for _, capability := range f.EAOwnedCapabilities {
		subjects = append(subjects, capability.anchor())
	}
	for _, ownership := range f.Ownerships {
		if relation, anchors := ownership.relation(); anchors {
			subjects = append(subjects, ownership.anchor(relation))
		}
	}
	for _, grant := range f.Grants {
		subjects = append(subjects, grant.anchor())
	}
	return Anchors{Stewardships: f.Stewardships, ArchitectedDomains: f.ArchitectedDomains, Subjects: subjects}
}

func (c OwnedCapability) anchor() SubjectAnchor {
	return SubjectAnchor{
		Subject:  Subject{Type: SubjectCapability, ID: c.ID},
		Name:     c.Name,
		Level:    c.Level,
		Relation: RelationEAOwner,
	}
}

func (o ApplicationOwnership) relation() (Relation, bool) {
	switch o.State {
	case ownershipOwned:
		return RelationOwner, true
	case ownershipNominated:
		return RelationNominated, true
	}
	return "", false
}

func (o ApplicationOwnership) anchor(relation Relation) SubjectAnchor {
	return SubjectAnchor{
		Subject:  Subject{Type: SubjectApplication, ID: o.ApplicationID},
		Name:     o.Name,
		Relation: relation,
	}
}

func (g EditGrant) anchor() SubjectAnchor {
	return SubjectAnchor{
		Subject:        g.Subject,
		Name:           g.Name,
		Level:          g.Level,
		Relation:       RelationEditGrant,
		GrantExpiresAt: g.ExpiresAt,
	}
}
