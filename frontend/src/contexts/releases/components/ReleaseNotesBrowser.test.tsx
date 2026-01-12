import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { ReleaseNotesBrowser } from './ReleaseNotesBrowser';
import { apiClient } from '../../../api/client';
import type { Release } from '../../../api/types';
import { toReleaseVersion } from '../../../api/types';

vi.mock('../../../api/client', () => ({
  apiClient: {
    getReleases: vi.fn(),
    getVersion: vi.fn(),
  },
}));

const mockReleases: Release[] = [
  {
    version: toReleaseVersion('2.0.0'),
    releaseDate: '2024-03-01T00:00:00Z',
    notes: '## Major Features\n- Complete redesign',
    _links: { self: { href: '/api/v1/releases/2.0.0', method: 'GET' } },
  },
  {
    version: toReleaseVersion('1.1.0'),
    releaseDate: '2024-02-01T00:00:00Z',
    notes: '## Features\n- New feature\n\n## Bug Fixes\n- Fixed issue',
    _links: { self: { href: '/api/v1/releases/1.1.0', method: 'GET' } },
  },
  {
    version: toReleaseVersion('1.0.0'),
    releaseDate: '2024-01-01T00:00:00Z',
    notes: '## Initial Release\n- First version',
    _links: { self: { href: '/api/v1/releases/1.0.0', method: 'GET' } },
  },
];

describe('ReleaseNotesBrowser', () => {
  const mockOnClose = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    HTMLDialogElement.prototype.showModal = vi.fn();
    HTMLDialogElement.prototype.close = vi.fn();
  });

  describe('Release Sorting and Selection', () => {
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
});
