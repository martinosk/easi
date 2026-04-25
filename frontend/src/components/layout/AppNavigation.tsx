import { useNavigate } from 'react-router-dom';
import logo from '../../assets/logo.svg';
import { ROUTES } from '../../routes/routePaths';
import { useUserStore } from '../../store/userStore';
import {
  BusinessDomainsIcon,
  CanvasIcon,
  EnterpriseArchIcon,
  InvitationsIcon,
  ReleaseNotesIcon,
  SettingsIcon,
  UsersIcon,
  ValueStreamsIcon,
} from './AppNavigation.icons';
import { UserMenu } from './UserMenu';

type AppView =
  | 'canvas'
  | 'business-domains'
  | 'value-streams'
  | 'invitations'
  | 'users'
  | 'settings'
  | 'enterprise-architecture'
  | 'my-edit-access';

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
};

interface NavEntry {
  view: AppView;
  label: string;
  testId: string;
  icon: React.ReactNode;
  permission?: string;
}

const NAV_ENTRIES: readonly NavEntry[] = [
  { view: 'canvas', label: 'Architecture Canvas', testId: 'nav-canvas', icon: CanvasIcon },
  { view: 'business-domains', label: 'Business Domains', testId: 'nav-business-domains', icon: BusinessDomainsIcon },
  { view: 'value-streams', label: 'Value Streams', testId: 'nav-value-streams', icon: ValueStreamsIcon, permission: 'valuestreams:read' },
  { view: 'enterprise-architecture', label: 'Enterprise Architecture', testId: 'nav-enterprise-architecture', icon: EnterpriseArchIcon, permission: 'enterprise-arch:read' },
  { view: 'users', label: 'Users', testId: 'nav-users', icon: UsersIcon, permission: 'users:read' },
  { view: 'invitations', label: 'Invitations', testId: 'nav-invitations', icon: InvitationsIcon, permission: 'invitations:manage' },
  { view: 'settings', label: 'Settings', testId: 'nav-settings', icon: SettingsIcon, permission: 'metamodel:write' },
];

function NavItems({ currentView, onNavigate }: { currentView: AppView; onNavigate: (view: AppView) => void }) {
  const hasPermission = useUserStore((state) => state.hasPermission);
  return (
    <nav className="app-header-nav">
      {NAV_ENTRIES.filter((e) => !e.permission || hasPermission(e.permission)).map((entry) => (
        <button
          key={entry.view}
          type="button"
          className={`app-header-nav-item ${currentView === entry.view ? 'app-header-nav-item-active' : ''}`}
          onClick={() => onNavigate(entry.view)}
          data-testid={entry.testId}
        >
          {entry.icon}
          {entry.label}
        </button>
      ))}
    </nav>
  );
}

export function AppNavigation({ currentView, onOpenReleaseNotes, chatButton }: AppNavigationProps) {
  const navigate = useNavigate();
  const handleNavigate = (view: AppView) => navigate(viewRouteMap[view]);

  return (
    <header className="app-header" data-testid="app-navigation">
      <div className="app-header-brand">
        <img src={logo} alt="easi logo" className="app-header-logo" />
        <span className="app-header-title">easi</span>
      </div>

      <NavItems currentView={currentView} onNavigate={handleNavigate} />

      <div className="app-header-actions">
        {chatButton}
        {onOpenReleaseNotes && (
          <button type="button" className="app-header-action-btn" onClick={onOpenReleaseNotes} title="View release notes">
            {ReleaseNotesIcon}
            What's New
          </button>
        )}
        <UserMenu />
      </div>
    </header>
  );
}
