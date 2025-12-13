import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ReleaseNotesBrowser } from './ReleaseNotesBrowser';
import { apiClient } from '../../../api/client';
import type { Release } from '../../../api/types';

vi.mock('../../../api/client', () => ({
  apiClient: {
    getReleases: vi.fn(),
    getVersion: vi.fn(),
  },
}));

const mockReleases: Release[] = [
  {
    version: '2.0.0',
    releaseDate: '2024-03-01T00:00:00Z',
    notes: '## Major Features\n- Complete redesign',
    _links: { self: { href: '/api/v1/releases/2.0.0' } },
  },
  {
    version: '1.1.0',
    releaseDate: '2024-02-01T00:00:00Z',
    notes: '## Features\n- New feature\n\n## Bug Fixes\n- Fixed issue',
    _links: { self: { href: '/api/v1/releases/1.1.0' } },
  },
  {
    version: '1.0.0',
    releaseDate: '2024-01-01T00:00:00Z',
    notes: '## Initial Release\n- First version',
    _links: { self: { href: '/api/v1/releases/1.0.0' } },
  },
];

describe('ReleaseNotesBrowser', () => {
  const mockOnClose = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    HTMLDialogElement.prototype.showModal = vi.fn();
    HTMLDialogElement.prototype.close = vi.fn();
  });

  describe('Loading State', () => {
    it('should show loading state when fetching releases', async () => {
      vi.mocked(apiClient.getReleases).mockImplementation(
        () => new Promise((resolve) => setTimeout(() => resolve(mockReleases), 100))
      );
      vi.mocked(apiClient.getVersion).mockResolvedValue('2.0.0');

      render(<ReleaseNotesBrowser isOpen={true} onClose={mockOnClose} />);

      expect(screen.getByText('Loading release notes...', {})).toBeInTheDocument();
    });

    it('should hide loading state after data is fetched', async () => {
      vi.mocked(apiClient.getReleases).mockResolvedValue(mockReleases);
      vi.mocked(apiClient.getVersion).mockResolvedValue('2.0.0');

      render(<ReleaseNotesBrowser isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.queryByText('Loading release notes...', {})).not.toBeInTheDocument();
      });
    });
  });

  describe('Error Handling', () => {
    it('should show error message when API fails', async () => {
      vi.mocked(apiClient.getReleases).mockRejectedValue(new Error('Network error'));
      vi.mocked(apiClient.getVersion).mockResolvedValue('2.0.0');

      render(<ReleaseNotesBrowser isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByText('Failed to load release notes', {})).toBeInTheDocument();
      });
    });
  });

  describe('Rendering Releases', () => {
    it('should render list of all releases', async () => {
      vi.mocked(apiClient.getReleases).mockResolvedValue(mockReleases);
      vi.mocked(apiClient.getVersion).mockResolvedValue('2.0.0');

      render(<ReleaseNotesBrowser isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByText('v2.0.0', {})).toBeInTheDocument();
        expect(screen.getByText('v1.1.0', {})).toBeInTheDocument();
        expect(screen.getByText('v1.0.0', {})).toBeInTheDocument();
      });
    });

    it('should sort releases by date descending', async () => {
      const unsortedReleases = [mockReleases[2], mockReleases[0], mockReleases[1]];
      vi.mocked(apiClient.getReleases).mockResolvedValue(unsortedReleases);
      vi.mocked(apiClient.getVersion).mockResolvedValue('2.0.0');

      render(<ReleaseNotesBrowser isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        const releaseItems = screen.getAllByRole('button', { hidden: true }).filter(
          btn => btn.classList.contains('release-browser-item')
        );
        expect(releaseItems[0]).toHaveTextContent('v2.0.0');
      });
    });

    it('should select the most recent release by default', async () => {
      vi.mocked(apiClient.getReleases).mockResolvedValue(mockReleases);
      vi.mocked(apiClient.getVersion).mockResolvedValue('2.0.0');

      render(<ReleaseNotesBrowser isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByText('Version 2.0.0', {})).toBeInTheDocument();
      });
    });

    it('should show current version badge on the current release', async () => {
      vi.mocked(apiClient.getReleases).mockResolvedValue(mockReleases);
      vi.mocked(apiClient.getVersion).mockResolvedValue('2.0.0');

      render(<ReleaseNotesBrowser isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByText('current', {})).toBeInTheDocument();
      });
    });

    it('should display current version in header', async () => {
      vi.mocked(apiClient.getReleases).mockResolvedValue(mockReleases);
      vi.mocked(apiClient.getVersion).mockResolvedValue('2.0.0');

      render(<ReleaseNotesBrowser isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByText('Current: v2.0.0', {})).toBeInTheDocument();
      });
    });
  });

  describe('Release Selection', () => {
    it('should switch displayed release when clicking a different version', async () => {
      vi.mocked(apiClient.getReleases).mockResolvedValue(mockReleases);
      vi.mocked(apiClient.getVersion).mockResolvedValue('2.0.0');

      render(<ReleaseNotesBrowser isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByText('Version 2.0.0', {})).toBeInTheDocument();
      });

      const releaseButton = screen.getByRole('button', { name: /v1.1.0/i, hidden: true });
      fireEvent.click(releaseButton);

      await waitFor(() => {
        expect(screen.getByText('Version 1.1.0', {})).toBeInTheDocument();
      });
    });

    it('should show release notes content for selected release', async () => {
      vi.mocked(apiClient.getReleases).mockResolvedValue(mockReleases);
      vi.mocked(apiClient.getVersion).mockResolvedValue('2.0.0');

      render(<ReleaseNotesBrowser isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByText('Major Features', {})).toBeInTheDocument();
        expect(screen.getByText('Complete redesign', {})).toBeInTheDocument();
      });

      const releaseButton = screen.getByRole('button', { name: /v1.1.0/i, hidden: true });
      fireEvent.click(releaseButton);

      await waitFor(() => {
        expect(screen.getByText('New feature', {})).toBeInTheDocument();
        expect(screen.getByText('Fixed issue', {})).toBeInTheDocument();
      });
    });
  });

  describe('Dialog Behavior', () => {
    it('should call showModal when isOpen changes to true', async () => {
      vi.mocked(apiClient.getReleases).mockResolvedValue(mockReleases);
      vi.mocked(apiClient.getVersion).mockResolvedValue('2.0.0');

      render(<ReleaseNotesBrowser isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(HTMLDialogElement.prototype.showModal).toHaveBeenCalled();
      });
    });

    it('should call close when isOpen changes to false', () => {
      render(<ReleaseNotesBrowser isOpen={false} onClose={mockOnClose} />);

      expect(HTMLDialogElement.prototype.close).toHaveBeenCalled();
    });

    it('should call onClose when close button is clicked', async () => {
      vi.mocked(apiClient.getReleases).mockResolvedValue(mockReleases);
      vi.mocked(apiClient.getVersion).mockResolvedValue('2.0.0');

      render(<ReleaseNotesBrowser isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.queryByText('Loading release notes...', {})).not.toBeInTheDocument();
      });

      const closeButtons = screen.getAllByRole('button', { name: /close/i, hidden: true });
      const primaryCloseButton = closeButtons.find(btn => btn.classList.contains('btn-primary'));
      expect(primaryCloseButton).toBeDefined();
      fireEvent.click(primaryCloseButton!);

      expect(mockOnClose).toHaveBeenCalled();
    });

    it('should call onClose when X button is clicked', async () => {
      vi.mocked(apiClient.getReleases).mockResolvedValue(mockReleases);
      vi.mocked(apiClient.getVersion).mockResolvedValue('2.0.0');

      render(<ReleaseNotesBrowser isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.queryByText('Loading release notes...', {})).not.toBeInTheDocument();
      });

      const closeXButton = screen.getByLabelText('Close', { selector: 'button' });
      fireEvent.click(closeXButton);

      expect(mockOnClose).toHaveBeenCalled();
    });
  });

  describe('Fetch Behavior', () => {
    it('should fetch data only when dialog opens', async () => {
      vi.mocked(apiClient.getReleases).mockResolvedValue(mockReleases);
      vi.mocked(apiClient.getVersion).mockResolvedValue('2.0.0');

      const { rerender } = render(<ReleaseNotesBrowser isOpen={false} onClose={mockOnClose} />);

      expect(apiClient.getReleases).not.toHaveBeenCalled();
      expect(apiClient.getVersion).not.toHaveBeenCalled();

      rerender(<ReleaseNotesBrowser isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(apiClient.getReleases).toHaveBeenCalledTimes(1);
        expect(apiClient.getVersion).toHaveBeenCalledTimes(1);
      });
    });

    it('should refetch data when dialog reopens', async () => {
      vi.mocked(apiClient.getReleases).mockResolvedValue(mockReleases);
      vi.mocked(apiClient.getVersion).mockResolvedValue('2.0.0');

      const { rerender } = render(<ReleaseNotesBrowser isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(apiClient.getReleases).toHaveBeenCalledTimes(1);
      });

      rerender(<ReleaseNotesBrowser isOpen={false} onClose={mockOnClose} />);
      rerender(<ReleaseNotesBrowser isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(apiClient.getReleases).toHaveBeenCalledTimes(2);
      });
    });
  });

  describe('Empty State', () => {
    it('should handle empty releases list', async () => {
      vi.mocked(apiClient.getReleases).mockResolvedValue([]);
      vi.mocked(apiClient.getVersion).mockResolvedValue('1.0.0');

      render(<ReleaseNotesBrowser isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByText('No release selected', {})).toBeInTheDocument();
      });
    });
  });

  describe('Markdown Rendering', () => {
    it('should parse sections from markdown notes', async () => {
      vi.mocked(apiClient.getReleases).mockResolvedValue(mockReleases);
      vi.mocked(apiClient.getVersion).mockResolvedValue('2.0.0');

      render(<ReleaseNotesBrowser isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByText('Major Features', {})).toBeInTheDocument();
      });
    });

    it('should render inline code in release notes', async () => {
      const releasesWithCode: Release[] = [{
        version: '1.0.0',
        releaseDate: '2024-01-01T00:00:00Z',
        notes: '## Features\n- Added `new-command` support',
        _links: { self: { href: '/api/v1/releases/1.0.0' } },
      }];

      vi.mocked(apiClient.getReleases).mockResolvedValue(releasesWithCode);
      vi.mocked(apiClient.getVersion).mockResolvedValue('1.0.0');

      render(<ReleaseNotesBrowser isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        const codeElement = screen.getByText('new-command', {});
        expect(codeElement.tagName.toLowerCase()).toBe('code');
      });
    });

    it('should render bold text in release notes', async () => {
      const releasesWithBold: Release[] = [{
        version: '1.0.0',
        releaseDate: '2024-01-01T00:00:00Z',
        notes: '## Features\n- **Important** update',
        _links: { self: { href: '/api/v1/releases/1.0.0' } },
      }];

      vi.mocked(apiClient.getReleases).mockResolvedValue(releasesWithBold);
      vi.mocked(apiClient.getVersion).mockResolvedValue('1.0.0');

      render(<ReleaseNotesBrowser isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        const boldElement = screen.getByText('Important', {});
        expect(boldElement.tagName.toLowerCase()).toBe('strong');
      });
    });

    it('should handle notes without sections as raw text', async () => {
      const releasesWithPlainText: Release[] = [{
        version: '1.0.0',
        releaseDate: '2024-01-01T00:00:00Z',
        notes: 'Plain text release notes without sections.',
        _links: { self: { href: '/api/v1/releases/1.0.0' } },
      }];

      vi.mocked(apiClient.getReleases).mockResolvedValue(releasesWithPlainText);
      vi.mocked(apiClient.getVersion).mockResolvedValue('1.0.0');

      render(<ReleaseNotesBrowser isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        expect(screen.getByText(/Plain text release notes/, {})).toBeInTheDocument();
      });
    });
  });

  describe('Date Formatting', () => {
    it('should display release date in main content area', async () => {
      vi.mocked(apiClient.getReleases).mockResolvedValue(mockReleases);
      vi.mocked(apiClient.getVersion).mockResolvedValue('2.0.0');

      render(<ReleaseNotesBrowser isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        const dateElement = document.querySelector('.release-browser-main-date');
        expect(dateElement).not.toBeNull();
        expect(dateElement?.textContent).toBeTruthy();
      });
    });

    it('should display release dates in sidebar', async () => {
      vi.mocked(apiClient.getReleases).mockResolvedValue(mockReleases);
      vi.mocked(apiClient.getVersion).mockResolvedValue('2.0.0');

      render(<ReleaseNotesBrowser isOpen={true} onClose={mockOnClose} />);

      await waitFor(() => {
        const sidebarDates = document.querySelectorAll('.release-browser-item-date');
        expect(sidebarDates.length).toBeGreaterThan(0);
        sidebarDates.forEach(dateEl => {
          expect(dateEl.textContent).toBeTruthy();
        });
      });
    });
  });
});
