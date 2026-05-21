import { z } from 'zod';

export const createEnterpriseCapabilitySchema = z.object({
  name: z
    .string()
    .min(1, 'Name is required')
    .max(200, 'Name must be 200 characters or less')
    .transform((val) => val.trim())
    .refine((val) => val.length > 0, 'Name is required'),
  description: z
    .string()
    .max(1000, 'Description must be 1000 characters or less')
    .transform((val) => val.trim()),
  category: z
    .string()
    .max(100, 'Category must be 100 characters or less')
    .transform((val) => val.trim()),
});

export type CreateEnterpriseCapabilityFormData = z.infer<typeof createEnterpriseCapabilitySchema>;
