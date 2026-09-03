import type { ZodType } from 'zod';
import type {
  AcquiredEntity,
  IntegrationStatus,
  InternalTeam,
  UpdateAcquiredEntityRequest,
  UpdateInternalTeamRequest,
  UpdateVendorRequest,
  Vendor,
} from '../../../api/types';
import {
  internalTeamContactPersonSchema,
  internalTeamDepartmentSchema,
  originEntityNotesSchema,
  vendorImplementationPartnerSchema,
} from '../../../lib/schemas/originEntity';

interface FieldBase<E> {
  key: string;
  label: string;
  editLabel: string;
  emptyPrompt: string;
  read: (entity: E) => string;
}

export interface TextFieldDefinition<E> extends FieldBase<E> {
  kind: 'text' | 'multiline';
  schema: ZodType<string, string>;
}

export interface SelectFieldDefinition<E> extends FieldBase<E> {
  kind: 'select';
  options: { value: string; label: string }[];
}

export interface DateFieldDefinition<E> extends FieldBase<E> {
  kind: 'date';
}

export type FieldDefinition<E> = TextFieldDefinition<E> | SelectFieldDefinition<E> | DateFieldDefinition<E>;

export interface OriginEntityDefinition<E, R> {
  typeLabel: string;
  fields: FieldDefinition<E>[];
  toRequest: (entity: E) => R;
}

export const INTEGRATION_STATUS_OPTIONS: { value: IntegrationStatus; label: string; color: string }[] = [
  { value: 'NOT_STARTED', label: 'Not Started', color: 'gray' },
  { value: 'IN_PROGRESS', label: 'In Progress', color: 'yellow' },
  { value: 'COMPLETED', label: 'Completed', color: 'green' },
];

export function integrationStatusMeta(status: string): { label: string; color: string } {
  return INTEGRATION_STATUS_OPTIONS.find((option) => option.value === status) ?? { label: status, color: 'gray' };
}

const notesField = <E extends { notes?: string }>(): TextFieldDefinition<E> => ({
  kind: 'multiline',
  key: 'notes',
  label: 'Notes',
  editLabel: 'Edit notes',
  emptyPrompt: 'Add notes',
  schema: originEntityNotesSchema,
  read: (entity) => entity.notes ?? '',
});

export const acquiredEntityDefinition: OriginEntityDefinition<AcquiredEntity, UpdateAcquiredEntityRequest> = {
  typeLabel: 'Acquired Entity',
  fields: [
    {
      kind: 'date',
      key: 'acquisitionDate',
      label: 'Acquisition Date',
      editLabel: 'Edit acquisition date',
      emptyPrompt: 'Set an acquisition date',
      read: (entity) => entity.acquisitionDate?.split('T')[0] ?? '',
    },
    {
      kind: 'select',
      key: 'integrationStatus',
      label: 'Integration Status',
      editLabel: 'Edit integration status',
      emptyPrompt: 'Set an integration status',
      options: INTEGRATION_STATUS_OPTIONS,
      read: (entity) => entity.integrationStatus,
    },
    notesField<AcquiredEntity>(),
  ],
  toRequest: (entity) => ({
    name: entity.name,
    acquisitionDate: entity.acquisitionDate?.split('T')[0] || undefined,
    integrationStatus: entity.integrationStatus,
    notes: entity.notes || undefined,
  }),
};

export const vendorDefinition: OriginEntityDefinition<Vendor, UpdateVendorRequest> = {
  typeLabel: 'Vendor',
  fields: [
    {
      kind: 'text',
      key: 'implementationPartner',
      label: 'Implementation Partner',
      editLabel: 'Edit implementation partner',
      emptyPrompt: 'Add an implementation partner',
      schema: vendorImplementationPartnerSchema,
      read: (vendor) => vendor.implementationPartner ?? '',
    },
    notesField<Vendor>(),
  ],
  toRequest: (vendor) => ({
    name: vendor.name,
    implementationPartner: vendor.implementationPartner || undefined,
    notes: vendor.notes || undefined,
  }),
};

export const internalTeamDefinition: OriginEntityDefinition<InternalTeam, UpdateInternalTeamRequest> = {
  typeLabel: 'Internal Team',
  fields: [
    {
      kind: 'text',
      key: 'department',
      label: 'Department',
      editLabel: 'Edit department',
      emptyPrompt: 'Add a department',
      schema: internalTeamDepartmentSchema,
      read: (team) => team.department ?? '',
    },
    {
      kind: 'text',
      key: 'contactPerson',
      label: 'Contact Person',
      editLabel: 'Edit contact person',
      emptyPrompt: 'Add a contact person',
      schema: internalTeamContactPersonSchema,
      read: (team) => team.contactPerson ?? '',
    },
    notesField<InternalTeam>(),
  ],
  toRequest: (team) => ({
    name: team.name,
    department: team.department || undefined,
    contactPerson: team.contactPerson || undefined,
    notes: team.notes || undefined,
  }),
};
