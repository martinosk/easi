import { z } from 'zod';

export const inviteToEditSchema = z.object({
  granteeEmail: z
    .string()
    .min(1, 'Email is required')
    .email('Enter a valid email address')
    .transform((val) => val.trim())
    .refine((val) => val.length > 0, 'Email is required'),
  reason: z
    .string()
    .max(500, 'Reason must be 500 characters or less')
    .transform((val) => val.trim()),
});

export type InviteToEditFormData = z.infer<typeof inviteToEditSchema>;
