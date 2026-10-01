package home

type Readable struct {
	Capabilities bool
	Applications bool
	Direction    bool
	Domains      bool
}

func (r Readable) subject(subjectType SubjectType) bool {
	if subjectType == SubjectCapability {
		return r.Capabilities
	}
	return r.Applications
}

func (h *Home) RedactFor(readable Readable) {
	if !readable.Domains {
		h.Scope.withoutDomainNames()
	}
	if h.Portfolio != nil {
		h.Portfolio.redactFor(readable)
	}
	if h.MyWork != nil {
		h.MyWork.redactFor(readable)
	}
}

func (s *Scope) withoutDomainNames() {
	s.ArchitectedDomains = nil
	for i := range s.Stewardships {
		s.Stewardships[i].Domain = nil
	}
}

func (p *Portfolio) redactFor(readable Readable) {
	if !readable.Capabilities {
		p.Capabilities = nil
	}
	if !readable.Applications {
		p.Applications = nil
	}
	if !readable.Direction {
		p.Time = nil
	}
	if !readable.Domains && p.Domains != nil {
		p.Domains.Names = nil
	}
}

func (m *MyWork) redactFor(readable Readable) {
	kept := make([]WorkItem, 0, len(m.Items))
	for _, item := range m.Items {
		if !readable.subject(item.SubjectType) {
			continue
		}
		if !readable.Direction {
			item.DominantGrade = ""
		}
		kept = append(kept, item)
	}
	m.Items = kept
	m.Total = len(kept)
}
