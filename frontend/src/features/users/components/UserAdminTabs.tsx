import { Tabs } from '@mantine/core';
import { useNavigate } from 'react-router-dom';
import { ROUTES } from '../../../routes/routePaths';
import { useUserStore } from '../../../store/userStore';

export type UserAdminSection = 'users' | 'invitations';

interface UserAdminTab {
  value: UserAdminSection;
  label: string;
  route: string;
  permission: string;
}

const TABS: readonly UserAdminTab[] = [
  { value: 'users', label: 'Users', route: ROUTES.USERS, permission: 'users:read' },
  { value: 'invitations', label: 'Invitations', route: ROUTES.INVITATIONS, permission: 'invitations:manage' },
];

interface UserAdminTabsProps {
  active: UserAdminSection;
}

export function UserAdminTabs({ active }: UserAdminTabsProps) {
  const navigate = useNavigate();
  const hasPermission = useUserStore((state) => state.hasPermission);
  const tabs = TABS.filter((tab) => hasPermission(tab.permission));

  const handleChange = (value: string | null) => {
    const target = tabs.find((tab) => tab.value === value);
    if (target && target.value !== active) navigate(target.route);
  };

  return (
    <Tabs value={active} onChange={handleChange} mb="xl" data-testid="user-admin-tabs">
      <Tabs.List>
        {tabs.map((tab) => (
          <Tabs.Tab key={tab.value} value={tab.value} data-testid={`user-admin-tab-${tab.value}`}>
            {tab.label}
          </Tabs.Tab>
        ))}
      </Tabs.List>
    </Tabs>
  );
}
