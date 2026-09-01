import { z } from 'zod';

export const ownerReferenceSchema = z.object({
  ownerKind: z.enum(['user', 'team']),
  ownerId: z.string().min(1, 'Select an owner'),
});

export type OwnerReferenceFormData = z.infer<typeof ownerReferenceSchema>;
