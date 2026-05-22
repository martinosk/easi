import { z } from 'zod';

export const businessDomainSchema = z.object({
  name: z
    .string()
    .min(1, 'Name is required')
    .max(100, 'Name must be 100 characters or less')
    .transform((val) => val.trim())
    .refine((val) => val.length > 0, 'Name is required'),
  description: z
    .string()
    .max(500, 'Description must be 500 characters or less')
    .transform((val) => val.trim()),
  domainArchitectId: z.string(),
});

export type BusinessDomainFormData = z.infer<typeof businessDomainSchema>;
