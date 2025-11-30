# Business Domain Assignment Consistency

## Description
Maintain consistency of business domain assignments when capability hierarchy changes. When an L1 capability assigned to a business domain gets a parent (becoming L2), the business domain should automatically include the new L1 parent instead.

## Problem
Business domains must only contain L1 capabilities (domain invariant). Currently:
1. When an L1 capability with business domain assignment(s) gets a parent assigned, it becomes L2
2. The capability remains in the business domain as L2, violating the invariant
3. A DB constraint (`CHECK (capability_level = 'L1')`) incorrectly enforces this - **invariants must be enforced by the domain model, not the database**

## Requirements

### Remove Database Constraint
- Create migration to remove `CHECK (capability_level = 'L1')` from `domain_capability_assignments` table
- The domain model is responsible for enforcing invariants, not the database

### Event Handler: CapabilityParentChanged
Create a handler that reacts to `CapabilityParentChanged` events:

**Trigger condition:** `OldLevel == "L1"` AND `NewLevel != "L1"`

**Handler logic:**
1. Query `DomainCapabilityAssignmentReadModel.GetByCapabilityID()` to find all business domains the capability is assigned to
2. For each assignment:
   - Load the `BusinessDomainAssignment` aggregate and call `Unassign()`
   - Check if the new parent (from `NewParentID`) is already assigned to that business domain
   - If not assigned, create a new `BusinessDomainAssignment` for the new parent

### Edge Cases
- New parent already assigned to domain: Skip creating duplicate assignment
- Capability in multiple domains: Handle all affected domains
- No assignments exist: Handler completes with no action

## Checklist
- [x] Migration to remove DB constraint
- [x] Event handler implementation
- [x] Unit tests for event handler
- [x] Integration test verifying end-to-end behavior
- [x] User sign-off
