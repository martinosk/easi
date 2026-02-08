import { useState, useRef, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useUserStore } from '../../store/userStore';
import { useMyEditGrants } from '../../features/edit-grants/hooks/useEditGrants';
import { ROUTES } from '../../routes/routes';

function useClickOutside(ref: React.RefObject<HTMLDivElement | null>, isOpen: boolean, onClose: () => void) {
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (ref.current && !ref.current.contains(event.target as Node)) {
        onClose();
      }
    }

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isOpen, ref, onClose]);
}

function UserMenuDropdown({ user, tenant, activeGrantCount, onClose, onLogout }: {
  user: { name: string; email: string; role: string };
  tenant: { name: string };
  activeGrantCount: number;
  onClose: () => void;
  onLogout: () => void;
}) {
  return (
    <div className="user-menu-dropdown" data-testid="user-menu-dropdown">
      <div className="user-menu-header">
        <div className="user-menu-name">{user.name}</div>
        <div className="user-menu-email">{user.email}</div>
      </div>

      <div className="user-menu-divider" />

      <div className="user-menu-info">
        <div className="user-menu-info-row">
          <span className="user-menu-info-label">Organization</span>
          <span className="user-menu-info-value">{tenant.name}</span>
        </div>
        <div className="user-menu-info-row">
          <span className="user-menu-info-label">Role</span>
          <span className="user-menu-role-badge">{user.role}</span>
        </div>
      </div>

      {activeGrantCount > 0 && (
        <>
          <div className="user-menu-divider" />
          <Link
            to={ROUTES.MY_EDIT_ACCESS}
            className="user-menu-item"
            onClick={onClose}
            data-testid="user-menu-edit-access"
          >
            <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M11 4H4C3.46957 4 2.96086 4.21071 2.58579 4.58579C2.21071 4.96086 2 5.46957 2 6V20C2 20.5304 2.21071 21.0391 2.58579 21.4142C2.96086 21.7893 3.46957 22 4 22H18C18.5304 22 19.0391 21.7893 19.4142 21.4142C19.7893 21.0391 20 20.5304 20 20V13" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              <path d="M18.5 2.50001C18.8978 2.10219 19.4374 1.87869 20 1.87869C20.5626 1.87869 21.1022 2.10219 21.5 2.50001C21.8978 2.89784 22.1213 3.4374 22.1213 4.00001C22.1213 4.56262 21.8978 5.10219 21.5 5.50001L12 15L8 16L9 12L18.5 2.50001Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
            My Edit Access
            <span className="user-menu-badge">{activeGrantCount}</span>
          </Link>
        </>
      )}

      <div className="user-menu-divider" />

      <button
        type="button"
        className="user-menu-item user-menu-logout"
        onClick={onLogout}
        data-testid="user-menu-logout"
      >
        <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M9 21H5C4.46957 21 3.96086 20.7893 3.58579 20.4142C3.21071 20.0391 3 19.5304 3 19V5C3 4.46957 3.21071 3.96086 3.58579 3.58579C3.96086 3.21071 4.46957 3 5 3H9" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
          <path d="M16 17L21 12L16 7" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
          <path d="M21 12H9" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
        </svg>
        Sign out
      </button>
    </div>
  );
}

export function UserMenu() {
  const [isOpen, setIsOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  const user = useUserStore((state) => state.user);
  const tenant = useUserStore((state) => state.tenant);
  const logout = useUserStore((state) => state.logout);
  const { data: grants } = useMyEditGrants();
  const activeGrantCount = (grants?.filter(g => g.status === 'active') ?? []).length;

  const close = () => setIsOpen(false);
  useClickOutside(menuRef, isOpen, close);

  if (!user || !tenant) {
    return null;
  }

  const initials = user.name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);

  const handleLogout = async () => {
    close();
    await logout();
    const basePath = import.meta.env.BASE_URL || '/';
    window.location.href = `${basePath}login`;
  };

  return (
    <div className="user-menu" ref={menuRef}>
      <button
        type="button"
        className="user-menu-trigger"
        onClick={() => setIsOpen(!isOpen)}
        aria-expanded={isOpen}
        aria-haspopup="true"
        data-testid="user-menu-trigger"
      >
        <span className="user-menu-avatar">{initials}</span>
        <svg
          className={`user-menu-chevron ${isOpen ? 'user-menu-chevron-open' : ''}`}
          viewBox="0 0 24 24"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path d="M6 9L12 15L18 9" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
        </svg>
      </button>

      {isOpen && (
        <UserMenuDropdown
          user={user}
          tenant={tenant}
          activeGrantCount={activeGrantCount}
          onClose={close}
          onLogout={handleLogout}
        />
      )}
    </div>
  );
}
