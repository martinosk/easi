import { z } from 'zod';

export const captureDirectionSchema = z.object({
  type: z.enum(['consolidate', 'decompose', 'stay']),
  sourceCapabilityIds: z.array(z.string()).min(1, 'Select at least one source capability.'),
  horizon: z.enum(['now', 'next', 'later']),
  narrative: z.string().transform((s) => s.trim()),
});

export type CaptureDirectionFormData = z.infer<typeof captureDirectionSchema>;

export const editDirectionSchema = z.object({
  horizon: z.enum(['now', 'next', 'later']),
  narrative: z.string(),
});

export type EditDirectionFormData = z.infer<typeof editDirectionSchema>;
