import { z } from 'zod';

export const businessDomainSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, 'Name is required')
    .max(100, 'Name must be 100 characters or less'),
  description: z.string().trim().max(500, 'Description must be 500 characters or less'),
  domainArchitectId: z.string(),
});

export type BusinessDomainFormData = z.infer<typeof businessDomainSchema>;
