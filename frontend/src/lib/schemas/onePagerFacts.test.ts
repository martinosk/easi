import { describe, expect, it } from 'vitest';
import {
  buildOnePagerFactsSchema,
  emptyFactValue,
  factEnvelope,
  factFormDefaults,
  isFactValueEmpty,
  VALUE_ENVELOPE_VERSION,
  type FactFieldDefinition,
} from './onePagerFacts';

const textField: FactFieldDefinition = { id: 'f-text', type: 'text' };
const numberField: FactFieldDefinition = { id: 'f-number', type: 'number' };
const dateField: FactFieldDefinition = { id: 'f-date', type: 'date' };
const linkField: FactFieldDefinition = { id: 'f-link', type: 'link' };
const selectionField: FactFieldDefinition = {
  id: 'f-selection',
  type: 'selection',
  options: [
    { id: 'opt-1', label: 'Tier 1', active: true },
    { id: 'opt-2', label: 'Tier 2', active: false },
  ],
};
const contactField: FactFieldDefinition = { id: 'f-contact', type: 'contact-person' };

const allFields = [textField, numberField, dateField, linkField, selectionField, contactField];

function emptyValues() {
  return {
    'f-text': '',
    'f-number': '' as const,
    'f-date': '',
    'f-link': { label: '', url: '' },
    'f-selection': '',
    'f-contact': { name: '', email: '', company: '' },
  };
}

function parseWith(overrides: Record<string, unknown>) {
  return buildOnePagerFactsSchema(allFields).safeParse({ ...emptyValues(), ...overrides });
}

describe('buildOnePagerFactsSchema', () => {
  it('accepts a form where every field is empty', () => {
    expect(parseWith({}).success).toBe(true);
  });

  describe('text', () => {
    it('accepts a non-empty text value', () => {
      expect(parseWith({ 'f-text': 'Runs on shared Kubernetes cluster' }).success).toBe(true);
    });

    it('accepts text at exactly 2000 characters', () => {
      expect(parseWith({ 'f-text': 'a'.repeat(2000) }).success).toBe(true);
    });

    it('rejects a whitespace-only text value', () => {
      expect(parseWith({ 'f-text': '   ' }).success).toBe(false);
    });

    it('rejects text exceeding 2000 characters', () => {
      expect(parseWith({ 'f-text': 'a'.repeat(2001) }).success).toBe(false);
    });
  });

  describe('number', () => {
    it('accepts a finite number', () => {
      expect(parseWith({ 'f-number': 42.5 }).success).toBe(true);
    });

    it('accepts zero and negative numbers', () => {
      expect(parseWith({ 'f-number': 0 }).success).toBe(true);
      expect(parseWith({ 'f-number': -3 }).success).toBe(true);
    });

    it('rejects NaN', () => {
      expect(parseWith({ 'f-number': Number.NaN }).success).toBe(false);
    });

    it('rejects Infinity', () => {
      expect(parseWith({ 'f-number': Number.POSITIVE_INFINITY }).success).toBe(false);
      expect(parseWith({ 'f-number': Number.NEGATIVE_INFINITY }).success).toBe(false);
    });
  });

  describe('date', () => {
    it('accepts an ISO date', () => {
      expect(parseWith({ 'f-date': '2026-03-01' }).success).toBe(true);
    });

    it('rejects a non-ISO date', () => {
      expect(parseWith({ 'f-date': 'March 1st' }).success).toBe(false);
    });

    it('rejects an impossible calendar date', () => {
      expect(parseWith({ 'f-date': '2026-02-30' }).success).toBe(false);
    });
  });

  describe('link', () => {
    it('accepts a label with an absolute https url', () => {
      expect(parseWith({ 'f-link': { label: 'MSA', url: 'https://contracts.example.com' } }).success).toBe(true);
    });

    it('accepts an absolute http url', () => {
      expect(parseWith({ 'f-link': { label: 'MSA', url: 'http://contracts.example.com' } }).success).toBe(true);
    });

    it('rejects a non-http scheme', () => {
      expect(parseWith({ 'f-link': { label: 'MSA', url: 'ftp://x' } }).success).toBe(false);
    });

    it('rejects a relative path', () => {
      expect(parseWith({ 'f-link': { label: 'MSA', url: '/contracts/msa' } }).success).toBe(false);
    });

    it('rejects a url without a label', () => {
      expect(parseWith({ 'f-link': { label: '  ', url: 'https://contracts.example.com' } }).success).toBe(false);
    });

    it('rejects a label without a url', () => {
      expect(parseWith({ 'f-link': { label: 'MSA', url: '' } }).success).toBe(false);
    });

    it('rejects a label exceeding 200 characters', () => {
      expect(parseWith({ 'f-link': { label: 'a'.repeat(201), url: 'https://x.example' } }).success).toBe(false);
    });

    it('rejects a url exceeding 2048 characters', () => {
      const url = `https://x.example/${'a'.repeat(2048)}`;
      expect(parseWith({ 'f-link': { label: 'MSA', url } }).success).toBe(false);
    });
  });

  describe('selection', () => {
    it('accepts an active option', () => {
      expect(parseWith({ 'f-selection': 'opt-1' }).success).toBe(true);
    });

    it('rejects an option not defined on the field', () => {
      expect(parseWith({ 'f-selection': 'opt-unknown' }).success).toBe(false);
    });

    it('rejects a retired option that is not the currently recorded value', () => {
      expect(parseWith({ 'f-selection': 'opt-2' }).success).toBe(false);
    });

    it('accepts a retired option when it is the currently recorded value', () => {
      const schema = buildOnePagerFactsSchema(allFields, {
        'f-selection': { type: 'selection', version: 1, value: { optionId: 'opt-2' } },
      });
      expect(schema.safeParse({ ...emptyValues(), 'f-selection': 'opt-2' }).success).toBe(true);
    });
  });

  describe('contact person', () => {
    it('accepts a full contact person', () => {
      expect(
        parseWith({ 'f-contact': { name: 'A. Larsen', email: 'al@ext.example', company: 'Ext ApS' } }).success,
      ).toBe(true);
    });

    it('accepts a contact person without a company', () => {
      expect(parseWith({ 'f-contact': { name: 'A. Larsen', email: 'al@ext.example', company: '' } }).success).toBe(
        true,
      );
    });

    it('rejects an empty name when an email is present', () => {
      expect(parseWith({ 'f-contact': { name: '', email: 'al@ext.example', company: '' } }).success).toBe(false);
    });

    it('rejects an invalid email', () => {
      expect(parseWith({ 'f-contact': { name: 'A. Larsen', email: 'not-an-email', company: '' } }).success).toBe(
        false,
      );
    });

    it('rejects a missing email when a name is present', () => {
      expect(parseWith({ 'f-contact': { name: 'A. Larsen', email: '', company: '' } }).success).toBe(false);
    });

    it('rejects a name exceeding 200 characters', () => {
      expect(
        parseWith({ 'f-contact': { name: 'a'.repeat(201), email: 'al@ext.example', company: '' } }).success,
      ).toBe(false);
    });

    it('rejects a company exceeding 200 characters', () => {
      expect(
        parseWith({ 'f-contact': { name: 'A. Larsen', email: 'al@ext.example', company: 'a'.repeat(201) } }).success,
      ).toBe(false);
    });

    it('requires a name when only a company is present', () => {
      expect(parseWith({ 'f-contact': { name: '', email: '', company: 'Ext ApS' } }).success).toBe(false);
    });
  });
});

describe('isFactValueEmpty', () => {
  it('treats blank text, number, date and selection as empty', () => {
    expect(isFactValueEmpty('text', '')).toBe(true);
    expect(isFactValueEmpty('text', '   ')).toBe(true);
    expect(isFactValueEmpty('number', '')).toBe(true);
    expect(isFactValueEmpty('date', '')).toBe(true);
    expect(isFactValueEmpty('selection', '')).toBe(true);
  });

  it('treats filled values as non-empty', () => {
    expect(isFactValueEmpty('text', 'x')).toBe(false);
    expect(isFactValueEmpty('number', 0)).toBe(false);
    expect(isFactValueEmpty('date', '2026-03-01')).toBe(false);
    expect(isFactValueEmpty('selection', 'opt-1')).toBe(false);
  });

  it('treats a link as empty only when label and url are both blank', () => {
    expect(isFactValueEmpty('link', { label: '', url: '' })).toBe(true);
    expect(isFactValueEmpty('link', { label: 'MSA', url: '' })).toBe(false);
  });

  it('treats a contact person as empty only when all parts are blank', () => {
    expect(isFactValueEmpty('contact-person', { name: '', email: '', company: '' })).toBe(true);
    expect(isFactValueEmpty('contact-person', { name: 'A', email: '', company: '' })).toBe(false);
  });
});

describe('factEnvelope', () => {
  it('returns null for empty values', () => {
    expect(factEnvelope(textField, '')).toBeNull();
    expect(factEnvelope(textField, '   ')).toBeNull();
    expect(factEnvelope(numberField, '')).toBeNull();
    expect(factEnvelope(linkField, { label: '', url: '' })).toBeNull();
    expect(factEnvelope(contactField, { name: '', email: '', company: '' })).toBeNull();
  });

  it('builds a trimmed text envelope', () => {
    expect(factEnvelope(textField, '  hello  ')).toEqual({
      type: 'text',
      version: VALUE_ENVELOPE_VERSION,
      value: 'hello',
    });
  });

  it('builds a number envelope', () => {
    expect(factEnvelope(numberField, 42.5)).toEqual({ type: 'number', version: VALUE_ENVELOPE_VERSION, value: 42.5 });
  });

  it('builds a date envelope', () => {
    expect(factEnvelope(dateField, '2026-03-01')).toEqual({
      type: 'date',
      version: VALUE_ENVELOPE_VERSION,
      value: '2026-03-01',
    });
  });

  it('builds a link envelope with trimmed parts', () => {
    expect(factEnvelope(linkField, { label: ' MSA ', url: ' https://contracts.example.com ' })).toEqual({
      type: 'link',
      version: VALUE_ENVELOPE_VERSION,
      value: { label: 'MSA', url: 'https://contracts.example.com' },
    });
  });

  it('builds a selection envelope', () => {
    expect(factEnvelope(selectionField, 'opt-1')).toEqual({
      type: 'selection',
      version: VALUE_ENVELOPE_VERSION,
      value: { optionId: 'opt-1' },
    });
  });

  it('builds a contact person envelope omitting a blank company', () => {
    expect(factEnvelope(contactField, { name: ' A. Larsen ', email: ' al@ext.example ', company: '  ' })).toEqual({
      type: 'contact-person',
      version: VALUE_ENVELOPE_VERSION,
      value: { name: 'A. Larsen', email: 'al@ext.example' },
    });
  });

  it('includes the company when present', () => {
    expect(factEnvelope(contactField, { name: 'A. Larsen', email: 'al@ext.example', company: 'Ext ApS' })).toEqual({
      type: 'contact-person',
      version: VALUE_ENVELOPE_VERSION,
      value: { name: 'A. Larsen', email: 'al@ext.example', company: 'Ext ApS' },
    });
  });
});

describe('factFormDefaults', () => {
  it('returns empty values for fields without a recorded envelope', () => {
    expect(factFormDefaults(allFields, {})).toEqual(emptyValues());
  });

  it('converts recorded envelopes into form values', () => {
    const defaults = factFormDefaults(allFields, {
      'f-text': { type: 'text', version: 1, value: 'hello' },
      'f-number': { type: 'number', version: 1, value: 42.5 },
      'f-date': { type: 'date', version: 1, value: '2026-03-01' },
      'f-link': { type: 'link', version: 1, value: { label: 'MSA', url: 'https://contracts.example.com' } },
      'f-selection': { type: 'selection', version: 1, value: { optionId: 'opt-1' } },
      'f-contact': { type: 'contact-person', version: 1, value: { name: 'A. Larsen', email: 'al@ext.example' } },
    });

    expect(defaults).toEqual({
      'f-text': 'hello',
      'f-number': 42.5,
      'f-date': '2026-03-01',
      'f-link': { label: 'MSA', url: 'https://contracts.example.com' },
      'f-selection': 'opt-1',
      'f-contact': { name: 'A. Larsen', email: 'al@ext.example', company: '' },
    });
  });

  it('falls back to an empty value when an envelope has an unexpected shape', () => {
    const defaults = factFormDefaults([linkField], { 'f-link': { type: 'link', version: 1, value: 'garbage' } });
    expect(defaults).toEqual({ 'f-link': { label: '', url: '' } });
  });
});

describe('emptyFactValue', () => {
  it('returns the neutral form value per field type', () => {
    expect(emptyFactValue('text')).toBe('');
    expect(emptyFactValue('number')).toBe('');
    expect(emptyFactValue('date')).toBe('');
    expect(emptyFactValue('link')).toEqual({ label: '', url: '' });
    expect(emptyFactValue('selection')).toBe('');
    expect(emptyFactValue('contact-person')).toEqual({ name: '', email: '', company: '' });
  });
});
