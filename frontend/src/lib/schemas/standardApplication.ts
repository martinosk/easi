import { z } from 'zod';

export const setStandardApplicationSchema = z.object({
  applicationId: z.string().min(1, 'Pick an application.'),
  narrative: z
    .string()
    .transform((s) => s.trim())
    .pipe(z.string().min(1, 'A narrative is required.').max(1000, 'Narrative is too long.')),
});

export type SetStandardApplicationFormValues = z.infer<typeof setStandardApplicationSchema>;
