import { z } from 'zod';

const placementInputSchema = z.object({
  targetBusinessDomainId: z.string(),
  resultingName: z.string().optional(),
});

export const captureDirectionSchema = z
  .object({
    type: z.enum(['consolidate', 'decompose', 'stay']),
    sourceCapabilityIds: z.array(z.string()),
    placements: z.array(placementInputSchema),
    horizon: z.enum(['now', 'next', 'later']),
    narrative: z.string().transform((s) => s.trim()),
  })
  .superRefine((data, ctx) => {
    const sourceCount = data.sourceCapabilityIds.length;
    if (data.type === 'consolidate' && sourceCount < 2) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ['sourceCapabilityIds'],
        message: 'Consolidate requires at least 2 source physical capabilities.',
      });
    }
    if ((data.type === 'decompose' || data.type === 'stay') && sourceCount !== 1) {
      const verb = data.type === 'decompose' ? 'Decompose' : 'Stay';
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ['sourceCapabilityIds'],
        message: `${verb} requires exactly 1 source physical capability.`,
      });
    }

    if (data.type === 'stay' && data.placements.length > 0) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ['placements'],
        message: 'Stay directions carry no placements.',
      });
    }
    if (data.type === 'consolidate' && data.placements.length !== 1) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ['placements'],
        message: 'Consolidate requires exactly one target placement (N physicals merge into 1).',
      });
    }
    if (data.type === 'decompose' && data.placements.length === 0) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ['placements'],
        message: 'Decompose requires at least one target placement.',
      });
    }
    for (const placement of data.placements) {
      if (!placement.targetBusinessDomainId) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: ['placements'],
          message: 'Every placement needs a target business domain.',
        });
        break;
      }
    }
  });

export type CaptureDirectionFormData = z.infer<typeof captureDirectionSchema>;
