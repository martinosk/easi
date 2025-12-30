package api

import (
	"easi/backend/internal/enterprisearchitecture/application/handlers"
	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/enterprisearchitecture/domain/valueobjects"
	"easi/backend/internal/enterprisearchitecture/infrastructure/repositories"
	sharedAPI "easi/backend/internal/shared/api"
)

func init() {
	registry := sharedAPI.GetErrorRegistry()

	registry.RegisterNotFound(repositories.ErrEnterpriseCapabilityNotFound, "Enterprise capability not found")
	registry.RegisterNotFound(repositories.ErrEnterpriseCapabilityLinkNotFound, "Enterprise capability link not found")
	registry.RegisterNotFound(repositories.ErrEnterpriseStrategicImportanceNotFound, "Strategic importance rating not found")

	registry.RegisterConflict(handlers.ErrEnterpriseCapabilityNameExists, "Enterprise capability with this name already exists")
	registry.RegisterConflict(handlers.ErrDomainCapabilityAlreadyLinked, "Domain capability is already linked to an enterprise capability")
	registry.RegisterConflict(handlers.ErrImportanceAlreadySet, "Strategic importance for this pillar is already set")

	registry.RegisterValidation(aggregates.ErrCannotLinkInactiveCapability, "Cannot link to an inactive enterprise capability")
	registry.RegisterValidation(valueobjects.ErrDescriptionTooLong, "Description exceeds maximum length of 1000 characters")
	registry.RegisterValidation(valueobjects.ErrLinkedByEmpty, "LinkedBy cannot be empty")
	registry.RegisterValidation(valueobjects.ErrLinkedByInvalid, "LinkedBy must be a valid email address or 'system'")
}
