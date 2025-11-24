import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ReleaseNotesOverlay } from './ReleaseNotesOverlay';
import type { Release } from '../api/types';

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

  describe('Rendering', () => {
    it('should render the dialog with release version in title', () => {
      render(
        <ReleaseNotesOverlay
          isOpen={true}
          release={mockRelease}
          onDismiss={mockOnDismiss}
        />
      );

      expect(screen.getByRole('heading', { level: 2, hidden: true })).toHaveTextContent("What's New in 1.2.0");
    });

    it('should render release date', () => {
      render(
        <ReleaseNotesOverlay
          isOpen={true}
          release={mockRelease}
          onDismiss={mockOnDismiss}
        />
      );

      const dateElement = document.querySelector('.release-notes-date');
      expect(dateElement).not.toBeNull();
      expect(dateElement?.textContent).toBeTruthy();
    });

    it('should parse and render markdown sections', () => {
      render(
        <ReleaseNotesOverlay
          isOpen={true}
          release={mockRelease}
          onDismiss={mockOnDismiss}
        />
      );

      expect(screen.getByText('Features', {})).toBeInTheDocument();
      expect(screen.getByText('Bug Fixes', {})).toBeInTheDocument();
    });

    it('should render list items from markdown', () => {
      render(
        <ReleaseNotesOverlay
          isOpen={true}
          release={mockRelease}
          onDismiss={mockOnDismiss}
        />
      );

      expect(screen.getByText('New dashboard', {})).toBeInTheDocument();
      expect(screen.getByText('Improved performance', {})).toBeInTheDocument();
      expect(screen.getByText('Fixed login issue', {})).toBeInTheDocument();
    });

    it('should render both dismiss buttons', () => {
      render(
        <ReleaseNotesOverlay
          isOpen={true}
          release={mockRelease}
          onDismiss={mockOnDismiss}
        />
      );

      expect(screen.getByTestId('release-notes-hide-forever')).toBeInTheDocument();
      expect(screen.getByTestId('release-notes-dismiss')).toBeInTheDocument();
    });
  });

  describe('Dialog Behavior', () => {
    it('should call showModal when isOpen is true', () => {
      render(
        <ReleaseNotesOverlay
          isOpen={true}
          release={mockRelease}
          onDismiss={mockOnDismiss}
        />
      );

      expect(HTMLDialogElement.prototype.showModal).toHaveBeenCalled();
    });

    it('should call close when isOpen is false', () => {
      render(
        <ReleaseNotesOverlay
          isOpen={false}
          release={mockRelease}
          onDismiss={mockOnDismiss}
        />
      );

      expect(HTMLDialogElement.prototype.close).toHaveBeenCalled();
    });

    it('should not call showModal when isOpen is false', () => {
      render(
        <ReleaseNotesOverlay
          isOpen={false}
          release={mockRelease}
          onDismiss={mockOnDismiss}
        />
      );

      expect(HTMLDialogElement.prototype.showModal).not.toHaveBeenCalled();
    });
  });

  describe('Dismiss Actions', () => {
    it('should call onDismiss with forever when clicking hide forever button', () => {
      render(
        <ReleaseNotesOverlay
          isOpen={true}
          release={mockRelease}
          onDismiss={mockOnDismiss}
        />
      );

      fireEvent.click(screen.getByTestId('release-notes-hide-forever'));

      expect(mockOnDismiss).toHaveBeenCalledWith('forever');
    });

    it('should call onDismiss with untilNext when clicking got it button', () => {
      render(
        <ReleaseNotesOverlay
          isOpen={true}
          release={mockRelease}
          onDismiss={mockOnDismiss}
        />
      );

      fireEvent.click(screen.getByTestId('release-notes-dismiss'));

      expect(mockOnDismiss).toHaveBeenCalledWith('untilNext');
    });
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

  describe('Section Styling', () => {
    it('should apply major section styling for features', () => {
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

      const section = document.querySelector('.release-notes-section-major');
      expect(section).not.toBeNull();
    });

    it('should apply bugs section styling for bug fixes', () => {
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

      const section = document.querySelector('.release-notes-section-bugs');
      expect(section).not.toBeNull();
    });
  });

  describe('Date Formatting', () => {
    it('should handle invalid date gracefully by showing fallback', () => {
      const releaseWithInvalidDate: Release = {
        ...mockRelease,
        releaseDate: 'not-a-date',
      };

      render(
        <ReleaseNotesOverlay
          isOpen={true}
          release={releaseWithInvalidDate}
          onDismiss={mockOnDismiss}
        />
      );

      const dateElement = document.querySelector('.release-notes-date');
      expect(dateElement).not.toBeNull();
    });
  });
});
