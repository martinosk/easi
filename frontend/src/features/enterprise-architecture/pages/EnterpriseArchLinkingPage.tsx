import { LinkingPage } from '../components/LinkingPage';
import { useUserStore } from '../../../store/userStore';

interface EnterpriseArchLinkingPageProps {
  onNavigateBack?: () => void;
}

export function EnterpriseArchLinkingPage({ onNavigateBack }: EnterpriseArchLinkingPageProps) {
  const hasPermission = useUserStore((state) => state.hasPermission);
  const canRead = hasPermission('enterprise-arch:read');

  if (!canRead) {
    return (
      <div style={{ padding: '2rem' }}>
        <div style={{ color: '#dc2626', fontSize: '1rem' }}>
          You do not have permission to view enterprise architecture.
        </div>
      </div>
    );
  }

  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      {onNavigateBack && (
        <div style={{ padding: '1rem', borderBottom: '1px solid #e5e7eb', backgroundColor: '#f9fafb' }}>
          <button
            onClick={onNavigateBack}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '0.5rem',
              padding: '0.5rem 1rem',
              border: '1px solid #d1d5db',
              borderRadius: '0.375rem',
              backgroundColor: '#ffffff',
              cursor: 'pointer',
              fontSize: '0.875rem',
              color: '#374151',
            }}
          >
            ← Back to Enterprise Capabilities
          </button>
        </div>
      )}
      <div style={{ flex: 1, overflow: 'hidden' }}>
        <LinkingPage />
      </div>
    </div>
  );
}
