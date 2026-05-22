import {
  Alert,
  Badge,
  Button,
  Center,
  Container,
  Group,
  Loader,
  NativeSelect,
  Paper,
  Stack,
  Table,
  Text,
  Title,
} from '@mantine/core';
import { useCallback, useEffect, useState } from 'react';
import toast from 'react-hot-toast';
import { ConfirmationDialog } from '../../../components/shared/ConfirmationDialog';
import { useUserStore } from '../../../store/userStore';
import type { UserRole } from '../../auth/types';
import { userApi } from '../api/userApi';
import { ChangeRoleModal } from '../components/ChangeRoleModal';
import type { User, UserStatus } from '../types';
import classes from './UsersPage.module.css';

const STATUS_OPTIONS = [
  { value: 'all', label: 'All' },
  { value: 'active', label: 'Active' },
  { value: 'disabled', label: 'Disabled' },
];

const ROLE_OPTIONS = [
  { value: 'all', label: 'All' },
  { value: 'admin', label: 'Admin' },
  { value: 'architect', label: 'Architect' },
  { value: 'stakeholder', label: 'Stakeholder' },
];

const STATUS_BADGE_COLORS: Record<UserStatus, string> = {
  active: 'green',
  disabled: 'gray',
};

function formatDate(dateString: string): string {
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
}

function formatDateTime(dateString: string | undefined): string {
  if (!dateString) return '-';
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

function useUserList(enabled: boolean, statusFilter: UserStatus | 'all', roleFilter: UserRole | 'all') {
  const [users, setUsers] = useState<User[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadUsers = useCallback(async () => {
    try {
      setIsLoading(true);
      setError(null);
      const statusParam = statusFilter !== 'all' ? statusFilter : undefined;
      const roleParam = roleFilter !== 'all' ? roleFilter : undefined;
      const allUsers = await userApi.getAll(statusParam, roleParam);
      setUsers(allUsers);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load users');
      toast.error('Failed to load users');
    } finally {
      setIsLoading(false);
    }
  }, [statusFilter, roleFilter]);

  useEffect(() => {
    if (enabled) {
      loadUsers();
    }
  }, [enabled, loadUsers]);

  return { users, isLoading, error, loadUsers };
}

function useUserActions(loadUsers: () => Promise<void>) {
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [isChangeRoleModalOpen, setIsChangeRoleModalOpen] = useState(false);
  const [disableTarget, setDisableTarget] = useState<User | null>(null);
  const [isDisabling, setIsDisabling] = useState(false);

  const handleChangeRole = async (newRole: UserRole) => {
    if (!selectedUser) return;
    try {
      await userApi.update(selectedUser.id, { role: newRole });
      toast.success(`Role changed to ${newRole} for ${selectedUser.email}`);
      setIsChangeRoleModalOpen(false);
      setSelectedUser(null);
      await loadUsers();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to change role');
    }
  };

  const confirmDisable = async () => {
    if (!disableTarget) return;
    setIsDisabling(true);
    try {
      await userApi.update(disableTarget.id, { status: 'disabled' });
      toast.success(`Account disabled for ${disableTarget.email}`);
      await loadUsers();
      setDisableTarget(null);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to disable user');
    } finally {
      setIsDisabling(false);
    }
  };

  const handleEnableUser = async (user: User) => {
    try {
      await userApi.update(user.id, { status: 'active' });
      toast.success(`Account enabled for ${user.email}`);
      await loadUsers();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to enable user');
    }
  };

  return {
    selectedUser,
    setSelectedUser,
    isChangeRoleModalOpen,
    setIsChangeRoleModalOpen,
    disableTarget,
    setDisableTarget,
    isDisabling,
    confirmDisable,
    handleChangeRole,
    handleEnableUser,
  };
}

export function UsersPage() {
  const hasPermission = useUserStore((state) => state.hasPermission);
  const currentUser = useUserStore((state) => state.user);
  const canReadUsers = hasPermission('users:read');
  const canManageUsers = hasPermission('users:manage');

  const [statusFilter, setStatusFilter] = useState<UserStatus | 'all'>('all');
  const [roleFilter, setRoleFilter] = useState<UserRole | 'all'>('all');
  const { users, isLoading, error, loadUsers } = useUserList(canReadUsers, statusFilter, roleFilter);
  const actions = useUserActions(loadUsers);

  if (!canReadUsers) {
    return (
      <PageShell>
        <Alert color="red">You do not have permission to view users.</Alert>
      </PageShell>
    );
  }

  const openChangeRoleModal = (user: User) => {
    actions.setSelectedUser(user);
    actions.setIsChangeRoleModalOpen(true);
  };

  return (
    <PageShell>
      <UsersHeader />
      <UsersFilters
        statusFilter={statusFilter}
        roleFilter={roleFilter}
        onStatusChange={setStatusFilter}
        onRoleChange={setRoleFilter}
      />
      <UsersContent
        isLoading={isLoading}
        error={error}
        users={users}
        currentUserId={currentUser?.id}
        canManageUsers={canManageUsers}
        onChangeRole={openChangeRoleModal}
        onDisable={actions.setDisableTarget}
        onEnable={actions.handleEnableUser}
      />
      {actions.selectedUser && (
        <ChangeRoleModal
          isOpen={actions.isChangeRoleModalOpen}
          onClose={() => {
            actions.setIsChangeRoleModalOpen(false);
            actions.setSelectedUser(null);
          }}
          onSubmit={actions.handleChangeRole}
          user={actions.selectedUser}
        />
      )}
      {actions.disableTarget && (
        <ConfirmationDialog
          title="Disable account?"
          message="Are you sure you want to disable the account for"
          itemName={actions.disableTarget.email}
          confirmText="Disable"
          onConfirm={actions.confirmDisable}
          onCancel={() => actions.setDisableTarget(null)}
          isLoading={actions.isDisabling}
        />
      )}
    </PageShell>
  );
}

function PageShell({ children }: { children: React.ReactNode }) {
  return (
    <div className={classes.page}>
      <Container size="xl" py="xl">
        {children}
      </Container>
    </div>
  );
}

function UsersHeader() {
  return (
    <Stack gap="xs" mb="xl">
      <Title order={1}>User Management</Title>
      <Text c="dimmed">View and manage users in your organization.</Text>
    </Stack>
  );
}

interface UsersFiltersProps {
  statusFilter: UserStatus | 'all';
  roleFilter: UserRole | 'all';
  onStatusChange: (status: UserStatus | 'all') => void;
  onRoleChange: (role: UserRole | 'all') => void;
}

function UsersFilters({ statusFilter, roleFilter, onStatusChange, onRoleChange }: UsersFiltersProps) {
  return (
    <Group gap="lg" mb="xl">
      <NativeSelect
        label="Status"
        data={STATUS_OPTIONS}
        value={statusFilter}
        onChange={(event) => onStatusChange(event.currentTarget.value as UserStatus | 'all')}
        data-testid="status-filter"
      />
      <NativeSelect
        label="Role"
        data={ROLE_OPTIONS}
        value={roleFilter}
        onChange={(event) => onRoleChange(event.currentTarget.value as UserRole | 'all')}
        data-testid="role-filter"
      />
    </Group>
  );
}

interface UsersContentProps {
  isLoading: boolean;
  error: string | null;
  users: User[];
  currentUserId: string | undefined;
  canManageUsers: boolean;
  onChangeRole: (user: User) => void;
  onDisable: (user: User) => void;
  onEnable: (user: User) => Promise<void>;
}

function UsersContent({
  isLoading,
  error,
  users,
  currentUserId,
  canManageUsers,
  onChangeRole,
  onDisable,
  onEnable,
}: UsersContentProps) {
  if (isLoading) {
    return (
      <Center py="xl">
        <Stack align="center" gap="md">
          <Loader />
          <Text>Loading users...</Text>
        </Stack>
      </Center>
    );
  }

  if (error) {
    return (
      <Alert color="red" data-testid="users-error">
        {error}
      </Alert>
    );
  }

  if (users.length === 0) {
    return (
      <Stack align="center" gap="md" py="xl">
        <Text size="lg" c="dimmed">
          No users found
        </Text>
      </Stack>
    );
  }

  return (
    <UsersTable
      users={users}
      currentUserId={currentUserId}
      canManageUsers={canManageUsers}
      onChangeRole={onChangeRole}
      onDisable={onDisable}
      onEnable={onEnable}
    />
  );
}

interface UsersTableProps {
  users: User[];
  currentUserId: string | undefined;
  canManageUsers: boolean;
  onChangeRole: (user: User) => void;
  onDisable: (user: User) => void;
  onEnable: (user: User) => Promise<void>;
}

function UsersTable({ users, currentUserId, canManageUsers, onChangeRole, onDisable, onEnable }: UsersTableProps) {
  return (
    <Paper shadow="sm" radius="lg" withBorder>
      <Table data-testid="users-table" striped highlightOnHover verticalSpacing="sm">
        <Table.Thead>
          <Table.Tr>
            <Table.Th>User</Table.Th>
            <Table.Th>Role</Table.Th>
            <Table.Th>Status</Table.Th>
            <Table.Th>Created</Table.Th>
            <Table.Th>Last Login</Table.Th>
            {canManageUsers && <Table.Th>Actions</Table.Th>}
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {users.map((user) => (
            <UserRow
              key={user.id}
              user={user}
              isCurrentUser={user.id === currentUserId}
              canManageUsers={canManageUsers}
              onChangeRole={onChangeRole}
              onDisable={onDisable}
              onEnable={onEnable}
            />
          ))}
        </Table.Tbody>
      </Table>
    </Paper>
  );
}

interface UserRowProps {
  user: User;
  isCurrentUser: boolean;
  canManageUsers: boolean;
  onChangeRole: (user: User) => void;
  onDisable: (user: User) => void;
  onEnable: (user: User) => Promise<void>;
}

function UserRow({ user, isCurrentUser, canManageUsers, onChangeRole, onDisable, onEnable }: UserRowProps) {
  return (
    <Table.Tr data-testid={`user-row-${user.id}`}>
      <Table.Td>
        <Stack gap={0}>
          <Text fw={500}>{user.email}</Text>
          {user.name && (
            <Text size="xs" c="dimmed">
              {user.name}
            </Text>
          )}
          {isCurrentUser && (
            <Badge size="xs" color="blue" variant="filled" mt="xs" w="fit-content">
              You
            </Badge>
          )}
        </Stack>
      </Table.Td>
      <Table.Td>
        <Badge variant="light" color="gray" tt="capitalize">
          {user.role}
        </Badge>
      </Table.Td>
      <Table.Td>
        <Badge variant="light" color={STATUS_BADGE_COLORS[user.status]} tt="capitalize">
          {user.status}
        </Badge>
      </Table.Td>
      <Table.Td>
        <Text c="dimmed" size="xs">
          {formatDate(user.createdAt)}
        </Text>
      </Table.Td>
      <Table.Td>
        <Text c="dimmed" size="xs">
          {formatDateTime(user.lastLoginAt)}
        </Text>
      </Table.Td>
      {canManageUsers && (
        <Table.Td>
          <UserRowActions
            user={user}
            isCurrentUser={isCurrentUser}
            onChangeRole={onChangeRole}
            onDisable={onDisable}
            onEnable={onEnable}
          />
        </Table.Td>
      )}
    </Table.Tr>
  );
}

interface UserRowActionsProps {
  user: User;
  isCurrentUser: boolean;
  onChangeRole: (user: User) => void;
  onDisable: (user: User) => void;
  onEnable: (user: User) => Promise<void>;
}

function UserRowActions({ user, isCurrentUser, onChangeRole, onDisable, onEnable }: UserRowActionsProps) {
  if (!user._links.update || isCurrentUser) return null;

  return (
    <Group gap="sm">
      <Button
        size="xs"
        variant="default"
        onClick={() => onChangeRole(user)}
        data-testid={`change-role-btn-${user.id}`}
      >
        Change Role
      </Button>
      {user.status === 'active' ? (
        <Button
          size="xs"
          color="red"
          onClick={() => onDisable(user)}
          data-testid={`disable-btn-${user.id}`}
        >
          Disable
        </Button>
      ) : (
        <Button
          size="xs"
          color="green"
          onClick={() => onEnable(user)}
          data-testid={`enable-btn-${user.id}`}
        >
          Enable
        </Button>
      )}
    </Group>
  );
}
