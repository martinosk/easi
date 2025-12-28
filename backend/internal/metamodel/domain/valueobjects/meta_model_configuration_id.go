package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type MetaModelConfigurationID struct {
	sharedvo.UUIDValue
}

func NewMetaModelConfigurationID() MetaModelConfigurationID {
	return MetaModelConfigurationID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewMetaModelConfigurationIDFromString(value string) (MetaModelConfigurationID, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return MetaModelConfigurationID{}, err
	}
	return MetaModelConfigurationID{UUIDValue: uuidValue}, nil
}

func (m MetaModelConfigurationID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(MetaModelConfigurationID); ok {
		return m.UUIDValue.EqualsValue(otherID.UUIDValue)
	}
	return false
}
