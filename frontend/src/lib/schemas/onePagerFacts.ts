import { z } from 'zod';
import type { onePagerFieldTypeValues } from './onePagerConfiguration';

export const VALUE_ENVELOPE_VERSION = 1;

export type OnePagerFactFieldType = (typeof onePagerFieldTypeValues)[number];

export interface ValueEnvelope {
  type: string;
  version: number;
  value: unknown;
}

export interface FactFieldOption {
  id: string;
  label: string;
  active: boolean;
}

export interface FactFieldDefinition {
  id: string;
  type: OnePagerFactFieldType;
  options?: FactFieldOption[];
}

export interface LinkFactValue {
  label: string;
  url: string;
}

export interface ContactPersonFactValue {
  name: string;
  email: string;
  company: string;
}

export type FactFormValue = string | number | LinkFactValue | ContactPersonFactValue;

export type OnePagerFactsFormValues = Record<string, FactFormValue>;

export type FactEnvelopesByField = Record<string, ValueEnvelope | undefined>;

const MAX_TEXT_LENGTH = 2000;
const MAX_LABEL_LENGTH = 200;
const MAX_URL_LENGTH = 2048;
const MAX_NAME_LENGTH = 200;
const MAX_COMPANY_LENGTH = 200;
const MAX_EMAIL_LENGTH = 255;

const ISO_DATE_PATTERN = /^(\d{4})-(\d{2})-(\d{2})$/;
const EMAIL_PATTERN = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;

function addIssue(ctx: z.RefinementCtx, message: string, path: (string | number)[] = []): void {
  ctx.addIssue({ code: z.ZodIssueCode.custom, message, path });
}

function isIsoDate(value: string): boolean {
  const match = ISO_DATE_PATTERN.exec(value);
  if (!match) return false;
  const [, year, month, day] = match.map(Number);
  const date = new Date(Date.UTC(year, month - 1, day));
  return date.getUTCFullYear() === year && date.getUTCMonth() === month - 1 && date.getUTCDate() === day;
}

function isAbsoluteHttpUrl(value: string): boolean {
  let parsed: URL;
  try {
    parsed = new URL(value);
  } catch {
    return false;
  }
  return (parsed.protocol === 'http:' || parsed.protocol === 'https:') && parsed.host !== '';
}

const textValueSchema = z.string().superRefine((value, ctx) => {
  const trimmed = value.trim();
  if (value !== '' && trimmed === '') addIssue(ctx, 'Text cannot be only whitespace');
  if (trimmed.length > MAX_TEXT_LENGTH) addIssue(ctx, `Text must be ${MAX_TEXT_LENGTH} characters or less`);
});

const numberValueSchema = z.union([
  z.literal(''),
  z.number().refine(Number.isFinite, 'Number must be a finite number'),
]);

const dateValueSchema = z
  .string()
  .refine((value) => value === '' || isIsoDate(value), 'Date must be an ISO date (YYYY-MM-DD)');

function validateLinkLabel(ctx: z.RefinementCtx, label: string): void {
  if (label === '') addIssue(ctx, 'Label is required', ['label']);
  if (label.length > MAX_LABEL_LENGTH) addIssue(ctx, `Label must be ${MAX_LABEL_LENGTH} characters or less`, ['label']);
}

function validateLinkUrl(ctx: z.RefinementCtx, url: string): void {
  if (url === '') {
    addIssue(ctx, 'URL is required', ['url']);
    return;
  }
  if (url.length > MAX_URL_LENGTH) addIssue(ctx, `URL must be ${MAX_URL_LENGTH} characters or less`, ['url']);
  if (!isAbsoluteHttpUrl(url)) addIssue(ctx, 'URL must be an absolute http(s) URL', ['url']);
}

const linkValueSchema = z.object({ label: z.string(), url: z.string() }).superRefine((value, ctx) => {
  const label = value.label.trim();
  const url = value.url.trim();
  if (label === '' && url === '') return;
  validateLinkLabel(ctx, label);
  validateLinkUrl(ctx, url);
});

function validateContactName(ctx: z.RefinementCtx, name: string): void {
  if (name === '') addIssue(ctx, 'Name is required', ['name']);
  if (name.length > MAX_NAME_LENGTH) addIssue(ctx, `Name must be ${MAX_NAME_LENGTH} characters or less`, ['name']);
}

function validateContactEmail(ctx: z.RefinementCtx, email: string): void {
  if (email === '') {
    addIssue(ctx, 'Email is required', ['email']);
    return;
  }
  if (email.length > MAX_EMAIL_LENGTH || !EMAIL_PATTERN.test(email)) {
    addIssue(ctx, 'Email must be a valid email address', ['email']);
  }
}

const contactPersonValueSchema = z
  .object({ name: z.string(), email: z.string(), company: z.string() })
  .superRefine((value, ctx) => {
    const name = value.name.trim();
    const email = value.email.trim();
    const company = value.company.trim();
    if (name === '' && email === '' && company === '') return;
    validateContactName(ctx, name);
    validateContactEmail(ctx, email);
    if (company.length > MAX_COMPANY_LENGTH) {
      addIssue(ctx, `Company must be ${MAX_COMPANY_LENGTH} characters or less`, ['company']);
    }
  });

function selectionOptionIdFrom(envelope: ValueEnvelope | undefined): string | undefined {
  if (envelope?.type !== 'selection') return undefined;
  const payload = envelope.value as { optionId?: unknown } | null | undefined;
  const optionId = payload?.optionId;
  return typeof optionId === 'string' ? optionId : undefined;
}

function selectionValueSchema(field: FactFieldDefinition, currentEnvelope: ValueEnvelope | undefined) {
  const allowed = new Set((field.options ?? []).filter((option) => option.active).map((option) => option.id));
  const currentOptionId = selectionOptionIdFrom(currentEnvelope);
  if (currentOptionId) allowed.add(currentOptionId);
  return z
    .string()
    .refine((value) => value === '' || allowed.has(value), 'Choose one of the options defined for this field');
}

function fieldSchema(field: FactFieldDefinition, currentEnvelope: ValueEnvelope | undefined): z.ZodTypeAny {
  switch (field.type) {
    case 'text':
      return textValueSchema;
    case 'number':
      return numberValueSchema;
    case 'date':
      return dateValueSchema;
    case 'link':
      return linkValueSchema;
    case 'selection':
      return selectionValueSchema(field, currentEnvelope);
    case 'contact-person':
      return contactPersonValueSchema;
  }
}

export function buildOnePagerFactsSchema(fields: FactFieldDefinition[], currentValues: FactEnvelopesByField = {}) {
  const shape: Record<string, z.ZodTypeAny> = {};
  for (const field of fields) {
    shape[field.id] = fieldSchema(field, currentValues[field.id]);
  }
  return z.object(shape);
}

export function emptyFactValue(type: OnePagerFactFieldType): FactFormValue {
  switch (type) {
    case 'link':
      return { label: '', url: '' };
    case 'contact-person':
      return { name: '', email: '', company: '' };
    default:
      return '';
  }
}

function asString(value: unknown): string {
  return typeof value === 'string' ? value : '';
}

function linkFromPayload(payload: unknown): LinkFactValue {
  if (!payload || typeof payload !== 'object') return { label: '', url: '' };
  const raw = payload as { label?: unknown; url?: unknown };
  return { label: asString(raw.label), url: asString(raw.url) };
}

function contactFromPayload(payload: unknown): ContactPersonFactValue {
  if (!payload || typeof payload !== 'object') return { name: '', email: '', company: '' };
  const raw = payload as { name?: unknown; email?: unknown; company?: unknown };
  return { name: asString(raw.name), email: asString(raw.email), company: asString(raw.company) };
}

export function factFormValueFromEnvelope(field: FactFieldDefinition, envelope: ValueEnvelope): FactFormValue {
  switch (field.type) {
    case 'text':
    case 'date':
      return asString(envelope.value);
    case 'number':
      return typeof envelope.value === 'number' ? envelope.value : '';
    case 'link':
      return linkFromPayload(envelope.value);
    case 'selection':
      return selectionOptionIdFrom(envelope) ?? '';
    case 'contact-person':
      return contactFromPayload(envelope.value);
  }
}

export function factFormDefaults(
  fields: FactFieldDefinition[],
  envelopes: FactEnvelopesByField,
): OnePagerFactsFormValues {
  const defaults: OnePagerFactsFormValues = {};
  for (const field of fields) {
    const envelope = envelopes[field.id];
    defaults[field.id] = envelope ? factFormValueFromEnvelope(field, envelope) : emptyFactValue(field.type);
  }
  return defaults;
}

function isBlank(value: FactFormValue): boolean {
  return typeof value !== 'number' && (typeof value !== 'string' || value.trim() === '');
}

export function isFactValueEmpty(type: OnePagerFactFieldType, value: FactFormValue): boolean {
  switch (type) {
    case 'link': {
      const link = value as LinkFactValue;
      return link.label.trim() === '' && link.url.trim() === '';
    }
    case 'contact-person': {
      const contact = value as ContactPersonFactValue;
      return contact.name.trim() === '' && contact.email.trim() === '' && contact.company.trim() === '';
    }
    default:
      return isBlank(value);
  }
}

function envelopePayload(field: FactFieldDefinition, value: FactFormValue): unknown {
  switch (field.type) {
    case 'text':
    case 'date':
      return (value as string).trim();
    case 'number':
      return value as number;
    case 'link': {
      const link = value as LinkFactValue;
      return { label: link.label.trim(), url: link.url.trim() };
    }
    case 'selection':
      return { optionId: value as string };
    case 'contact-person': {
      const contact = value as ContactPersonFactValue;
      const company = contact.company.trim();
      return { name: contact.name.trim(), email: contact.email.trim(), ...(company ? { company } : {}) };
    }
  }
}

export function factEnvelope(field: FactFieldDefinition, value: FactFormValue): ValueEnvelope | null {
  if (isFactValueEmpty(field.type, value)) return null;
  return { type: field.type, version: VALUE_ENVELOPE_VERSION, value: envelopePayload(field, value) };
}
