import { queryKeys } from './queryClient';

/**
 * Centralized registry of cache invalidation effects for mutations.
 *
 * This module defines which query keys should be invalidated after each mutation type.
 * Use these effects with the `invalidateFor` helper in mutation hooks.
 *
 * @example
 * ```ts
 * invalidateFor(queryClient, mutationEffects.capabilities.create({ parentId: 'cap-1' }))
 * ```
 */
export const mutationEffects = {
  components: {
    /**
     * Invalidates queries after creating a component.
     */
    create: () => [
      queryKeys.components.lists(),
    ],

    /**
     * Invalidates queries after updating a component.
     * @param componentId - The updated component's ID
     */
    update: (componentId: string) => [
      queryKeys.components.lists(),
      queryKeys.components.detail(componentId),
      queryKeys.businessDomains.all,
      queryKeys.audit.history(componentId),
    ],

    /**
     * Invalidates queries after deleting a component.
     * @param componentId - The deleted component's ID
     */
    delete: (componentId: string) => [
      queryKeys.components.lists(),
      queryKeys.components.detail(componentId),
    ],

    /**
     * Invalidates queries after adding an expert to a component.
     * @param componentId - The component receiving the expert
     */
    addExpert: (componentId: string) => [
      queryKeys.components.detail(componentId),
      queryKeys.components.lists(),
      queryKeys.audit.history(componentId),
    ],

    /**
     * Invalidates queries after removing an expert from a component.
     * @param componentId - The component losing the expert
     */
    removeExpert: (componentId: string) => [
      queryKeys.components.detail(componentId),
      queryKeys.components.lists(),
      queryKeys.audit.history(componentId),
    ],
  },

  relations: {
    /**
     * Invalidates queries after creating a relation.
     */
    create: () => [
      queryKeys.relations.lists(),
    ],

    /**
     * Invalidates queries after updating a relation.
     * @param relationId - The updated relation's ID
     */
    update: (relationId: string) => [
      queryKeys.relations.lists(),
      queryKeys.relations.detail(relationId),
      queryKeys.audit.history(relationId),
    ],

    /**
     * Invalidates queries after deleting a relation.
     * @param relationId - The deleted relation's ID
     */
    delete: (relationId: string) => [
      queryKeys.relations.lists(),
      queryKeys.relations.detail(relationId),
    ],
  },

  capabilities: {
    /**
     * Invalidates queries after creating a capability.
     * @param context.parentId - Parent capability ID if creating a child capability
     * @param context.businessDomainId - Domain ID if capability is associated with a domain
     */
    create: (context: { parentId?: string; businessDomainId?: string }) => [
      queryKeys.capabilities.lists(),
      ...(context.parentId ? [queryKeys.capabilities.children(context.parentId)] : []),
      ...(context.businessDomainId
        ? [queryKeys.businessDomains.capabilities(context.businessDomainId)]
        : []),
      queryKeys.maturityAnalysis.unlinked(),
    ],

    /**
     * Invalidates queries after updating a capability's core properties.
     * @param capabilityId - The updated capability's ID
     */
    update: (capabilityId: string) => [
      queryKeys.capabilities.lists(),
      queryKeys.capabilities.detail(capabilityId),
      queryKeys.audit.history(capabilityId),
    ],

    /**
     * Invalidates queries after deleting a capability.
     * @param context.id - The deleted capability's ID
     * @param context.parentId - Parent capability ID if capability was nested
     * @param context.domainId - Domain ID if capability was associated with a domain
     */
    delete: (context: { id: string; parentId?: string; domainId?: string }) => [
      queryKeys.capabilities.lists(),
      queryKeys.capabilities.detail(context.id),
      ...(context.parentId ? [queryKeys.capabilities.children(context.parentId)] : []),
      ...(context.domainId ? [queryKeys.businessDomains.capabilities(context.domainId)] : []),
      queryKeys.businessDomains.lists(),
      queryKeys.maturityAnalysis.unlinked(),
    ],

    /**
     * Invalidates queries after assigning a capability to a domain.
     * @param capabilityId - The capability being assigned
     * @param domainId - The target domain
     */
    assignToDomain: (capabilityId: string, domainId: string) => [
      queryKeys.businessDomains.capabilities(domainId),
      queryKeys.businessDomains.detail(domainId),
      queryKeys.capabilities.detail(capabilityId),
      queryKeys.maturityAnalysis.unlinked(),
    ],

    /**
     * Invalidates queries after removing a capability from a domain.
     * @param capabilityId - The capability being unassigned
     * @param domainId - The domain losing the association
     */
    unassignFromDomain: (capabilityId: string, domainId: string) => [
      queryKeys.businessDomains.capabilities(domainId),
      queryKeys.businessDomains.detail(domainId),
      queryKeys.capabilities.detail(capabilityId),
      queryKeys.maturityAnalysis.unlinked(),
    ],

    /**
     * Invalidates queries after changing a capability's parent in the hierarchy.
     * @param context.id - The capability being moved
     * @param context.oldParentId - Previous parent capability ID
     * @param context.newParentId - New parent capability ID
     */
    changeParent: (context: { id: string; oldParentId?: string; newParentId?: string }) => [
      queryKeys.capabilities.detail(context.id),
      ...(context.oldParentId ? [queryKeys.capabilities.children(context.oldParentId)] : []),
      ...(context.newParentId ? [queryKeys.capabilities.children(context.newParentId)] : []),
      queryKeys.capabilities.lists(),
      queryKeys.audit.history(context.id),
    ],

    /**
     * Invalidates queries after creating a dependency between capabilities.
     * @param sourceCapabilityId - The capability that depends on another
     * @param targetCapabilityId - The capability being depended upon
     */
    addDependency: (sourceCapabilityId: string, targetCapabilityId: string) => [
      queryKeys.capabilities.dependencies(),
      queryKeys.capabilities.outgoing(sourceCapabilityId),
      queryKeys.capabilities.incoming(targetCapabilityId),
    ],

    /**
     * Invalidates queries after removing a dependency between capabilities.
     * @param sourceCapabilityId - The capability that had the dependency
     * @param targetCapabilityId - The capability that was depended upon
     */
    removeDependency: (sourceCapabilityId: string, targetCapabilityId: string) => [
      queryKeys.capabilities.dependencies(),
      queryKeys.capabilities.outgoing(sourceCapabilityId),
      queryKeys.capabilities.incoming(targetCapabilityId),
    ],

    /**
     * Invalidates queries after linking a system/component to a capability.
     * @param capabilityId - The capability receiving the link
     * @param componentId - The component being linked
     */
    linkSystem: (capabilityId: string, componentId: string) => [
      queryKeys.capabilities.realizations(capabilityId),
      queryKeys.capabilities.byComponent(componentId),
      queryKeys.capabilities.realizationsByComponents(),
    ],

    /**
     * Invalidates queries after updating a realization (capability-component link).
     * @param capabilityId - The capability whose realization changed
     * @param componentId - The component involved in the realization
     */
    updateRealization: (capabilityId: string, componentId: string) => [
      queryKeys.capabilities.realizations(capabilityId),
      queryKeys.capabilities.byComponent(componentId),
      queryKeys.capabilities.realizationsByComponents(),
    ],

    /**
     * Invalidates queries after deleting a realization.
     * @param capabilityId - The capability that lost the realization
     * @param componentId - The component that was unlinked
     */
    deleteRealization: (capabilityId: string, componentId: string) => [
      queryKeys.capabilities.realizations(capabilityId),
      queryKeys.capabilities.byComponent(componentId),
      queryKeys.capabilities.realizationsByComponents(),
    ],

    /**
     * Invalidates queries after adding an expert to a capability.
     * @param capabilityId - The capability receiving the expert
     */
    addExpert: (capabilityId: string) => [
      queryKeys.capabilities.detail(capabilityId),
      queryKeys.capabilities.lists(),
      queryKeys.audit.history(capabilityId),
    ],

    /**
     * Invalidates queries after adding a tag to a capability.
     * @param capabilityId - The capability receiving the tag
     */
    addTag: (capabilityId: string) => [
      queryKeys.capabilities.detail(capabilityId),
      queryKeys.capabilities.lists(),
      queryKeys.audit.history(capabilityId),
    ],
  },

  businessDomains: {
    /**
     * Invalidates queries after creating a business domain.
     */
    create: () => [
      queryKeys.businessDomains.lists(),
    ],

    /**
     * Invalidates queries after deleting a business domain.
     * @param domainId - The deleted domain's ID
     */
    delete: (domainId: string) => [
      queryKeys.businessDomains.lists(),
      queryKeys.businessDomains.detail(domainId),
      queryKeys.maturityAnalysis.unlinked(),
    ],

    /**
     * Invalidates queries after updating a business domain.
     * @param domainId - The updated domain's ID
     */
    update: (domainId: string) => [
      queryKeys.businessDomains.lists(),
      queryKeys.businessDomains.detail(domainId),
      queryKeys.audit.history(domainId),
    ],

    /**
     * Invalidates queries after associating a capability with a domain.
     * @param domainId - The domain gaining the capability
     * @param capabilityId - The capability being associated
     */
    associateCapability: (domainId: string, capabilityId: string) => [
      queryKeys.businessDomains.capabilities(domainId),
      queryKeys.businessDomains.detail(domainId),
      queryKeys.businessDomains.lists(),
      queryKeys.capabilities.detail(capabilityId),
      queryKeys.maturityAnalysis.unlinked(),
    ],

    /**
     * Invalidates queries after removing a capability from a domain.
     * @param domainId - The domain losing the capability
     * @param capabilityId - The capability being dissociated
     */
    dissociateCapability: (domainId: string, capabilityId: string) => [
      queryKeys.businessDomains.capabilities(domainId),
      queryKeys.businessDomains.detail(domainId),
      queryKeys.businessDomains.lists(),
      queryKeys.capabilities.detail(capabilityId),
      queryKeys.maturityAnalysis.unlinked(),
    ],
  },

  enterpriseCapabilities: {
    /**
     * Invalidates queries after creating an enterprise capability.
     */
    create: () => [
      queryKeys.enterpriseCapabilities.lists(),
    ],

    /**
     * Invalidates queries after deleting an enterprise capability.
     * @param enterpriseCapabilityId - The deleted enterprise capability's ID
     */
    delete: (enterpriseCapabilityId: string) => [
      queryKeys.enterpriseCapabilities.lists(),
      queryKeys.enterpriseCapabilities.detail(enterpriseCapabilityId),
    ],

    /**
     * Invalidates queries after linking a domain capability to an enterprise capability.
     * @param enterpriseCapabilityId - The enterprise capability receiving the link
     */
    link: (enterpriseCapabilityId: string) => [
      queryKeys.enterpriseCapabilities.links(enterpriseCapabilityId),
      queryKeys.enterpriseCapabilities.detail(enterpriseCapabilityId),
      queryKeys.enterpriseCapabilities.lists(),
      queryKeys.maturityAnalysis.candidates(),
      queryKeys.maturityAnalysis.unlinked(),
      queryKeys.enterpriseCapabilities.linkStatuses(),
    ],

    /**
     * Invalidates queries after unlinking a domain capability from an enterprise capability.
     * @param enterpriseCapabilityId - The enterprise capability losing the link
     */
    unlink: (enterpriseCapabilityId: string) => [
      queryKeys.enterpriseCapabilities.links(enterpriseCapabilityId),
      queryKeys.enterpriseCapabilities.detail(enterpriseCapabilityId),
      queryKeys.enterpriseCapabilities.lists(),
      queryKeys.maturityAnalysis.candidates(),
      queryKeys.maturityAnalysis.unlinked(),
      queryKeys.enterpriseCapabilities.linkStatuses(),
    ],

    /**
     * Invalidates queries after setting target maturity on an enterprise capability.
     * @param enterpriseCapabilityId - The enterprise capability with updated target
     */
    setTargetMaturity: (enterpriseCapabilityId: string) => [
      queryKeys.enterpriseCapabilities.detail(enterpriseCapabilityId),
      queryKeys.enterpriseCapabilities.lists(),
      queryKeys.maturityAnalysis.all,
      queryKeys.enterpriseCapabilities.maturityGap(enterpriseCapabilityId),
    ],
  },

  views: {
    create: () => [
      queryKeys.views.lists(),
    ],

    delete: (viewId: string) => [
      queryKeys.views.lists(),
      queryKeys.views.detail(viewId),
    ],

    rename: (viewId: string) => [
      queryKeys.views.lists(),
      queryKeys.views.detail(viewId),
    ],

    setDefault: () => [
      queryKeys.views.lists(),
    ],

    changeVisibility: (viewId: string) => [
      queryKeys.views.lists(),
      queryKeys.views.detail(viewId),
    ],

    updateDetail: (viewId: string) => [
      queryKeys.views.detail(viewId),
    ],
  },

  layouts: {
    upsert: (contextType: string, contextRef: string) => [
      queryKeys.layouts.detail(contextType, contextRef),
    ],

    delete: (contextType: string, contextRef: string) => [
      queryKeys.layouts.detail(contextType, contextRef),
    ],

    updatePreferences: (contextType: string, contextRef: string) => [
      queryKeys.layouts.detail(contextType, contextRef),
    ],

    updateElement: (contextType: string, contextRef: string) => [
      queryKeys.layouts.detail(contextType, contextRef),
    ],
  },

  fitScores: {
    set: (componentId: string) => [
      queryKeys.fitScores.byComponent(componentId),
      queryKeys.strategicFitAnalysis.all,
    ],

    delete: (componentId: string) => [
      queryKeys.fitScores.byComponent(componentId),
      queryKeys.strategicFitAnalysis.all,
    ],
  },

  strategyPillars: {
    batchUpdate: () => [
      queryKeys.metadata.strategyPillarsConfig(),
    ],
  },

  strategyImportance: {
    set: (domainId: string, capabilityId: string) => [
      queryKeys.strategyImportance.byDomainAndCapability(domainId, capabilityId),
      queryKeys.strategyImportance.byDomain(domainId),
      queryKeys.strategyImportance.byCapability(capabilityId),
    ],

    update: (domainId: string, capabilityId: string) => [
      queryKeys.strategyImportance.byDomainAndCapability(domainId, capabilityId),
      queryKeys.strategyImportance.byDomain(domainId),
      queryKeys.strategyImportance.byCapability(capabilityId),
    ],

    remove: (domainId: string, capabilityId: string) => [
      queryKeys.strategyImportance.byDomainAndCapability(domainId, capabilityId),
      queryKeys.strategyImportance.byDomain(domainId),
      queryKeys.strategyImportance.byCapability(capabilityId),
    ],
  },

  maturityScale: {
    update: () => [
      queryKeys.metadata.maturityScale(),
      queryKeys.metadata.maturityLevels(),
    ],

    reset: () => [
      queryKeys.metadata.maturityScale(),
      queryKeys.metadata.maturityLevels(),
    ],
  },
} as const;
