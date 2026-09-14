import { z } from 'zod';

export const attachComponentSchema = z.object({
  parentId: z.string().min(1, 'Select a parent application'),
  kind: z.enum(['composition', 'aggregation']),
});

export type AttachComponentFormData = z.infer<typeof attachComponentSchema>;
