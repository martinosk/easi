import { UnstyledButton } from '@mantine/core';
import { useNavigate } from 'react-router-dom';
import logo from '../../assets/logo.svg';
import type { AppView } from '../../routes/routePaths';
import { ROUTES } from '../../routes/routePaths';
import { useUserStore } from '../../store/userStore';
import {
  BusinessDomainsIcon,
  CanvasIcon,
  EnterpriseArchIcon,
  InvitationsIcon,
  OnePagerQualityIcon,
  ReleaseNotesIcon,
  SettingsIcon,
  UsersIcon,
  ValueStreamsIcon,
} from './AppNavigation.icons';
import classes from './AppNavigation.module.css';
import { UserMenu } from './UserMenu';

interface AppNavigationProps {
  currentView: AppView;
  onOpenReleaseNotes?: () => void;
  chatButton?: React.ReactNode;
}

const viewRouteMap: Record<AppView, string> = {
  canvas: ROUTES.HOME,
  'business-domains': ROUTES.BUSINESS_DOMAINS,
  'value-streams': ROUTES.VALUE_STREAMS,
  'enterprise-architecture': ROUTES.ENTERPRISE_ARCHITECTURE,
  invitations: ROUTES.INVITATIONS,
  users: ROUTES.USERS,
  settings: ROUTES.SETTINGS,
  'my-edit-access': ROUTES.MY_EDIT_ACCESS,
  'one-pagers': ROUTES.ONE_PAGERS,
  'one-pager-quality': ROUTES.ONE_PAGER_QUALITY,
};

interface NavEntry {
  view: AppView;
  label: string;
  testId: string;
  icon: React.ReactNode;
  permission?: string;
  sessionLink?: string;
}

const NAV_ENTRIES: readonly NavEntry[] = [
  { view: 'canvas', label: 'Architecture Canvas', testId: 'nav-canvas', icon: CanvasIcon },
  { view: 'business-domains', label: 'Business Domains', testId: 'nav-business-domains', icon: BusinessDomainsIcon },
  {
    view: 'value-streams',
    label: 'Value Streams',
    testId: 'nav-value-streams',
    icon: ValueStreamsIcon,
    permission: 'valuestreams:read',
  },
  {
    view: 'enterprise-architecture',
    label: 'Enterprise Architecture',
    testId: 'nav-enterprise-architecture',
    icon: EnterpriseArchIcon,
    permission: 'enterprise-arch:read',
  },
  { view: 'users', label: 'Users', testId: 'nav-users', icon: UsersIcon, permission: 'users:read' },
  {
    view: 'invitations',
    label: 'Invitations',
    testId: 'nav-invitations',
    icon: InvitationsIcon,
    permission: 'invitations:manage',
  },
  {
    view: 'one-pager-quality',
    label: 'One-Pager Quality',
    testId: 'nav-one-pager-quality',
    icon: OnePagerQualityIcon,
    sessionLink: 'x-one-pager-quality',
  },
  { view: 'settings', label: 'Settings', testId: 'nav-settings', icon: SettingsIcon, permission: 'metamodel:write' },
];

function NavItems({ currentView, onNavigate }: { currentView: AppView; onNavigate: (view: AppView) => void }) {
  const hasPermission = useUserStore((state) => state.hasPermission);
  const sessionLinks = useUserStore((state) => state.sessionLinks);
  const visibleEntries = NAV_ENTRIES.filter(
    (e) => (!e.permission || hasPermission(e.permission)) && (!e.sessionLink || Boolean(sessionLinks?.[e.sessionLink])),
  );
  return (
    <nav className={classes.nav}>
      {visibleEntries.map((entry) => (
        <UnstyledButton
          key={entry.view}
          component="button"
          type="button"
          className={`${classes.navItem} ${currentView === entry.view ? classes.navItemActive : ''}`}
          onClick={() => onNavigate(entry.view)}
          data-testid={entry.testId}
        >
          {entry.icon}
          {entry.label}
        </UnstyledButton>
      ))}
    </nav>
  );
}

export function AppNavigation({ currentView, onOpenReleaseNotes, chatButton }: AppNavigationProps) {
  const navigate = useNavigate();
  const tenant = useUserStore((state) => state.tenant);
  const handleNavigate = (view: AppView) => navigate(viewRouteMap[view]);

  return (
    <header className={classes.header} data-testid="app-navigation">
      <div className={classes.brand}>
        <img src={logo} alt="easi logo" className={classes.logo} />
        <span className={classes.wordmark}>easi</span>
      </div>

      <NavItems currentView={currentView} onNavigate={handleNavigate} />

      <div className={classes.actions}>
        {tenant && <span className={classes.tenant}>{tenant.name}</span>}
        {chatButton}
        {onOpenReleaseNotes && (
          <UnstyledButton
            component="button"
            type="button"
            className={classes.actionButton}
            onClick={onOpenReleaseNotes}
            title="View release notes"
          >
            {ReleaseNotesIcon}
            What's New
          </UnstyledButton>
        )}
        <UserMenu />
      </div>
    </header>
  );
}
