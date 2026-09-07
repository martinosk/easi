import { Menu, Tooltip, UnstyledButton } from '@mantine/core';
import { useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import logo from '../../assets/logo.svg';
import type { AppView } from '../../routes/routePaths';
import { ROUTES } from '../../routes/routePaths';
import { useUserStore } from '../../store/userStore';
import {
  BusinessDomainsIcon,
  CanvasIcon,
  MoreIcon,
  OnePagerQualityIcon,
  ReleaseNotesIcon,
  StrategicFitIcon,
  UsersIcon,
  ValueStreamsIcon,
} from './AppNavigation.icons';
import classes from './AppNavigation.module.css';
import { HeaderActionButton } from './HeaderActionButton';
import { UserMenu } from './UserMenu';
import { useNavLayout } from './useNavLayout';

interface AppNavigationProps {
  currentView: AppView;
  onOpenReleaseNotes?: () => void;
  chatButton?: React.ReactNode;
}

const navViewForView: Partial<Record<AppView, AppView>> = {
  invitations: 'users',
};

interface NavEntry {
  view: AppView;
  route: string;
  label: string;
  testId: string;
  icon: React.ReactNode;
  permission?: string;
  sessionLink?: string;
}

const NAV_ENTRIES: readonly NavEntry[] = [
  { view: 'canvas', route: ROUTES.HOME, label: 'Architecture Canvas', testId: 'nav-canvas', icon: CanvasIcon },
  {
    view: 'business-domains',
    route: ROUTES.BUSINESS_DOMAINS,
    label: 'Business Domains',
    testId: 'nav-business-domains',
    icon: BusinessDomainsIcon,
  },
  {
    view: 'value-streams',
    route: ROUTES.VALUE_STREAMS,
    label: 'Value Streams',
    testId: 'nav-value-streams',
    icon: ValueStreamsIcon,
    permission: 'valuestreams:read',
  },
  {
    view: 'strategic-fit',
    route: ROUTES.STRATEGIC_FIT,
    label: 'Strategic Fit',
    testId: 'nav-strategic-fit',
    icon: StrategicFitIcon,
    permission: 'capabilities:read',
  },
  {
    view: 'users',
    route: ROUTES.USERS,
    label: 'Users',
    testId: 'nav-users',
    icon: UsersIcon,
    permission: 'users:read',
  },
  {
    view: 'one-pager-quality',
    route: ROUTES.ONE_PAGER_QUALITY,
    label: 'One-Pager Quality',
    testId: 'nav-one-pager-quality',
    icon: OnePagerQualityIcon,
    sessionLink: 'x-one-pager-quality',
  },
];

const MORE_LABEL = 'More';

function navItemClass(active: boolean): string {
  return active ? `${classes.navItem} ${classes.navItemActive}` : classes.navItem;
}

function useVisibleEntries(): NavEntry[] {
  const hasPermission = useUserStore((state) => state.hasPermission);
  const sessionLinks = useUserStore((state) => state.sessionLinks);
  return NAV_ENTRIES.filter(
    (e) => (!e.permission || hasPermission(e.permission)) && (!e.sessionLink || Boolean(sessionLinks?.[e.sessionLink])),
  );
}

interface NavButtonProps {
  entry: NavEntry;
  active: boolean;
  compact: boolean;
  onNavigate: (entry: NavEntry) => void;
}

function NavButton({ entry, active, compact, onNavigate }: NavButtonProps) {
  return (
    <Tooltip label={entry.label} openDelay={300} withinPortal>
      <UnstyledButton
        component="button"
        type="button"
        className={navItemClass(active)}
        onClick={() => onNavigate(entry)}
        aria-current={active ? 'page' : undefined}
        aria-label={compact ? entry.label : undefined}
        data-testid={entry.testId}
      >
        {entry.icon}
        {!compact && entry.label}
      </UnstyledButton>
    </Tooltip>
  );
}

interface NavOverflowMenuProps {
  entries: NavEntry[];
  activeView: AppView;
  onNavigate: (entry: NavEntry) => void;
}

function NavOverflowMenu({ entries, activeView, onNavigate }: NavOverflowMenuProps) {
  const active = entries.some((e) => e.view === activeView);
  return (
    <Menu shadow="md" position="bottom-start" withinPortal>
      <Tooltip label={MORE_LABEL} openDelay={300} withinPortal>
        <Menu.Target>
          <UnstyledButton
            component="button"
            type="button"
            className={navItemClass(active)}
            aria-label={MORE_LABEL}
            aria-current={active ? 'page' : undefined}
            data-testid="nav-more"
          >
            {MoreIcon}
          </UnstyledButton>
        </Menu.Target>
      </Tooltip>
      <Menu.Dropdown data-testid="nav-more-menu">
        {entries.map((entry) => (
          <Menu.Item
            key={entry.view}
            leftSection={entry.icon}
            onClick={() => onNavigate(entry)}
            data-testid={`${entry.testId}-overflow`}
            data-active={entry.view === activeView || undefined}
          >
            {entry.label}
          </Menu.Item>
        ))}
      </Menu.Dropdown>
    </Menu>
  );
}

function NavMeasureTwin({ entries }: { entries: NavEntry[] }) {
  return (
    <div className={classes.measure} data-measure-twin aria-hidden="true">
      {entries.map((entry) => (
        <span key={entry.view} className={classes.navItem} data-measure="full">
          {entry.icon}
          {entry.label}
        </span>
      ))}
      <span className={classes.navItem} data-measure="compact">
        {UsersIcon}
      </span>
      <span className={classes.navItem} data-measure="more">
        {MoreIcon}
      </span>
    </div>
  );
}

function PrimaryNav({ currentView, onNavigate }: { currentView: AppView; onNavigate: (entry: NavEntry) => void }) {
  const entries = useVisibleEntries();
  const navRef = useRef<HTMLElement>(null);
  const { mode, visibleCount } = useNavLayout(navRef, entries.length);
  const activeView = navViewForView[currentView] ?? currentView;
  const visible = entries.slice(0, visibleCount);
  const overflow = entries.slice(visibleCount);
  const compact = mode !== 'full';

  return (
    <nav ref={navRef} className={classes.nav} data-testid="app-primary-nav" aria-label="Primary">
      <NavMeasureTwin entries={entries} />
      <div className={classes.navRow} data-nav-row>
        {visible.map((entry) => (
          <NavButton
            key={entry.view}
            entry={entry}
            active={entry.view === activeView}
            compact={compact}
            onNavigate={onNavigate}
          />
        ))}
        {overflow.length > 0 && <NavOverflowMenu entries={overflow} activeView={activeView} onNavigate={onNavigate} />}
      </div>
    </nav>
  );
}

export function AppNavigation({ currentView, onOpenReleaseNotes, chatButton }: AppNavigationProps) {
  const navigate = useNavigate();
  const tenant = useUserStore((state) => state.tenant);
  const handleNavigate = (entry: NavEntry) => navigate(entry.route);

  return (
    <header className={classes.header} data-testid="app-navigation">
      <div className={classes.brand}>
        <img src={logo} alt="easi logo" className={classes.logo} />
        <span className={classes.wordmark}>easi</span>
      </div>

      <PrimaryNav currentView={currentView} onNavigate={handleNavigate} />

      <div className={classes.actions}>
        {tenant && <span className={classes.tenant}>{tenant.name}</span>}
        {chatButton}
        {onOpenReleaseNotes && (
          <HeaderActionButton
            icon={ReleaseNotesIcon}
            label="What's New"
            onClick={onOpenReleaseNotes}
            testId="nav-whats-new"
          />
        )}
        <UserMenu />
      </div>
    </header>
  );
}
