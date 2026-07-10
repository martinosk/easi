import { Anchor, Text } from '@mantine/core';
import {
  factFormValueFromEnvelope,
  type ContactPersonFactValue,
  type LinkFactValue,
} from '../../../lib/schemas/onePagerFacts';
import type { CustomField, FieldValue } from '../types';

function contactText(contact: ContactPersonFactValue): string {
  const base = `${contact.name} (${contact.email})`;
  return contact.company ? `${base}, ${contact.company}` : base;
}

function selectionText(field: CustomField, fieldValue: FieldValue, optionId: string): string {
  const option = (field.options ?? []).find((candidate) => candidate.id === optionId);
  const label = option?.label ?? fieldValue.displayText;
  return fieldValue.retiredOption ? `${label} (retired)` : label;
}

interface FactValueDisplayProps {
  field: CustomField;
  fieldValue?: FieldValue;
}

export function FactValueDisplay({ field, fieldValue }: FactValueDisplayProps) {
  if (!fieldValue) {
    return (
      <Text size="sm" c="dimmed">
        —
      </Text>
    );
  }
  const formValue = factFormValueFromEnvelope(field, fieldValue.value);
  switch (field.type) {
    case 'link': {
      const link = formValue as LinkFactValue;
      return (
        <Anchor size="sm" href={link.url} target="_blank" rel="noopener noreferrer">
          {link.label}
        </Anchor>
      );
    }
    case 'selection':
      return <Text size="sm">{selectionText(field, fieldValue, formValue as string)}</Text>;
    case 'contact-person':
      return <Text size="sm">{contactText(formValue as ContactPersonFactValue)}</Text>;
    default:
      return <Text size="sm">{String(formValue)}</Text>;
  }
}
