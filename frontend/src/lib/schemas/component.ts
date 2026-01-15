import { z } from 'zod';

export const componentNameSchema = z
  .string()
  .min(1, 'Application name is required')
  .max(200, 'Name must be 200 characters or less')
  .transform((val) => val.trim())
  .refine((val) => val.length > 0, 'Application name is required');

export const componentDescriptionSchema = z
  .string()
  .max(1000, 'Description must be 1000 characters or less')
  .transform((val) => val.trim());

export const createComponentSchema = z.object({
  name: componentNameSchema,
  description: componentDescriptionSchema,
});

export type CreateComponentFormData = z.infer<typeof createComponentSchema>;

export const editComponentSchema = z.object({
  name: componentNameSchema,
  description: componentDescriptionSchema,
});

export type EditComponentFormData = z.infer<typeof editComponentSchema>;

export const addComponentExpertSchema = z.object({
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

export type AddComponentExpertFormData = z.infer<typeof addComponentExpertSchema>;
