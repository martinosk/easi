import type { ReactElement } from 'react';
import { screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { Release } from '../../../api/types';
import { toReleaseVersion } from '../../../api/types';
import { renderWithProviders } from '../../../test/helpers';
import { ReleaseNotesOverlay } from './ReleaseNotesOverlay';

const render = (ui: ReactElement) => renderWithProviders(ui, { withRouter: false });

describe('ReleaseNotesOverlay', () => {
  const mockOnDismiss = vi.fn();

  const mockRelease: Release = {
    version: toReleaseVersion('1.2.0'),
    releaseDate: '2024-01-15T00:00:00Z',
    notes: '## Features\n- New dashboard\n- Improved performance\n\n## Bug Fixes\n- Fixed login issue',
    _links: { self: { href: '/api/v1/releases/1.2.0', method: 'GET' } },
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Markdown Parsing', () => {
    it('should render inline code formatting', () => {
      const releaseWithCode: Release = {
        ...mockRelease,
        notes: '## Features\n- Added `new-feature` support',
      };

      render(<ReleaseNotesOverlay isOpen={true} release={releaseWithCode} onDismiss={mockOnDismiss} />);

      const codeElement = screen.getByText('new-feature', {});
      expect(codeElement.tagName.toLowerCase()).toBe('code');
    });

    it('should render bold text formatting', () => {
      const releaseWithBold: Release = {
        ...mockRelease,
        notes: '## Features\n- **Important** new feature',
      };

      render(<ReleaseNotesOverlay isOpen={true} release={releaseWithBold} onDismiss={mockOnDismiss} />);

      const boldElement = screen.getByText('Important', {});
      expect(boldElement.tagName.toLowerCase()).toBe('strong');
    });

    it('should render italic text formatting', () => {
      const releaseWithItalic: Release = {
        ...mockRelease,
        notes: '## Features\n- *Slightly* improved',
      };

      render(<ReleaseNotesOverlay isOpen={true} release={releaseWithItalic} onDismiss={mockOnDismiss} />);

      const italicElement = screen.getByText('Slightly', {});
      expect(italicElement.tagName.toLowerCase()).toBe('em');
    });

    it('should handle notes without markdown sections as raw text', () => {
      const releaseWithPlainText: Release = {
        ...mockRelease,
        notes: 'This is a plain text release note without sections.',
      };

      render(<ReleaseNotesOverlay isOpen={true} release={releaseWithPlainText} onDismiss={mockOnDismiss} />);

      expect(screen.getByText(/This is a plain text release note/, {})).toBeInTheDocument();
    });
  });

  describe('Section Icons', () => {
    it.each([
      { heading: 'Major Features', item: 'New feature', icon: '★' },
      { heading: 'Bug Fixes', item: 'Fixed issue', icon: '✓' },
      { heading: 'Breaking Changes', item: 'API change', icon: '⚠' },
      { heading: 'API Changes', item: 'New endpoint', icon: '⚡' },
    ])('should show $icon icon for $heading section', ({ heading, item, icon }) => {
      const release: Release = { ...mockRelease, notes: `## ${heading}\n- ${item}` };

      render(<ReleaseNotesOverlay isOpen={true} release={release} onDismiss={mockOnDismiss} />);

      const sections = document.querySelectorAll('[data-testid="release-notes-section-icon"]');
      const section = Array.from(sections).find((s) => s.textContent?.includes(icon));
      expect(section).toBeDefined();
    });
  });
});
