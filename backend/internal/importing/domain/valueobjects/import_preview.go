package valueobjects

import (
	"easi/backend/internal/shared/eventsourcing"
	"reflect"
)

type SupportedCounts struct {
	Capabilities             int `json:"capabilities"`
	Components               int `json:"components"`
	ParentChildRelationships int `json:"parentChildRelationships"`
	Realizations             int `json:"realizations"`
	ComponentRelationships   int `json:"componentRelationships"`
}

type UnsupportedCounts struct {
	Elements      map[string]int `json:"elements"`
	Relationships map[string]int `json:"relationships"`
}

type ImportPreview struct {
	supported   SupportedCounts
	unsupported UnsupportedCounts
}

func NewImportPreview(supported SupportedCounts, unsupported UnsupportedCounts) ImportPreview {
	if unsupported.Elements == nil {
		unsupported.Elements = make(map[string]int)
	}
	if unsupported.Relationships == nil {
		unsupported.Relationships = make(map[string]int)
	}
	return ImportPreview{
		supported:   supported,
		unsupported: unsupported,
	}
}

func (ip ImportPreview) Supported() SupportedCounts {
	return ip.supported
}

func (ip ImportPreview) Unsupported() UnsupportedCounts {
	return ip.unsupported
}

func (ip ImportPreview) TotalSupportedItems() int {
	return ip.supported.Capabilities +
		ip.supported.Components +
		ip.supported.ParentChildRelationships +
		ip.supported.Realizations +
		ip.supported.ComponentRelationships
}

func (ip ImportPreview) HasUnsupportedElements() bool {
	return len(ip.unsupported.Elements) > 0 || len(ip.unsupported.Relationships) > 0
}

func (ip ImportPreview) Equals(other domain.ValueObject) bool {
	if otherIP, ok := other.(ImportPreview); ok {
		return reflect.DeepEqual(ip.supported, otherIP.supported) &&
			reflect.DeepEqual(ip.unsupported, otherIP.unsupported)
	}
	return false
}

func (ip ImportPreview) String() string {
	return "ImportPreview"
}
