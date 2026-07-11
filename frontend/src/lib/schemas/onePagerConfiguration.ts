import { z } from 'zod';

export const onePagerFieldTypeValues = ['text', 'number', 'date', 'link', 'selection', 'contact-person'] as const;

export const onePagerFieldNameSchema = z
  .string()
  .min(1, 'Name is required')
  .max(100, 'Name must be 100 characters or less')
  .transform((val) => val.trim())
  .refine((val) => val.length > 0, 'Name is required');

export const onePagerHelpTextSchema = z
  .string()
  .max(500, 'Help text must be 500 characters or less')
  .transform((val) => val.trim());

export const onePagerOptionLabelSchema = z
  .string()
  .min(1, 'Option label is required')
  .max(100, 'Option label must be 100 characters or less')
  .transform((val) => val.trim())
  .refine((val) => val.length > 0, 'Option label is required');

function validateSelectionOptions(options: string[], ctx: z.RefinementCtx): void {
  if (options.length === 0) {
    ctx.addIssue({ code: z.ZodIssueCode.custom, message: 'Add at least one option', path: ['options'] });
    return;
  }
  const seen = new Set<string>();
  for (const option of options) {
    const key = option.toLowerCase();
    if (seen.has(key)) {
      ctx.addIssue({ code: z.ZodIssueCode.custom, message: 'Option labels must be unique', path: ['options'] });
      return;
    }
    seen.add(key);
  }
}

export const defineCustomFieldSchema = z
  .object({
    name: onePagerFieldNameSchema,
    fieldType: z.enum(onePagerFieldTypeValues),
    required: z.boolean(),
    helpText: onePagerHelpTextSchema,
    options: z.array(onePagerOptionLabelSchema),
  })
  .superRefine((data, ctx) => {
    if (data.fieldType === 'selection') validateSelectionOptions(data.options, ctx);
  });

export type DefineCustomFieldFormData = z.infer<typeof defineCustomFieldSchema>;

export const renameCustomFieldSchema = z.object({
  name: onePagerFieldNameSchema,
  helpText: onePagerHelpTextSchema,
});

export type RenameCustomFieldFormData = z.infer<typeof renameCustomFieldSchema>;

export const addSelectionOptionSchema = z.object({
  label: onePagerOptionLabelSchema,
});

export type AddSelectionOptionFormData = z.infer<typeof addSelectionOptionSchema>;
