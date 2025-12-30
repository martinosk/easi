package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/google/uuid"
)

var enterpriseStrategicImportanceNamespace = uuid.MustParse("d3b07384-d113-4ec6-a645-3e6a4c7f8b1a")

type EnterpriseStrategicImportanceID struct {
	sharedvo.UUIDValue
}

func NewEnterpriseStrategicImportanceID() EnterpriseStrategicImportanceID {
	return EnterpriseStrategicImportanceID{UUIDValue: sharedvo.NewUUIDValue()}
}

func NewEnterpriseStrategicImportanceIDFromComposite(
	enterpriseCapabilityID EnterpriseCapabilityID,
	pillarID PillarID,
) EnterpriseStrategicImportanceID {
	composite := enterpriseCapabilityID.Value() + ":" + pillarID.Value()
	deterministicID := uuid.NewSHA1(enterpriseStrategicImportanceNamespace, []byte(composite))
	uuidValue, _ := sharedvo.NewUUIDValueFromString(deterministicID.String())
	return EnterpriseStrategicImportanceID{UUIDValue: uuidValue}
}

func NewEnterpriseStrategicImportanceIDFromString(value string) (EnterpriseStrategicImportanceID, error) {
	uuidValue, err := sharedvo.NewUUIDValueFromString(value)
	if err != nil {
		return EnterpriseStrategicImportanceID{}, err
	}
	return EnterpriseStrategicImportanceID{UUIDValue: uuidValue}, nil
}

func (e EnterpriseStrategicImportanceID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(EnterpriseStrategicImportanceID); ok {
		return e.UUIDValue.EqualsValue(otherID.UUIDValue)
	}
	return false
}
