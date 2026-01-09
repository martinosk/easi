import { z } from 'zod';
import type { MaturityBounds } from '../../api/types';

export const capabilityNameSchema = z
  .string()
  .min(1, 'Name is required')
  .max(200, 'Name must be 200 characters or less')
  .transform((val) => val.trim())
  .refine((val) => val.length > 0, 'Name is required');

export const capabilityDescriptionSchema = z
  .string()
  .max(1000, 'Description must be 1000 characters or less')
  .transform((val) => val.trim());

const DEFAULT_MATURITY_BOUNDS: MaturityBounds = { min: 0, max: 99 };

export function createCapabilitySchema(bounds: MaturityBounds = DEFAULT_MATURITY_BOUNDS) {
  return z.object({
    name: capabilityNameSchema,
    description: capabilityDescriptionSchema,
    status: z.string().min(1),
    maturityValue: z.number().min(bounds.min).max(bounds.max),
  });
}

export type CreateCapabilityFormData = z.infer<ReturnType<typeof createCapabilitySchema>>;

export function editCapabilitySchema(bounds: MaturityBounds = DEFAULT_MATURITY_BOUNDS) {
  return z.object({
    name: capabilityNameSchema,
    description: capabilityDescriptionSchema,
    status: z.string().min(1),
    maturityValue: z.number().min(bounds.min).max(bounds.max),
    ownershipModel: z.string().transform((val) => val.trim()),
    primaryOwner: z.string().transform((val) => val.trim()),
    eaOwner: z.string(),
  });
}

export type EditCapabilityFormData = z.infer<ReturnType<typeof editCapabilitySchema>>;

export const addTagSchema = z.object({
  tag: z
    .string()
    .min(1, 'Tag name is required')
    .max(100, 'Tag must be 100 characters or less')
    .transform((val) => val.trim())
    .refine((val) => val.length > 0, 'Tag name is required'),
});

export type AddTagFormData = z.infer<typeof addTagSchema>;

export const addExpertSchema = z.object({
  name: z
    .string()
    .min(1, 'Name is required')
    .max(200, 'Name must be 200 characters or less')
    .transform((val) => val.trim())
    .refine((val) => val.length > 0, 'Name is required'),
  role: z
    .string()
    .min(1, 'Role is required')
    .max(200, 'Role must be 200 characters or less')
    .transform((val) => val.trim())
    .refine((val) => val.length > 0, 'Role is required'),
  contact: z
    .string()
    .min(1, 'Contact is required')
    .max(500, 'Contact must be 500 characters or less')
    .transform((val) => val.trim())
    .refine((val) => val.length > 0, 'Contact is required'),
});

export type AddExpertFormData = z.infer<typeof addExpertSchema>;
