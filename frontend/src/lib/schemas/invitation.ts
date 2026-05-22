import { z } from 'zod';

export const inviteUserSchema = z.object({
  email: z
    .string()
    .min(1, 'Email is required')
    .email('Enter a valid email address')
    .transform((val) => val.trim())
    .refine((val) => val.length > 0, 'Email is required'),
  role: z.enum(['stakeholder', 'architect', 'admin']),
});

export type InviteUserFormData = z.infer<typeof inviteUserSchema>;
