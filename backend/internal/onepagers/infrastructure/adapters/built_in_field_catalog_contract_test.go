package adapters

import (
	"reflect"
	"strings"
	"testing"

	amContracts "easi/backend/internal/architecturemodeling/publishedlanguage/contracts"
	capContracts "easi/backend/internal/capabilitymapping/publishedlanguage/contracts"
	eaContracts "easi/backend/internal/enterprisearchitecture/publishedlanguage/contracts"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var contractPayloadsBySubjectType = map[string][]reflect.Type{
	"capability": {
		reflect.TypeOf(capContracts.CapabilityCreatedPayload{}),
		reflect.TypeOf(capContracts.CapabilityUpdatedPayload{}),
		reflect.TypeOf(capContracts.CapabilityMetadataUpdatedPayload{}),
	},
	"enterprise-capability": {
		reflect.TypeOf(eaContracts.EnterpriseCapabilityCreatedPayload{}),
		reflect.TypeOf(eaContracts.EnterpriseCapabilityUpdatedPayload{}),
	},
	"application": {
		reflect.TypeOf(amContracts.ApplicationComponentCreatedPayload{}),
		reflect.TypeOf(amContracts.ApplicationComponentUpdatedPayload{}),
	},
	"acquired-entity": {
		reflect.TypeOf(amContracts.AcquiredEntityCreatedPayload{}),
		reflect.TypeOf(amContracts.AcquiredEntityUpdatedPayload{}),
	},
	"vendor": {
		reflect.TypeOf(amContracts.VendorCreatedPayload{}),
		reflect.TypeOf(amContracts.VendorUpdatedPayload{}),
	},
	"internal-team": {
		reflect.TypeOf(amContracts.InternalTeamCreatedPayload{}),
		reflect.TypeOf(amContracts.InternalTeamUpdatedPayload{}),
	},
}

func isGenericMergeAttribute(attribute builtInAttribute) bool {
	return reflect.ValueOf(attribute.decode).Pointer() != reflect.ValueOf(decodeExperts).Pointer()
}

func jsonFieldNames(structType reflect.Type) map[string]bool {
	names := make(map[string]bool, structType.NumField())
	for i := 0; i < structType.NumField(); i++ {
		name, _, _ := strings.Cut(structType.Field(i).Tag.Get("json"), ",")
		if name == "" || name == "-" {
			continue
		}
		names[name] = true
	}
	return names
}

func hasMatchingContractField(contractTypes []reflect.Type, attributeName string) bool {
	for _, contractType := range contractTypes {
		if jsonFieldNames(contractType)[attributeName] {
			return true
		}
	}
	return false
}

func TestBuiltInAttributesBySubjectType_MatchSupplierPublishedContracts(t *testing.T) {
	for subjectType, attributes := range builtInAttributesBySubjectType {
		contractTypes, registered := contractPayloadsBySubjectType[subjectType]
		require.Truef(t, registered, "no published-language contract registered for subject type %q", subjectType)

		for _, attribute := range attributes {
			if !isGenericMergeAttribute(attribute) {
				continue
			}
			assert.Truef(t, hasMatchingContractField(contractTypes, attribute.attribute),
				"%s bound attribute %q has no matching field on its supplier's published-language contract", subjectType, attribute.attribute)
		}
	}
}
