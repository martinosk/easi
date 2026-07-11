import { Avatar, Badge, Divider, Group, Menu, Stack, Text, UnstyledButton } from '@mantine/core';
import { IconChevronDown, IconEdit, IconLogout } from '@tabler/icons-react';
import { useNavigate } from 'react-router-dom';
import { useMyEditGrants } from '../../features/edit-grants/hooks/useEditGrants';
import { ROUTES } from '../../routes/routePaths';
import { useUserStore } from '../../store/userStore';
import classes from './UserMenu.module.css';

const CHEVRON_DOWN = <IconChevronDown size={14} stroke={1.75} aria-hidden="true" />;

const EDIT_ICON = <IconEdit size={16} stroke={1.75} aria-hidden="true" />;

const SIGN_OUT_ICON = <IconLogout size={16} stroke={1.75} aria-hidden="true" />;

function getInitials(name: string): string {
  return name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);
}

interface UserMenuHeaderProps {
  name: string;
  email: string;
}

function UserMenuHeader({ name, email }: UserMenuHeaderProps) {
  return (
    <Stack gap="xs" px="md" py="sm">
      <Text fw={600} size="sm">
        {name}
      </Text>
      <Text size="xs" c="dimmed">
        {email}
      </Text>
    </Stack>
  );
}

interface UserMenuInfoProps {
  organizationName: string;
  role: string;
}

function UserMenuInfo({ organizationName, role }: UserMenuInfoProps) {
  return (
    <Stack gap="xs" px="md" py="sm">
      <Group justify="space-between">
        <Text size="xs" c="dimmed">
          Organization
        </Text>
        <Text size="xs" fw={500}>
          {organizationName}
        </Text>
      </Group>
      <Group justify="space-between">
        <Text size="xs" c="dimmed">
          Role
        </Text>
        <Badge size="sm" variant="light" color="blue">
          {role}
        </Badge>
      </Group>
    </Stack>
  );
}

export function UserMenu() {
  const navigate = useNavigate();
  const user = useUserStore((state) => state.user);
  const tenant = useUserStore((state) => state.tenant);
  const logout = useUserStore((state) => state.logout);
  const { data: grants } = useMyEditGrants();
  const activeGrantCount = (grants?.filter((g) => g.status === 'active') ?? []).length;

  if (!user || !tenant) {
    return null;
  }

  const handleLogout = async () => {
    await logout();
    const basePath = import.meta.env.BASE_URL || '/';
    window.location.href = `${basePath}login`;
  };

  return (
    <Menu shadow="md" classNames={{ dropdown: classes.dropdown }} position="bottom-end" withinPortal>
      <Menu.Target>
        <UnstyledButton data-testid="user-menu-trigger" aria-label="User menu" p="xs" className={classes.trigger}>
          <Group gap="xs" wrap="nowrap">
            <Avatar size="sm" color="blue" radius="xl">
              {getInitials(user.name)}
            </Avatar>
            {CHEVRON_DOWN}
          </Group>
        </UnstyledButton>
      </Menu.Target>

      <Menu.Dropdown data-testid="user-menu-dropdown">
        <UserMenuHeader name={user.name} email={user.email} />
        <Divider />
        <UserMenuInfo organizationName={tenant.name} role={user.role} />

        {activeGrantCount > 0 && (
          <>
            <Divider />
            <Menu.Item
              leftSection={EDIT_ICON}
              rightSection={
                <Badge size="xs" variant="light" color="blue">
                  {activeGrantCount}
                </Badge>
              }
              onClick={() => navigate(ROUTES.MY_EDIT_ACCESS)}
              data-testid="user-menu-edit-access"
            >
              My Edit Access
            </Menu.Item>
          </>
        )}

        <Divider />
        <Menu.Item leftSection={SIGN_OUT_ICON} onClick={handleLogout} data-testid="user-menu-logout" c="red">
          Sign out
        </Menu.Item>
      </Menu.Dropdown>
    </Menu>
  );
}
