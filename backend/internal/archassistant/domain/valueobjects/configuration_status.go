package valueobjects

import "errors"

var ErrInvalidConfigurationStatus = errors.New("invalid configuration status: must be not_configured, configured, or error")

type ConfigurationStatus struct {
	value string
}

var (
	StatusNotConfigured = ConfigurationStatus{value: "not_configured"}
	StatusConfigured    = ConfigurationStatus{value: "configured"}
	StatusError         = ConfigurationStatus{value: "error"}
)

func ConfigurationStatusFromString(s string) (ConfigurationStatus, error) {
	switch s {
	case "not_configured":
		return StatusNotConfigured, nil
	case "configured":
		return StatusConfigured, nil
	case "error":
		return StatusError, nil
	default:
		return ConfigurationStatus{}, ErrInvalidConfigurationStatus
	}
}

func (s ConfigurationStatus) Value() string { return s.value }

func (s ConfigurationStatus) IsConfigured() bool {
	return s.value == "configured"
}
