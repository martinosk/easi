import { z } from 'zod';

export const originEntityNameSchema = z
  .string()
  .min(1, 'Name is required')
  .max(100, 'Name must be 100 characters or less')
  .transform((val) => val.trim())
  .refine((val) => val.length > 0, 'Name is required');

export const originEntityNotesSchema = z
  .string()
  .max(500, 'Notes must be 500 characters or less')
  .transform((val) => val.trim());

export const integrationStatusSchema = z.enum(['NOT_STARTED', 'IN_PROGRESS', 'COMPLETED']);

export const vendorImplementationPartnerSchema = z
  .string()
  .max(100, 'Implementation partner must be 100 characters or less')
  .transform((val) => val.trim());

export const internalTeamDepartmentSchema = z
  .string()
  .max(100, 'Department must be 100 characters or less')
  .transform((val) => val.trim());

export const internalTeamContactPersonSchema = z
  .string()
  .max(100, 'Contact person must be 100 characters or less')
  .transform((val) => val.trim());

const nameSchema = originEntityNameSchema;
const notesSchema = originEntityNotesSchema.optional();

export const createAcquiredEntitySchema = z.object({
  name: nameSchema,
  acquisitionDate: z.string().optional(),
  integrationStatus: z.enum(['NotStarted', 'InProgress', 'Completed', 'OnHold']).optional(),
  notes: notesSchema,
});

export type CreateAcquiredEntityFormData = z.infer<typeof createAcquiredEntitySchema>;

export const createVendorSchema = z.object({
  name: nameSchema,
  implementationPartner: vendorImplementationPartnerSchema.optional(),
  notes: notesSchema,
});

export type CreateVendorFormData = z.infer<typeof createVendorSchema>;

export const createInternalTeamSchema = z.object({
  name: nameSchema,
  department: internalTeamDepartmentSchema.optional(),
  contactPerson: internalTeamContactPersonSchema.optional(),
  notes: notesSchema,
});

export type CreateInternalTeamFormData = z.infer<typeof createInternalTeamSchema>;
