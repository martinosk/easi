import { Stack, Text, Title } from '@mantine/core';
import { Navigate, NavLink, Route, Routes } from 'react-router-dom';
import { useUserStore } from '../../../store/userStore';
import { OnePagersSettings } from '../../one-pagers';
import { AIConfigurationSettings } from '../components/AIConfigurationSettings';
import { AppearanceSettings } from '../components/AppearanceSettings';
import { MaturityScaleSettings } from '../components/MaturityScaleSettings';
import { SettingsErrorNotice } from '../components/SettingsSection';
import { StrategyPillarsSettings } from '../components/StrategyPillarsSettings';
import classes from './SettingsPage.module.css';

const SETTINGS_TABS = [
  { path: 'maturity-scale', label: 'Maturity Scale', element: <MaturityScaleSettings /> },
  { path: 'strategy-pillars', label: 'Strategy Pillars', element: <StrategyPillarsSettings /> },
  { path: 'ai-configuration', label: 'AI Configuration', element: <AIConfigurationSettings /> },
  { path: 'one-pagers', label: 'One-Pagers', element: <OnePagersSettings /> },
  { path: 'appearance', label: 'Appearance', element: <AppearanceSettings /> },
];

const tabClassName = ({ isActive }: { isActive: boolean }) =>
  isActive ? `${classes.tab} ${classes.tabActive}` : classes.tab;

export function SettingsPage() {
  const hasPermission = useUserStore((state) => state.hasPermission);
  const canManageMetaModel = hasPermission('metamodel:write');

  if (!canManageMetaModel) {
    return (
      <div className={classes.page}>
        <div className={classes.container}>
          <SettingsErrorNotice ta="center">You do not have permission to manage settings.</SettingsErrorNotice>
        </div>
      </div>
    );
  }

  return (
    <div className={classes.page}>
      <div className={classes.container}>
        <Stack gap="xs" mb="xl">
          <Title order={1}>Settings</Title>
          <Text c="dimmed" size="lg">
            Configure system-wide settings for your organization.
          </Text>
        </Stack>

        <nav className={classes.tabs}>
          {SETTINGS_TABS.map((tab) => (
            <NavLink key={tab.path} to={`/settings/${tab.path}`} className={tabClassName}>
              {tab.label}
            </NavLink>
          ))}
        </nav>

        <Routes>
          <Route path="/" element={<Navigate to="/settings/maturity-scale" replace />} />
          {SETTINGS_TABS.map((tab) => (
            <Route key={tab.path} path={`/${tab.path}`} element={tab.element} />
          ))}
        </Routes>
      </div>
    </div>
  );
}
