import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ReleaseNotesOverlay } from './ReleaseNotesOverlay';
import type { Release } from '../../../api/types';

describe('ReleaseNotesOverlay', () => {
  const mockOnDismiss = vi.fn();

  const mockRelease: Release = {
    version: '1.2.0',
    releaseDate: '2024-01-15T00:00:00Z',
    notes: '## Features\n- New dashboard\n- Improved performance\n\n## Bug Fixes\n- Fixed login issue',
    _links: { self: { href: '/api/v1/releases/1.2.0' } },
  };

  beforeEach(() => {
    vi.clearAllMocks();
    HTMLDialogElement.prototype.showModal = vi.fn();
    HTMLDialogElement.prototype.close = vi.fn();
  });

  describe('Markdown Parsing', () => {
    it('should render inline code formatting', () => {
      const releaseWithCode: Release = {
        ...mockRelease,
        notes: '## Features\n- Added `new-feature` support',
      };

      render(
        <ReleaseNotesOverlay
          isOpen={true}
          release={releaseWithCode}
          onDismiss={mockOnDismiss}
        />
      );

      const codeElement = screen.getByText('new-feature', {});
      expect(codeElement.tagName.toLowerCase()).toBe('code');
    });

    it('should render bold text formatting', () => {
      const releaseWithBold: Release = {
        ...mockRelease,
        notes: '## Features\n- **Important** new feature',
      };

      render(
        <ReleaseNotesOverlay
          isOpen={true}
          release={releaseWithBold}
          onDismiss={mockOnDismiss}
        />
      );

      const boldElement = screen.getByText('Important', {});
      expect(boldElement.tagName.toLowerCase()).toBe('strong');
    });

    it('should render italic text formatting', () => {
      const releaseWithItalic: Release = {
        ...mockRelease,
        notes: '## Features\n- *Slightly* improved',
      };

      render(
        <ReleaseNotesOverlay
          isOpen={true}
          release={releaseWithItalic}
          onDismiss={mockOnDismiss}
        />
      );

      const italicElement = screen.getByText('Slightly', {});
      expect(italicElement.tagName.toLowerCase()).toBe('em');
    });

    it('should handle notes without markdown sections as raw text', () => {
      const releaseWithPlainText: Release = {
        ...mockRelease,
        notes: 'This is a plain text release note without sections.',
      };

      render(
        <ReleaseNotesOverlay
          isOpen={true}
          release={releaseWithPlainText}
          onDismiss={mockOnDismiss}
        />
      );

      expect(screen.getByText(/This is a plain text release note/, {})).toBeInTheDocument();
    });
  });

  describe('Section Icons', () => {
    it('should show star icon for features section', () => {
      const releaseWithFeatures: Release = {
        ...mockRelease,
        notes: '## Major Features\n- New feature',
      };

      render(
        <ReleaseNotesOverlay
          isOpen={true}
          release={releaseWithFeatures}
          onDismiss={mockOnDismiss}
        />
      );

      const iconElement = document.querySelector('.release-notes-section-icon');
      expect(iconElement?.textContent).toContain('★');
    });

    it('should show checkmark icon for bug fixes section', () => {
      const releaseWithBugFixes: Release = {
        ...mockRelease,
        notes: '## Bug Fixes\n- Fixed issue',
      };

      render(
        <ReleaseNotesOverlay
          isOpen={true}
          release={releaseWithBugFixes}
          onDismiss={mockOnDismiss}
        />
      );

      const sections = document.querySelectorAll('.release-notes-section-icon');
      const bugFixSection = Array.from(sections).find(s => s.textContent?.includes('✓'));
      expect(bugFixSection).toBeDefined();
    });

    it('should show warning icon for breaking changes section', () => {
      const releaseWithBreaking: Release = {
        ...mockRelease,
        notes: '## Breaking Changes\n- API change',
      };

      render(
        <ReleaseNotesOverlay
          isOpen={true}
          release={releaseWithBreaking}
          onDismiss={mockOnDismiss}
        />
      );

      const sections = document.querySelectorAll('.release-notes-section-icon');
      const breakingSection = Array.from(sections).find(s => s.textContent?.includes('⚠'));
      expect(breakingSection).toBeDefined();
    });

    it('should show lightning icon for API changes section', () => {
      const releaseWithAPI: Release = {
        ...mockRelease,
        notes: '## API Changes\n- New endpoint',
      };

      render(
        <ReleaseNotesOverlay
          isOpen={true}
          release={releaseWithAPI}
          onDismiss={mockOnDismiss}
        />
      );

      const sections = document.querySelectorAll('.release-notes-section-icon');
      const apiSection = Array.from(sections).find(s => s.textContent?.includes('⚡'));
      expect(apiSection).toBeDefined();
    });
  });
});
