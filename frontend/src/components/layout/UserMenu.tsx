import { Avatar, Badge, Divider, Group, Menu, Stack, Text, UnstyledButton } from '@mantine/core';
import { useNavigate } from 'react-router-dom';
import { useMyEditGrants } from '../../features/edit-grants/hooks/useEditGrants';
import { ROUTES } from '../../routes/routePaths';
import { useUserStore } from '../../store/userStore';
import classes from './UserMenu.module.css';

const CHEVRON_DOWN = (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <path d="M6 9L12 15L18 9" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
);

const EDIT_ICON = (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <path
      d="M11 4H4C3.46957 4 2.96086 4.21071 2.58579 4.58579C2.21071 4.96086 2 5.46957 2 6V20C2 20.5304 2.21071 21.0391 2.58579 21.4142C2.96086 21.7893 3.46957 22 4 22H18C18.5304 22 19.0391 21.7893 19.4142 21.4142C19.7893 21.0391 20 20.5304 20 20V13"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path
      d="M18.5 2.50001C18.8978 2.10219 19.4374 1.87869 20 1.87869C20.5626 1.87869 21.1022 2.10219 21.5 2.50001C21.8978 2.89784 22.1213 3.4374 22.1213 4.00001C22.1213 4.56262 21.8978 5.10219 21.5 5.50001L12 15L8 16L9 12L18.5 2.50001Z"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
);

const SIGN_OUT_ICON = (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <path
      d="M9 21H5C4.46957 21 3.96086 20.7893 3.58579 20.4142C3.21071 20.0391 3 19.5304 3 19V5C3 4.46957 3.21071 3.96086 3.58579 3.58579C3.96086 3.21071 4.46957 3 5 3H9"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path d="M16 17L21 12L16 7" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
    <path d="M21 12H9" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
);

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
        <UnstyledButton data-testid="user-menu-trigger" aria-label="User menu" p="xs">
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
