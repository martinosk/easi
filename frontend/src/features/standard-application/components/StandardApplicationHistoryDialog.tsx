import { Alert, Loader, Modal, Stack, Table, Text } from '@mantine/core';
import { useMemo } from 'react';
import type { EnterpriseCapabilityId } from '../../../api/types';
import { useComponents } from '../../components/hooks/useComponents';
import { useStandardApplicationHistory } from '../hooks/useStandardApplication';

interface StandardApplicationHistoryDialogProps {
  enterpriseCapabilityId: EnterpriseCapabilityId;
  opened: boolean;
  onClose: () => void;
}

export function StandardApplicationHistoryDialog({
  enterpriseCapabilityId,
  opened,
  onClose,
}: StandardApplicationHistoryDialogProps) {
  return (
    <Modal opened={opened} onClose={onClose} title="Standard application history" size="lg" centered>
      {opened && <HistoryDialogBody enterpriseCapabilityId={enterpriseCapabilityId} />}
    </Modal>
  );
}

function HistoryDialogBody({ enterpriseCapabilityId }: { enterpriseCapabilityId: EnterpriseCapabilityId }) {
  const { data: history, isLoading, error } = useStandardApplicationHistory(enterpriseCapabilityId, true);
  const { data: components } = useComponents();

  const nameByApplicationId = useMemo(() => {
    const map = new Map<string, string>();
    (components ?? []).forEach((c) => map.set(String(c.id), c.name));
    return map;
  }, [components]);

  if (isLoading) return <Loader size="sm" />;
  if (error) return <Alert color="red">Failed to load history.</Alert>;
  if (!history) return null;

  if (history.entries.length === 0) {
    return (
      <Stack gap="sm">
        <Text c="dimmed">No history yet.</Text>
      </Stack>
    );
  }

  return (
    <Stack gap="sm">
      <Table data-testid="standard-application-history-table">
        <Table.Thead>
          <Table.Tr>
            <Table.Th>Set at</Table.Th>
            <Table.Th>Application</Table.Th>
            <Table.Th>Previously</Table.Th>
            <Table.Th>Narrative</Table.Th>
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {history.entries.map((entry, idx) => (
            <Table.Tr key={`${entry.setAt}-${idx}`}>
              <Table.Td>{new Date(entry.setAt).toLocaleString()}</Table.Td>
              <Table.Td>{nameByApplicationId.get(String(entry.applicationId)) ?? entry.applicationId}</Table.Td>
              <Table.Td>
                {entry.previousApplicationId
                  ? (nameByApplicationId.get(String(entry.previousApplicationId)) ?? entry.previousApplicationId)
                  : '—'}
              </Table.Td>
              <Table.Td>{entry.narrative}</Table.Td>
            </Table.Tr>
          ))}
        </Table.Tbody>
      </Table>
    </Stack>
  );
}
