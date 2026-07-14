import {
  Alert,
  Anchor,
  Button,
  Center,
  Container,
  Group,
  Loader,
  Paper,
  Stack,
  Table,
  Text,
  Title,
  UnstyledButton,
} from '@mantine/core';
import { IconChevronDown, IconChevronUp } from '@tabler/icons-react';
import { useState } from 'react';
import { Link } from 'react-router-dom';
import { ApiError } from '../../../api/types';
import { ROUTES } from '../../../routes/routePaths';
import { InviteToEditButton } from '../../edit-grants/components/InviteToEditButton';
import { subjectTypeLabel } from '../../one-pagers/subjectTypes';
import { CompletenessBadge } from '../components/CompletenessBadge';
import { useOnePagerQualityList } from '../hooks/useOnePagerQualityList';
import { subjectArtifactType } from '../subjectArtifactType';
import type { OnePagerQualityOrder, OnePagerQualityRow, OnePagerQualitySort } from '../types';

const DEFAULT_SORT: OnePagerQualitySort = 'completeness';
const DEFAULT_ORDER: OnePagerQualityOrder = 'asc';

interface SortColumn {
  key: OnePagerQualitySort;
  label: string;
}

const NAME_COLUMN: SortColumn = { key: 'name', label: 'Name' };

const COLUMNS_AFTER_TYPE: SortColumn[] = [
  { key: 'completeness', label: 'Completeness' },
  { key: 'creator', label: 'Creator' },
  { key: 'created', label: 'Created' },
  { key: 'updated', label: 'Updated' },
];

function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
}

function onePagerDetailPath(subjectType: string, subjectId: string): string {
  return ROUTES.ONE_PAGER_DETAIL.replace(':subjectType', subjectType).replace(':subjectId', subjectId);
}

function PageShell({ children }: { children: React.ReactNode }) {
  return (
    <Container size="xl" py="xl">
      {children}
    </Container>
  );
}

function SortIndicator({ active, order }: { active: boolean; order: OnePagerQualityOrder }) {
  if (!active) return null;
  return order === 'asc' ? <IconChevronUp size={14} stroke={1.75} /> : <IconChevronDown size={14} stroke={1.75} />;
}

interface SortableHeaderProps {
  column: SortColumn;
  sort: OnePagerQualitySort;
  order: OnePagerQualityOrder;
  onSort: (key: OnePagerQualitySort) => void;
}

function SortableHeader({ column, sort, order, onSort }: SortableHeaderProps) {
  const active = sort === column.key;
  return (
    <Table.Th>
      <UnstyledButton onClick={() => onSort(column.key)} data-testid={`sort-header-${column.key}`}>
        <Group gap="xs" wrap="nowrap">
          <Text fw={600} size="sm">
            {column.label}
          </Text>
          <SortIndicator active={active} order={order} />
        </Group>
      </UnstyledButton>
    </Table.Th>
  );
}

function QualityRowActions({ row }: { row: OnePagerQualityRow }) {
  const artifactType = subjectArtifactType(row.subjectType);
  if (!artifactType) return null;
  return <InviteToEditButton resource={row} artifactType={artifactType} artifactId={row.subjectId} />;
}

function QualityRow({ row }: { row: OnePagerQualityRow }) {
  return (
    <Table.Tr data-testid={`quality-row-${row.subjectId}`}>
      <Table.Td>
        <Anchor component={Link} to={onePagerDetailPath(row.subjectType, row.subjectId)} fw={500}>
          {row.name}
        </Anchor>
      </Table.Td>
      <Table.Td>
        <Text size="sm" c="dimmed">
          {subjectTypeLabel(row.subjectType)}
        </Text>
      </Table.Td>
      <Table.Td>
        <CompletenessBadge completeness={row.completeness} />
      </Table.Td>
      <Table.Td>
        <Text size="sm">{row.creatorEmail}</Text>
      </Table.Td>
      <Table.Td>
        <Text size="xs" c="dimmed">
          {formatDate(row.createdAt)}
        </Text>
      </Table.Td>
      <Table.Td>
        <Text size="xs" c="dimmed">
          {formatDate(row.lastUpdatedAt)}
        </Text>
      </Table.Td>
      <Table.Td>
        <QualityRowActions row={row} />
      </Table.Td>
    </Table.Tr>
  );
}

interface QualityTableProps {
  rows: OnePagerQualityRow[];
  sort: OnePagerQualitySort;
  order: OnePagerQualityOrder;
  onSort: (key: OnePagerQualitySort) => void;
}

function QualityTable({ rows, sort, order, onSort }: QualityTableProps) {
  return (
    <Paper shadow="sm" radius="lg" withBorder>
      <Table data-testid="one-pager-quality-table" striped highlightOnHover verticalSpacing="sm">
        <Table.Thead>
          <Table.Tr>
            <SortableHeader column={NAME_COLUMN} sort={sort} order={order} onSort={onSort} />
            <Table.Th>Type</Table.Th>
            {COLUMNS_AFTER_TYPE.map((column) => (
              <SortableHeader key={column.key} column={column} sort={sort} order={order} onSort={onSort} />
            ))}
            <Table.Th>Actions</Table.Th>
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {rows.map((row) => (
            <QualityRow key={`${row.subjectType}-${row.subjectId}`} row={row} />
          ))}
        </Table.Tbody>
      </Table>
    </Paper>
  );
}

function EmptyState() {
  return (
    <Stack align="center" gap="md" py="xl">
      <Title order={4}>No one-pagers found</Title>
      <Text c="dimmed">There is nothing to show for the current sort.</Text>
    </Stack>
  );
}

function ErrorState({ error }: { error: unknown }) {
  const isForbidden = error instanceof ApiError && error.statusCode === 403;
  return (
    <Alert color="red" title={isForbidden ? 'Access denied' : 'Failed to load'} data-testid="one-pager-quality-error">
      {isForbidden
        ? 'You do not have permission to view one-pager quality for any subject type.'
        : 'Something went wrong while loading one-pager quality data. Please try again.'}
    </Alert>
  );
}

interface PaginationControlsProps {
  hasMore: boolean;
  canGoBack: boolean;
  onNext: () => void;
  onPrev: () => void;
}

function PaginationControls({ hasMore, canGoBack, onNext, onPrev }: PaginationControlsProps) {
  return (
    <Group justify="flex-end" mt="md" gap="sm">
      <Button variant="default" size="sm" disabled={!canGoBack} onClick={onPrev} data-testid="pagination-prev">
        Previous
      </Button>
      <Button variant="default" size="sm" disabled={!hasMore} onClick={onNext} data-testid="pagination-next">
        Next
      </Button>
    </Group>
  );
}

interface OnePagerQualityContentProps {
  isLoading: boolean;
  error: unknown;
  rows: OnePagerQualityRow[];
  sort: OnePagerQualitySort;
  order: OnePagerQualityOrder;
  onSort: (key: OnePagerQualitySort) => void;
  hasMore: boolean;
  canGoBack: boolean;
  onNext: () => void;
  onPrev: () => void;
}

function OnePagerQualityContent({
  isLoading,
  error,
  rows,
  sort,
  order,
  onSort,
  hasMore,
  canGoBack,
  onNext,
  onPrev,
}: OnePagerQualityContentProps) {
  if (isLoading) {
    return (
      <Center py="xl">
        <Stack align="center" gap="md">
          <Loader />
          <Text>Loading one-pager quality...</Text>
        </Stack>
      </Center>
    );
  }

  if (error) {
    return <ErrorState error={error} />;
  }

  if (rows.length === 0) {
    return <EmptyState />;
  }

  return (
    <>
      <QualityTable rows={rows} sort={sort} order={order} onSort={onSort} />
      <PaginationControls hasMore={hasMore} canGoBack={canGoBack} onNext={onNext} onPrev={onPrev} />
    </>
  );
}

function useCursorPagination() {
  const [cursorHistory, setCursorHistory] = useState<(string | undefined)[]>([]);
  const [cursor, setCursor] = useState<string | undefined>(undefined);

  const goNext = (nextCursor: string) => {
    setCursorHistory((prev) => [...prev, cursor]);
    setCursor(nextCursor);
  };

  const goPrev = () => {
    setCursorHistory((prev) => {
      if (prev.length === 0) return prev;
      setCursor(prev[prev.length - 1]);
      return prev.slice(0, -1);
    });
  };

  const reset = () => {
    setCursorHistory([]);
    setCursor(undefined);
  };

  return { cursor, canGoBack: cursorHistory.length > 0, goNext, goPrev, reset };
}

export function OnePagerQualityPage() {
  const [sort, setSort] = useState<OnePagerQualitySort>(DEFAULT_SORT);
  const [order, setOrder] = useState<OnePagerQualityOrder>(DEFAULT_ORDER);
  const pagination = useCursorPagination();

  const { data, isLoading, error } = useOnePagerQualityList({ sort, order, cursor: pagination.cursor });

  const handleSort = (key: OnePagerQualitySort) => {
    if (key === sort) {
      setOrder((current) => (current === 'asc' ? 'desc' : 'asc'));
    } else {
      setSort(key);
      setOrder('asc');
    }
    pagination.reset();
  };

  const handleNext = () => {
    if (data?.pagination.hasMore && data.pagination.cursor) {
      pagination.goNext(data.pagination.cursor);
    }
  };

  return (
    <PageShell>
      <Stack gap="xs" mb="xl">
        <Title order={1}>One-Pager Quality</Title>
        <Text c="dimmed">Track one-pager completeness across capabilities, applications, and other subject types.</Text>
      </Stack>
      <OnePagerQualityContent
        isLoading={isLoading}
        error={error}
        rows={data?.data ?? []}
        sort={sort}
        order={order}
        onSort={handleSort}
        hasMore={data?.pagination.hasMore ?? false}
        canGoBack={pagination.canGoBack}
        onNext={handleNext}
        onPrev={pagination.goPrev}
      />
    </PageShell>
  );
}
