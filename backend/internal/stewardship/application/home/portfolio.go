package home

import (
	"errors"
	"fmt"
)

const domainNamesShown = 5

var ErrUnknownCapabilityStatus = errors.New("unknown capability status")

type PortfolioSeed struct {
	All            bool
	DomainIDs      []string
	CapabilityIDs  []string
	ApplicationIDs []string
}

type PortfolioFacts struct {
	Statuses     StatusCounts
	Applications int
	DomainNames  []string
	Grades       GradeCounts
}

type StatusCounts struct {
	Active     int
	Planned    int
	Deprecated int
}

func (s *StatusCounts) Add(recordedStatus string, n int) error {
	switch recordedStatus {
	case "Active":
		s.Active += n
	case "Planned":
		s.Planned += n
	case "Deprecated":
		s.Deprecated += n
	default:
		return fmt.Errorf("%w: %q", ErrUnknownCapabilityStatus, recordedStatus)
	}
	return nil
}

func (s StatusCounts) total() int {
	return s.Active + s.Planned + s.Deprecated
}

func (f PortfolioFacts) portfolio() *Portfolio {
	names := f.DomainNames
	if len(names) > domainNamesShown {
		names = names[:domainNamesShown]
	}
	return &Portfolio{
		Capabilities: &CapabilityTile{Total: f.Statuses.total(), ByStatus: StatusBreakdown(f.Statuses)},
		Applications: &CountTile{Total: f.Applications},
		Domains:      &DomainTile{Total: len(f.DomainNames), Names: names},
		Time:         &TimeTile{Total: f.Grades.total(), Shares: timeSharesOf(f.Grades)},
	}
}
