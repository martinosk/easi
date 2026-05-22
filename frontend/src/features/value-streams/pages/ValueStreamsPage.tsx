import { useCallback, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Alert,
  Badge,
  Box,
  Button,
  Card,
  Center,
  Container,
  Group,
  Loader,
  Modal,
  Stack,
  Text,
  Textarea,
  TextInput,
  Title,
} from '@mantine/core';
import type { HATEOASLinks, ValueStream } from '../../../api/types';
import { useUserStore } from '../../../store/userStore';
import { hasLink } from '../../../utils/hateoas';
import { useValueStreams } from '../hooks/useValueStreams';
import classes from './ValueStreamsPage.module.css';

interface ValueStreamFormData {
  name: string;
  description: string;
}

const EMPTY_FORM: ValueStreamFormData = { name: '', description: '' };

interface ValueStreamFormModalProps {
  isOpen: boolean;
  isEditing: boolean;
  formData: ValueStreamFormData;
  onFormDataChange: (data: ValueStreamFormData) => void;
  onSubmit: () => void;
  onCancel: () => void;
}

function ValueStreamFormModal({
  isOpen,
  isEditing,
  formData,
  onFormDataChange,
  onSubmit,
  onCancel,
}: ValueStreamFormModalProps) {
  return (
    <Modal
      opened={isOpen}
      onClose={onCancel}
      title={isEditing ? 'Edit Value Stream' : 'Create Value Stream'}
      centered
      data-testid="value-stream-form"
    >
      <Stack gap="md">
        <TextInput
          id="vs-name"
          label="Name"
          value={formData.name}
          onChange={(e) => onFormDataChange({ ...formData, name: e.currentTarget.value })}
          placeholder="e.g. Customer Onboarding"
          maxLength={100}
          data-autofocus
        />
        <Textarea
          id="vs-description"
          label="Description"
          value={formData.description}
          onChange={(e) => onFormDataChange({ ...formData, description: e.currentTarget.value })}
          placeholder="Optional description..."
          maxLength={500}
          rows={3}
        />
        <Group justify="flex-end" gap="sm">
          <Button variant="default" onClick={onCancel}>
            Cancel
          </Button>
          <Button onClick={onSubmit} disabled={!formData.name.trim()}>
            {isEditing ? 'Save' : 'Create'}
          </Button>
        </Group>
      </Stack>
    </Modal>
  );
}

interface DeleteConfirmModalProps {
  streamName: string | null;
  onConfirm: () => void;
  onCancel: () => void;
}

function DeleteConfirmModal({ streamName, onConfirm, onCancel }: DeleteConfirmModalProps) {
  return (
    <Modal
      opened={streamName !== null}
      onClose={onCancel}
      title="Delete Value Stream"
      centered
      data-testid="delete-confirmation"
    >
      <Stack gap="md">
        <Text size="sm">
          Are you sure you want to delete &ldquo;{streamName}&rdquo;? This will also remove all its stages and mappings.
        </Text>
        <Group justify="flex-end" gap="sm">
          <Button variant="default" onClick={onCancel}>
            Cancel
          </Button>
          <Button color="red" onClick={onConfirm}>
            Delete
          </Button>
        </Group>
      </Stack>
    </Modal>
  );
}

interface ValueStreamCardProps {
  stream: ValueStream;
  canWrite: boolean;
  canDelete: boolean;
  onNavigate: (id: string) => void;
  onEdit: (stream: ValueStream) => void;
  onDelete: (stream: ValueStream) => void;
}

function ValueStreamCard({ stream, canWrite, canDelete, onNavigate, onEdit, onDelete }: ValueStreamCardProps) {
  const showEdit = canWrite && hasLink(stream, 'edit');
  const showDelete = canDelete && hasLink(stream, 'delete');

  return (
    <Card
      withBorder
      radius="md"
      padding="lg"
      data-testid={`value-stream-${stream.id}`}
      onClick={() => onNavigate(stream.id)}
      onKeyDown={(e) => e.key === 'Enter' && onNavigate(stream.id)}
      role="button"
      tabIndex={0}
      className={classes.card}
    >
      <Group justify="space-between" align="flex-start" wrap="nowrap">
        <Box style={{ minWidth: 0, flex: 1 }}>
          <Text size="lg" fw={600}>
            {stream.name}
          </Text>
          {stream.description && (
            <Text size="sm" c="dimmed" mt="xs">
              {stream.description}
            </Text>
          )}
          <Group gap="lg" mt="sm">
            <Text size="xs" c="dimmed">
              {stream.stageCount} stages
            </Text>
            <Text size="xs" c="dimmed">
              Created {new Date(stream.createdAt).toLocaleDateString()}
            </Text>
          </Group>
        </Box>
        <Group gap="xs" wrap="nowrap">
          {showEdit && (
            <Button
              size="xs"
              variant="default"
              onClick={(e) => {
                e.stopPropagation();
                onEdit(stream);
              }}
              data-testid={`edit-${stream.id}`}
            >
              Edit
            </Button>
          )}
          {showDelete && (
            <Button
              size="xs"
              variant="default"
              color="red"
              onClick={(e) => {
                e.stopPropagation();
                onDelete(stream);
              }}
              data-testid={`delete-${stream.id}`}
            >
              Delete
            </Button>
          )}
        </Group>
      </Group>
    </Card>
  );
}

interface SubmitFormArgs {
  editingStream: ValueStream | null;
  createFn: (name: string, description?: string) => Promise<unknown>;
  updateFn: (stream: ValueStream, name: string, description?: string) => Promise<unknown>;
  name: string;
  desc: string | undefined;
}

async function submitForm({ editingStream, createFn, updateFn, name, desc }: SubmitFormArgs) {
  if (editingStream) {
    await updateFn(editingStream, name, desc);
  } else {
    await createFn(name, desc);
  }
}

function useValueStreamFormState(
  createFn: (name: string, description?: string) => Promise<unknown>,
  updateFn: (stream: ValueStream, name: string, description?: string) => Promise<unknown>,
  deleteFn: (stream: ValueStream) => Promise<unknown>,
) {
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [editingStream, setEditingStream] = useState<ValueStream | null>(null);
  const [deletingStream, setDeletingStream] = useState<ValueStream | null>(null);
  const [formData, setFormData] = useState<ValueStreamFormData>(EMPTY_FORM);

  const handleSubmit = useCallback(async () => {
    if (!formData.name.trim()) return;
    const desc = formData.description || undefined;
    await submitForm({ editingStream, createFn, updateFn, name: formData.name, desc });
    setEditingStream(null);
    setShowCreateForm(false);
    setFormData(EMPTY_FORM);
  }, [createFn, updateFn, editingStream, formData]);

  const handleDelete = useCallback(async () => {
    if (!deletingStream) return;
    await deleteFn(deletingStream);
    setDeletingStream(null);
  }, [deleteFn, deletingStream]);

  const startEdit = useCallback((stream: ValueStream) => {
    setEditingStream(stream);
    setFormData({ name: stream.name, description: stream.description || '' });
  }, []);

  const closeForm = useCallback(() => {
    setShowCreateForm(false);
    setEditingStream(null);
    setFormData(EMPTY_FORM);
  }, []);

  const openCreateForm = useCallback(() => {
    setFormData(EMPTY_FORM);
    setShowCreateForm(true);
  }, []);

  return {
    isFormOpen: showCreateForm || editingStream !== null,
    isEditing: editingStream !== null,
    deletingStream,
    formData,
    setFormData,
    handleSubmit,
    handleDelete,
    startEdit,
    closeForm,
    openCreateForm,
    setDeletingStream,
  };
}

function checkCanCreate(canWrite: boolean, collectionLinks: HATEOASLinks | undefined): boolean {
  return canWrite && !!collectionLinks && hasLink({ _links: collectionLinks }, 'create');
}

const PLUS_ICON = (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <path d="M12 5V19M5 12H19" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
);

const STREAM_ICON = (
  <svg width="64" height="64" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <path
      d="M22 12H18L15 21L9 3L6 12H2"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
);

interface EmptyStateProps {
  canCreate: boolean;
  onCreate: () => void;
}

function EmptyState({ canCreate, onCreate }: EmptyStateProps) {
  return (
    <Card withBorder padding="xl" radius="md" data-testid="empty-state">
      <Stack align="center" gap="md">
        <Box c="gray.4">{STREAM_ICON}</Box>
        <Title order={3}>No value streams yet</Title>
        <Text size="sm" c="dimmed" ta="center" maw={400}>
          Value streams model how your organization delivers value end-to-end. Create your first value stream to get
          started.
        </Text>
        {canCreate && <Button onClick={onCreate}>Create your first value stream</Button>}
      </Stack>
    </Card>
  );
}

function PageShell({ children }: { children: React.ReactNode }) {
  return (
    <Box className={classes.page}>
      <Container size="xl" py="xl">
        {children}
      </Container>
    </Box>
  );
}

function LoadingPage() {
  return (
    <PageShell>
      <Center py="xl">
        <Group gap="sm">
          <Loader size="sm" />
          <Text c="dimmed">Loading value streams...</Text>
        </Group>
      </Center>
    </PageShell>
  );
}

function ErrorPage({ message }: { message: string }) {
  return (
    <PageShell>
      <Alert color="red" variant="light">
        Failed to load value streams: {message}
      </Alert>
    </PageShell>
  );
}

export function ValueStreamsPage() {
  const { valueStreams, isLoading, error, createValueStream, updateValueStream, deleteValueStream, collectionLinks } =
    useValueStreams();
  const hasPermission = useUserStore((state) => state.hasPermission);
  const navigate = useNavigate();
  const canWrite = hasPermission('valuestreams:write');
  const canDelete = hasPermission('valuestreams:delete');
  const canCreate = checkCanCreate(canWrite, collectionLinks);

  const form = useValueStreamFormState(createValueStream, updateValueStream, deleteValueStream);

  if (isLoading) return <LoadingPage />;
  if (error) return <ErrorPage message={error.message} />;

  return (
    <Box className={classes.page} data-testid="value-streams-page">
      <Container size="xl" py="xl">
        <Group justify="space-between" align="flex-start" mb="xl">
          <Box>
            <Group gap="sm" align="center">
              <Title order={1}>Value Streams</Title>
              <Badge size="lg" variant="light" color="gray">
                {valueStreams.length}
              </Badge>
            </Group>
            <Text c="dimmed" mt="xs">
              Model how your organization delivers value end-to-end.
            </Text>
          </Box>
          {canCreate && (
            <Button onClick={form.openCreateForm} leftSection={PLUS_ICON} data-testid="create-value-stream-btn">
              Create Value Stream
            </Button>
          )}
        </Group>

        <ValueStreamFormModal
          isOpen={form.isFormOpen}
          isEditing={form.isEditing}
          formData={form.formData}
          onFormDataChange={form.setFormData}
          onSubmit={form.handleSubmit}
          onCancel={form.closeForm}
        />

        <DeleteConfirmModal
          streamName={form.deletingStream?.name ?? null}
          onConfirm={form.handleDelete}
          onCancel={() => form.setDeletingStream(null)}
        />

        {valueStreams.length === 0 ? (
          <EmptyState canCreate={canCreate} onCreate={form.openCreateForm} />
        ) : (
          <Stack gap="md" data-testid="value-streams-list">
            {valueStreams.map((stream) => (
              <ValueStreamCard
                key={stream.id}
                stream={stream}
                canWrite={canWrite}
                canDelete={canDelete}
                onNavigate={navigate}
                onEdit={form.startEdit}
                onDelete={form.setDeletingStream}
              />
            ))}
          </Stack>
        )}
      </Container>
    </Box>
  );
}
