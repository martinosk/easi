package valueobjects

import (
	"errors"
	"fmt"

	domain "easi/backend/internal/shared/eventsourcing"
)

var ErrInvalidHostingClassification = errors.New("invalid hosting classification: must be on-premises, cloud, saas, third-party-hosted, or unknown")

const (
	HostingOnPremises       = "on-premises"
	HostingCloud            = "cloud"
	HostingSaaS             = "saas"
	HostingThirdPartyHosted = "third-party-hosted"
	HostingUnknown          = "unknown"
)

type HostingClassification struct {
	value string
}

func NewHostingClassification(value string) (HostingClassification, error) {
	switch value {
	case HostingOnPremises, HostingCloud, HostingSaaS, HostingThirdPartyHosted, HostingUnknown:
		return HostingClassification{value: value}, nil
	default:
		return HostingClassification{}, fmt.Errorf("%w: %s", ErrInvalidHostingClassification, value)
	}
}

func UnknownHostingClassification() HostingClassification {
	return HostingClassification{value: HostingUnknown}
}

func (h HostingClassification) String() string {
	return h.value
}

func (h HostingClassification) Equals(other domain.ValueObject) bool {
	if otherHosting, ok := other.(HostingClassification); ok {
		return h.value == otherHosting.value
	}
	return false
}
