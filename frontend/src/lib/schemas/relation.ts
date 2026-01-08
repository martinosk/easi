import { z } from 'zod';

export const relationTypeSchema = z.enum(['Triggers', 'Serves']);

export type RelationType = z.infer<typeof relationTypeSchema>;

export const relationNameSchema = z
  .string()
  .max(200, 'Name must be 200 characters or less')
  .transform((val) => val.trim());

export const relationDescriptionSchema = z
  .string()
  .max(1000, 'Description must be 1000 characters or less')
  .transform((val) => val.trim());

export const createRelationSchema = z
  .object({
    sourceComponentId: z.string().min(1, 'Source component is required'),
    targetComponentId: z.string().min(1, 'Target component is required'),
    relationType: relationTypeSchema,
    name: relationNameSchema,
    description: relationDescriptionSchema,
  })
  .refine((data) => data.sourceComponentId !== data.targetComponentId, {
    message: 'Source and target components must be different',
    path: ['targetComponentId'],
  });

export type CreateRelationFormData = z.infer<typeof createRelationSchema>;

export const editRelationSchema = z.object({
  name: relationNameSchema,
  description: relationDescriptionSchema,
});

export type EditRelationFormData = z.infer<typeof editRelationSchema>;

export const realizationLevelSchema = z.enum(['Full', 'Partial', 'Planned']);

export type RealizationLevelType = z.infer<typeof realizationLevelSchema>;

export const editRealizationSchema = z.object({
  realizationLevel: realizationLevelSchema,
  notes: z
    .string()
    .max(1000, 'Notes must be 1000 characters or less')
    .transform((val) => val.trim()),
});

export type EditRealizationFormData = z.infer<typeof editRealizationSchema>;
