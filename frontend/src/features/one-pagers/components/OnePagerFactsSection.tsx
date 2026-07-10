import { Divider, Stack, Title } from '@mantine/core';
import { DetailField } from '../../../components/shared/DetailField';
import { hasLink } from '../../../utils/hateoas';
import { activeCustomFieldsInOrder } from '../factFields';
import { useOnePagerConfiguration } from '../hooks/useOnePagerConfiguration';
import { useOnePagerFacts } from '../hooks/useOnePagerFacts';
import type { CustomField, OnePagerFacts, OnePagerSubjectType } from '../types';
import { FactValueDisplay } from './FactValueDisplay';
import { OnePagerFactsForm } from './OnePagerFactsForm';

function OnePagerFactsReadOnly({ fields, facts }: { fields: CustomField[]; facts: OnePagerFacts }) {
  const valuesByFieldId = new Map(facts.values.map((fieldValue) => [fieldValue.fieldId, fieldValue]));
  return (
    <Stack gap="sm">
      {fields.map((field) => (
        <DetailField key={field.id} label={field.name}>
          <FactValueDisplay field={field} fieldValue={valuesByFieldId.get(field.id)} />
        </DetailField>
      ))}
    </Stack>
  );
}

interface OnePagerFactsSectionProps {
  subjectType: OnePagerSubjectType;
  subjectId: string;
}

export function OnePagerFactsSection({ subjectType, subjectId }: OnePagerFactsSectionProps) {
  const configurationQuery = useOnePagerConfiguration(subjectType);
  const factsQuery = useOnePagerFacts(subjectType, subjectId);
  const configuration = configurationQuery.data;
  const facts = factsQuery.data;

  if (!configuration || !facts) return null;

  const fields = activeCustomFieldsInOrder(configuration);
  if (fields.length === 0) return null;

  const canEdit = hasLink(facts, 'x-record');

  return (
    <Stack gap="sm" data-testid="one-pager-facts-section">
      <Divider />
      <Title order={5}>One-Pager</Title>
      {canEdit ? (
        <OnePagerFactsForm fields={fields} facts={facts} />
      ) : (
        <OnePagerFactsReadOnly fields={fields} facts={facts} />
      )}
    </Stack>
  );
}
