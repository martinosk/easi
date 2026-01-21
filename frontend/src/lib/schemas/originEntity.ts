import { z } from 'zod';

const nameSchema = z
  .string()
  .min(1, 'Name is required')
  .max(100, 'Name must be 100 characters or less')
  .transform((val) => val.trim())
  .refine((val) => val.length > 0, 'Name is required');

const notesSchema = z
  .string()
  .max(500, 'Notes must be 500 characters or less')
  .transform((val) => val.trim())
  .optional();

export const createAcquiredEntitySchema = z.object({
  name: nameSchema,
  acquisitionDate: z.string().optional(),
  integrationStatus: z.enum(['NotStarted', 'InProgress', 'Completed', 'OnHold']).optional(),
  notes: notesSchema,
});

export type CreateAcquiredEntityFormData = z.infer<typeof createAcquiredEntitySchema>;

export const editAcquiredEntitySchema = createAcquiredEntitySchema;
export type EditAcquiredEntityFormData = z.infer<typeof editAcquiredEntitySchema>;

export const createVendorSchema = z.object({
  name: nameSchema,
  implementationPartner: z
    .string()
    .max(100, 'Implementation partner must be 100 characters or less')
    .transform((val) => val.trim())
    .optional(),
  notes: notesSchema,
});

export type CreateVendorFormData = z.infer<typeof createVendorSchema>;

export const editVendorSchema = createVendorSchema;
export type EditVendorFormData = z.infer<typeof editVendorSchema>;

export const createInternalTeamSchema = z.object({
  name: nameSchema,
  department: z
    .string()
    .max(100, 'Department must be 100 characters or less')
    .transform((val) => val.trim())
    .optional(),
  contactPerson: z
    .string()
    .max(100, 'Contact person must be 100 characters or less')
    .transform((val) => val.trim())
    .optional(),
  notes: notesSchema,
});

export type CreateInternalTeamFormData = z.infer<typeof createInternalTeamSchema>;

export const editInternalTeamSchema = createInternalTeamSchema;
export type EditInternalTeamFormData = z.infer<typeof editInternalTeamSchema>;
