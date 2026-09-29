import { Alert, Box, Button, Group, Loader, Modal, Paper, Stack, Text } from '@mantine/core';
import { DetailField } from '../../../components/shared/DetailField';
import { InlineSelectField } from '../../../components/shared/InlineSelectField';
import { getLinkObject, hasLink } from '../../../utils/hateoas';
import { useActiveUsers } from '../../users/hooks/useUsers';
import { useAssignSteward, useDomainStewardships, useReleaseSteward } from '../hooks/useStewardships';
import type { ConcernStewardship, StewardshipPerson, StewardsTarget } from '../types';
import classes from './StewardsDialog.module.css';

const UNKNOWN_USER = 'Unknown user';

interface StewardsDialogProps {
  target: StewardsTarget;
  onClose: () => void;
}

interface UserOption {
  value: string;
  label: string;
}

function personName(person: StewardshipPerson): string {
  return person.name ?? UNKNOWN_USER;
}

function optionsIncludingSteward(options: UserOption[], steward: StewardshipPerson | null): UserOption[] {
  if (!steward || options.some((option) => option.value === steward.id)) return options;
  return [...options, { value: steward.id, label: personName(steward) }];
}

interface ConcernRowProps {
  domainId: string;
  item: ConcernStewardship;
  fallback: StewardshipPerson | null;
  candidates: UserOption[];
}

function StewardControl({ domainId, item, candidates }: Omit<ConcernRowProps, 'fallback'>) {
  const assign = useAssignSteward(domainId);
  const release = useReleaseSteward(domainId);
  const assignLink = getLinkObject(item, 'x-assign');
  const releaseLink = getLinkObject(item, 'x-release');
  const steward = item.steward;

  if (!assignLink && !steward) {
    return (
      <DetailField label="Steward">
        <Text size="sm">Unassigned</Text>
      </DetailField>
    );
  }

  return (
    <Stack gap="xs">
      <InlineSelectField
        label="Steward"
        value={steward?.id ?? ''}
        options={optionsIncludingSteward(candidates, steward)}
        canEdit={!!assignLink}
        onSave={(stewardId) => assign.mutateAsync({ link: assignLink!, stewardId })}
        editLabel={`Change ${item.label.toLowerCase()} steward`}
        emptyPrompt="Assign a steward"
        testId={`steward-${item.concern}`}
        searchable
        nothingFoundMessage="No active users"
        renderValue={(_, label) => <Text size="sm">{steward ? personName(steward) : label}</Text>}
      />
      {releaseLink && (
        <Group justify="flex-end">
          <Button
            variant="subtle"
            color="red"
            size="compact-xs"
            disabled={release.isPending}
            data-testid={`release-steward-${item.concern}`}
            onClick={() => release.mutate(releaseLink)}
          >
            Release
          </Button>
        </Group>
      )}
    </Stack>
  );
}

function ConcernRow({ domainId, item, fallback, candidates }: ConcernRowProps) {
  return (
    <Paper withBorder p="sm" data-testid={`stewardship-${item.concern}`}>
      <Group justify="space-between" wrap="nowrap" align="flex-start">
        <Stack gap={0}>
          <Text fw={500}>{item.label}</Text>
          <Text size="xs" c="dimmed">
            {item.description}
          </Text>
          {!item.steward && fallback && (
            <Text size="xs" c="dimmed" mt="xs">
              Falls back to {personName(fallback)}
            </Text>
          )}
        </Stack>
        <Box className={classes.stewardControl}>
          <StewardControl domainId={domainId} item={item} candidates={candidates} />
        </Box>
      </Group>
    </Paper>
  );
}

export function StewardsDialog({ target, onClose }: StewardsDialogProps) {
  const { data, isLoading, error } = useDomainStewardships(target.domainId, target.stewardshipsHref);
  const canAssign = data?.data.some((item) => hasLink(item, 'x-assign')) ?? false;
  const { data: users = [] } = useActiveUsers({ enabled: canAssign });
  const candidates = users.map((user) => ({ value: user.id, label: user.name || user.email }));

  return (
    <Modal
      opened
      onClose={onClose}
      title={`Stewards — ${target.domainName}`}
      size="lg"
      centered
      data-testid="stewards-dialog"
    >
      {isLoading && (
        <Group justify="center" p="md">
          <Loader size="sm" />
        </Group>
      )}
      {error && (
        <Alert color="red" title="Could not load stewards">
          {error.message}
        </Alert>
      )}
      {data && (
        <Stack gap="sm">
          {data.data.map((item) => (
            <ConcernRow
              key={item.concern}
              domainId={target.domainId}
              item={item}
              fallback={data.fallback}
              candidates={candidates}
            />
          ))}
        </Stack>
      )}
    </Modal>
  );
}
