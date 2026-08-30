package api

import (
	"context"

	adServices "easi/backend/internal/architecturedirection/application/services"
	adDomainServices "easi/backend/internal/architecturedirection/domain/services"
	directionAPI "easi/backend/internal/architecturedirection/infrastructure/api"
	archReadModels "easi/backend/internal/architecturemodeling/application/readmodels"
	capReadModels "easi/backend/internal/capabilitymapping/application/readmodels"
	eaReadModels "easi/backend/internal/enterprisearchitecture/application/readmodels"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/onepagers/application/ports"
)

type onePagerRelationModels struct {
	capabilities  *capReadModels.CapabilityReadModel
	realizations  *capReadModels.RealizationReadModel
	dependencies  *capReadModels.DependencyReadModel
	assignments   *capReadModels.DomainCapabilityAssignmentReadModel
	applications  *archReadModels.ApplicationComponentReadModel
	builtBy       *archReadModels.BuiltByRelationshipReadModel
	purchasedFrom *archReadModels.PurchasedFromRelationshipReadModel
	acquiredVia   *archReadModels.AcquiredViaRelationshipReadModel
	componentRels *archReadModels.ComponentRelationReadModel
	composition   *adServices.CompositionService
}

func newOnePagerRelationModels(db *database.TenantAwareDB) onePagerRelationModels {
	return onePagerRelationModels{
		capabilities:  capReadModels.NewCapabilityReadModel(db),
		realizations:  capReadModels.NewRealizationReadModel(db),
		dependencies:  capReadModels.NewDependencyReadModel(db),
		assignments:   capReadModels.NewDomainCapabilityAssignmentReadModel(db),
		applications:  archReadModels.NewApplicationComponentReadModel(db),
		builtBy:       archReadModels.NewBuiltByRelationshipReadModel(db),
		purchasedFrom: archReadModels.NewPurchasedFromRelationshipReadModel(db),
		acquiredVia:   archReadModels.NewAcquiredViaRelationshipReadModel(db),
		componentRels: archReadModels.NewComponentRelationReadModel(db),
		composition:   directionAPI.NewCompositionService(db),
	}
}

type namesResolver = func(context.Context, []string) (map[string]string, error)

func mapReferences[E any](edges []E, toReference func(E) ports.Reference) ports.ReferenceListValue {
	references := make([]ports.Reference, len(edges))
	for i, edge := range edges {
		references[i] = toReference(edge)
	}
	return ports.ReferenceListValue{References: references}
}

func namedReferences(ids []string, subjectType string, names map[string]string) ports.ReferenceListValue {
	references := make([]ports.Reference, len(ids))
	for i, id := range ids {
		references[i] = ports.Reference{ID: id, Label: names[id], SubjectType: subjectType}
	}
	return ports.ReferenceListValue{References: references}
}

func namesByID[T any](ctx context.Context, fetch func(context.Context, []string) ([]T, error), ids []string, idOf, nameOf func(T) string) (map[string]string, error) {
	dtos, err := fetch(ctx, ids)
	if err != nil {
		return nil, err
	}
	names := make(map[string]string, len(dtos))
	for i := range dtos {
		names[idOf(dtos[i])] = nameOf(dtos[i])
	}
	return names, nil
}

func resolveNamedReferences(ctx context.Context, ids []string, subjectType string, names namesResolver) (ports.ReferenceListValue, error) {
	if len(ids) == 0 {
		return ports.ReferenceListValue{}, nil
	}
	resolved, err := names(ctx, ids)
	if err != nil {
		return ports.ReferenceListValue{}, err
	}
	return namedReferences(ids, subjectType, resolved), nil
}

func resolveEdgeReferences[E any](ctx context.Context, fetch func(context.Context, string) ([]E, error), subjectID string, targetID func(E) string, subjectType string, names namesResolver) (ports.ReferenceListValue, error) {
	edges, err := fetch(ctx, subjectID)
	if err != nil {
		return ports.ReferenceListValue{}, err
	}
	ids := make([]string, len(edges))
	for i, edge := range edges {
		ids[i] = targetID(edge)
	}
	return resolveNamedReferences(ctx, ids, subjectType, names)
}

func (m onePagerRelationModels) capabilityNames(ctx context.Context, ids []string) (map[string]string, error) {
	return namesByID(ctx, m.capabilities.GetByIDs, ids,
		func(c capReadModels.CapabilityDTO) string { return c.ID },
		func(c capReadModels.CapabilityDTO) string { return c.Name })
}

func (m onePagerRelationModels) applicationNames(ctx context.Context, ids []string) (map[string]string, error) {
	return namesByID(ctx, m.applications.GetByIDs, ids,
		func(a archReadModels.ApplicationComponentDTO) string { return a.ID },
		func(a archReadModels.ApplicationComponentDTO) string { return a.Name })
}

func (m onePagerRelationModels) capabilityRelations() []relationBinding[capReadModels.CapabilityDTO] {
	return []relationBinding[capReadModels.CapabilityDTO]{
		{entryID: "realizing-applications", resolve: m.realizingApplications},
		{entryID: "business-domains", resolve: m.businessDomains},
		{entryID: "parent-capability", resolve: m.parentCapability},
		{entryID: "child-capabilities", resolve: m.childCapabilities},
		{entryID: "depends-on", resolve: m.dependsOn},
	}
}

func (m onePagerRelationModels) realizingApplications(ctx context.Context, dto *capReadModels.CapabilityDTO) (ports.ReferenceListValue, error) {
	edges, err := m.realizations.GetByCapabilityID(ctx, dto.ID)
	if err != nil {
		return ports.ReferenceListValue{}, err
	}
	return mapReferences(edges, func(e capReadModels.RealizationDTO) ports.Reference {
		return ports.Reference{ID: e.ComponentID, Label: e.ComponentName, SubjectType: "application"}
	}), nil
}

func (m onePagerRelationModels) businessDomains(ctx context.Context, dto *capReadModels.CapabilityDTO) (ports.ReferenceListValue, error) {
	edges, err := m.assignments.GetByCapabilityID(ctx, dto.ID)
	if err != nil {
		return ports.ReferenceListValue{}, err
	}
	return mapReferences(edges, func(e capReadModels.AssignmentDTO) ports.Reference {
		return ports.Reference{ID: e.BusinessDomainID, Label: e.BusinessDomainName}
	}), nil
}

func (m onePagerRelationModels) parentCapability(ctx context.Context, dto *capReadModels.CapabilityDTO) (ports.ReferenceListValue, error) {
	if dto.ParentID == "" {
		return ports.ReferenceListValue{}, nil
	}
	return resolveNamedReferences(ctx, []string{dto.ParentID}, "capability", m.capabilityNames)
}

func (m onePagerRelationModels) childCapabilities(ctx context.Context, dto *capReadModels.CapabilityDTO) (ports.ReferenceListValue, error) {
	children, err := m.capabilities.GetChildren(ctx, dto.ID)
	if err != nil {
		return ports.ReferenceListValue{}, err
	}
	return mapReferences(children, func(c capReadModels.CapabilityDTO) ports.Reference {
		return ports.Reference{ID: c.ID, Label: c.Name, SubjectType: "capability"}
	}), nil
}

func (m onePagerRelationModels) dependsOn(ctx context.Context, dto *capReadModels.CapabilityDTO) (ports.ReferenceListValue, error) {
	return resolveEdgeReferences(ctx, m.dependencies.GetOutgoing, dto.ID,
		func(e capReadModels.DependencyDTO) string { return e.TargetCapabilityID }, "capability", m.capabilityNames)
}

func (m onePagerRelationModels) enterpriseCapabilityRelations() []relationBinding[eaReadModels.EnterpriseCapabilityDTO] {
	return []relationBinding[eaReadModels.EnterpriseCapabilityDTO]{
		{entryID: "included-capabilities", resolve: m.includedCapabilities},
	}
}

func (m onePagerRelationModels) includedCapabilities(ctx context.Context, dto *eaReadModels.EnterpriseCapabilityDTO) (ports.ReferenceListValue, error) {
	result, err := m.composition.CompositionForEC(ctx, dto.ID)
	if err != nil {
		return ports.ReferenceListValue{}, err
	}
	references := make([]ports.Reference, 0, len(result.Resolved))
	for _, resolved := range result.Resolved {
		if resolved.Role == adDomainServices.RoleCarvedOut {
			continue
		}
		references = append(references, ports.Reference{ID: resolved.Node.ID, Label: resolved.Node.Name, SubjectType: "capability"})
	}
	return ports.ReferenceListValue{References: references}, nil
}

func (m onePagerRelationModels) applicationRelations() []relationBinding[archReadModels.ApplicationComponentDTO] {
	return []relationBinding[archReadModels.ApplicationComponentDTO]{
		{entryID: "realized-capabilities", resolve: m.realizedCapabilities},
		{entryID: "built-by", resolve: m.builtByTeam},
		{entryID: "purchased-from", resolve: m.purchasedFromVendor},
		{entryID: "acquired-via", resolve: m.acquiredViaEntity},
		{entryID: "component-relations", resolve: m.componentRelations},
	}
}

func (m onePagerRelationModels) realizedCapabilities(ctx context.Context, dto *archReadModels.ApplicationComponentDTO) (ports.ReferenceListValue, error) {
	return resolveEdgeReferences(ctx, m.realizations.GetByComponentID, dto.ID,
		func(e capReadModels.RealizationDTO) string { return e.CapabilityID }, "capability", m.capabilityNames)
}

func (m onePagerRelationModels) builtByTeam(ctx context.Context, dto *archReadModels.ApplicationComponentDTO) (ports.ReferenceListValue, error) {
	edges, err := m.builtBy.GetByComponentID(ctx, dto.ID)
	if err != nil {
		return ports.ReferenceListValue{}, err
	}
	return mapReferences(edges, func(e archReadModels.BuiltByRelationshipDTO) ports.Reference {
		return ports.Reference{ID: e.InternalTeamID, Label: e.InternalTeamName, SubjectType: "internal-team"}
	}), nil
}

func (m onePagerRelationModels) purchasedFromVendor(ctx context.Context, dto *archReadModels.ApplicationComponentDTO) (ports.ReferenceListValue, error) {
	edges, err := m.purchasedFrom.GetByComponentID(ctx, dto.ID)
	if err != nil {
		return ports.ReferenceListValue{}, err
	}
	return mapReferences(edges, func(e archReadModels.PurchasedFromRelationshipDTO) ports.Reference {
		return ports.Reference{ID: e.VendorID, Label: e.VendorName, SubjectType: "vendor"}
	}), nil
}

func (m onePagerRelationModels) acquiredViaEntity(ctx context.Context, dto *archReadModels.ApplicationComponentDTO) (ports.ReferenceListValue, error) {
	edges, err := m.acquiredVia.GetByComponentID(ctx, dto.ID)
	if err != nil {
		return ports.ReferenceListValue{}, err
	}
	return mapReferences(edges, func(e archReadModels.AcquiredViaRelationshipDTO) ports.Reference {
		return ports.Reference{ID: e.AcquiredEntityID, Label: e.AcquiredEntityName, SubjectType: "acquired-entity"}
	}), nil
}

func (m onePagerRelationModels) componentRelations(ctx context.Context, dto *archReadModels.ApplicationComponentDTO) (ports.ReferenceListValue, error) {
	return resolveEdgeReferences(ctx, m.componentRels.GetBySourceID, dto.ID,
		func(e archReadModels.ComponentRelationDTO) string { return e.TargetComponentID }, "application", m.applicationNames)
}

func (m onePagerRelationModels) acquiredEntityRelations() []relationBinding[archReadModels.AcquiredEntityDTO] {
	return []relationBinding[archReadModels.AcquiredEntityDTO]{
		{entryID: "acquired-applications", resolve: m.acquiredApplications},
	}
}

func (m onePagerRelationModels) acquiredApplications(ctx context.Context, dto *archReadModels.AcquiredEntityDTO) (ports.ReferenceListValue, error) {
	edges, err := m.acquiredVia.GetByEntityID(ctx, dto.ID)
	if err != nil {
		return ports.ReferenceListValue{}, err
	}
	return mapReferences(edges, componentReference), nil
}

func (m onePagerRelationModels) vendorRelations() []relationBinding[archReadModels.VendorDTO] {
	return []relationBinding[archReadModels.VendorDTO]{
		{entryID: "purchased-applications", resolve: m.purchasedApplications},
	}
}

func (m onePagerRelationModels) purchasedApplications(ctx context.Context, dto *archReadModels.VendorDTO) (ports.ReferenceListValue, error) {
	edges, err := m.purchasedFrom.GetByVendorID(ctx, dto.ID)
	if err != nil {
		return ports.ReferenceListValue{}, err
	}
	return mapReferences(edges, func(e archReadModels.PurchasedFromRelationshipDTO) ports.Reference {
		return ports.Reference{ID: e.ComponentID, Label: e.ComponentName, SubjectType: "application"}
	}), nil
}

func (m onePagerRelationModels) internalTeamRelations() []relationBinding[archReadModels.InternalTeamDTO] {
	return []relationBinding[archReadModels.InternalTeamDTO]{
		{entryID: "built-applications", resolve: m.builtApplications},
	}
}

func (m onePagerRelationModels) builtApplications(ctx context.Context, dto *archReadModels.InternalTeamDTO) (ports.ReferenceListValue, error) {
	edges, err := m.builtBy.GetByTeamID(ctx, dto.ID)
	if err != nil {
		return ports.ReferenceListValue{}, err
	}
	return mapReferences(edges, func(e archReadModels.BuiltByRelationshipDTO) ports.Reference {
		return ports.Reference{ID: e.ComponentID, Label: e.ComponentName, SubjectType: "application"}
	}), nil
}

func componentReference(e archReadModels.AcquiredViaRelationshipDTO) ports.Reference {
	return ports.Reference{ID: e.ComponentID, Label: e.ComponentName, SubjectType: "application"}
}
