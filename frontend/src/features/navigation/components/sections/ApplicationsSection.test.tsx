import type { ReactElement } from 'react';
import { screen, waitFor } from '@testing-library/react';
import { renderWithProviders } from '../../../../test/helpers';
import { seedOnePagerCompleteness } from '../../../../test/mocks/onePagerCompleteness';
import { seedComponentStatistics } from '../../../../test/mocks/componentStatistics';
const render = (ui: ReactElement) => renderWithProviders(ui, { withRouter: false });
import { describe, expect, it, vi } from 'vitest';
import type { Component, ComponentId, HATEOASLinks } from '../../../../api/types';
import { ApplicationsSection } from './ApplicationsSection';

describe('ApplicationsSection', () => {
  const mockLinks: HATEOASLinks = { self: { href: '/test', method: 'GET' } };

  const createMockComponent = (overrides: Partial<Component> = {}): Component => ({
    id: 'comp-123' as ComponentId,
    name: 'Billing System',
    ownershipState: 'unknown',
    hosting: 'unknown',
    createdAt: '2021-01-01T00:00:00Z',
    _links: mockLinks,
    ...overrides,
  });

  const mockMultiSelect = {
    isMultiSelected: () => false,
    handleItemClick: () => 'single' as const,
    handleContextMenu: () => false,
    handleDragStart: () => false,
    selectedItems: [],
  };

  const defaultProps = {
    components: [],
    currentView: null,
    selectedNodeId: null,
    isExpanded: true,
    onToggle: vi.fn(),
    onAddComponent: vi.fn(),
    onComponentSelect: vi.fn(),
    onComponentContextMenu: vi.fn(),
    editingState: null,
    setEditingState: vi.fn(),
    onRenameSubmit: vi.fn(),
    editInputRef: { current: null },
    multiSelect: mockMultiSelect,
  };

  describe('one-pager completeness indicator', () => {
    it('shows the indicator only for applications the completeness collection reports as incomplete', async () => {
      seedOnePagerCompleteness('application', [
        { subjectId: 'comp-123', complete: false },
        { subjectId: 'comp-456', complete: true },
      ]);
      render(
        <ApplicationsSection
          {...defaultProps}
          components={[
            createMockComponent({ id: 'comp-123' as ComponentId }),
            createMockComponent({ id: 'comp-456' as ComponentId, name: 'Other System' }),
          ]}
        />,
      );

      expect(await screen.findByTestId('one-pager-incomplete-comp-123')).toBeInTheDocument();
      expect(screen.queryByTestId('one-pager-incomplete-comp-456')).not.toBeInTheDocument();
    });

    it('shows no indicator when the subject type has no required field', async () => {
      seedOnePagerCompleteness('application', []);
      render(
        <ApplicationsSection {...defaultProps} components={[createMockComponent({ id: 'comp-123' as ComponentId })]} />,
      );

      await waitFor(() => expect(screen.queryByTestId('one-pager-incomplete-comp-123')).not.toBeInTheDocument());
    });
  });

  describe('ownership', () => {
    it('shows counts per ownership state including the orphan count', async () => {
      seedComponentStatistics({ unknown: 2, nominated: 1, owned: 3, managed: 0, total: 6 });
      render(<ApplicationsSection {...defaultProps} components={[createMockComponent()]} />);

      const summary = await screen.findByTestId('ownership-distribution');
      expect(summary).toHaveTextContent('Unknown2');
      expect(summary).toHaveTextContent('Nominated1');
      expect(summary).toHaveTextContent('Owned3');
      expect(summary).toHaveTextContent('Managed0');
    });

    it('filters applications by ownership state', async () => {
      const owned = createMockComponent({ id: 'comp-owned' as ComponentId, name: 'Owned App', ownershipState: 'owned' });
      const orphan = createMockComponent({
        id: 'comp-orphan' as ComponentId,
        name: 'Orphan App',
        ownershipState: 'unknown',
        hosting: 'unknown',
      });
      render(<ApplicationsSection {...defaultProps} components={[owned, orphan]} />);
      await screen.findByTestId('ownership-distribution');

      expect(screen.getByText('Owned App')).toBeInTheDocument();
      expect(screen.getByText('Orphan App')).toBeInTheDocument();

      const user = (await import('@testing-library/user-event')).default;
      await user.click(screen.getByTestId('ownership-filter'));
      await user.click(await screen.findByRole('option', { name: 'Owned', hidden: true }));

      expect(screen.getByText('Owned App')).toBeInTheDocument();
      expect(screen.queryByText('Orphan App')).not.toBeInTheDocument();
    });
  });

  describe('hosting', () => {
    it('shows counts per hosting classification', async () => {
      seedComponentStatistics({
        hosting: { 'on-premises': 1, cloud: 2, saas: 3, 'third-party-hosted': 0, unknown: 4 },
        total: 10,
      });
      render(<ApplicationsSection {...defaultProps} components={[createMockComponent()]} />);

      const distribution = await screen.findByTestId('hosting-distribution');
      expect(distribution).toHaveTextContent('On-premises1');
      expect(distribution).toHaveTextContent('Cloud2');
      expect(distribution).toHaveTextContent('SaaS3');
      expect(distribution).toHaveTextContent('Third-party hosted0');
      expect(distribution).toHaveTextContent('Unknown4');
    });

    it('filters applications by hosting classification', async () => {
      const saas = createMockComponent({ id: 'comp-saas' as ComponentId, name: 'SaaS App', hosting: 'saas' });
      const onPrem = createMockComponent({
        id: 'comp-onprem' as ComponentId,
        name: 'OnPrem App',
        hosting: 'on-premises',
      });
      render(<ApplicationsSection {...defaultProps} components={[saas, onPrem]} />);
      await screen.findByTestId('hosting-distribution');

      expect(screen.getByText('SaaS App')).toBeInTheDocument();
      expect(screen.getByText('OnPrem App')).toBeInTheDocument();

      const user = (await import('@testing-library/user-event')).default;
      await user.click(screen.getByTestId('hosting-filter'));
      await user.click(await screen.findByRole('option', { name: 'SaaS', hidden: true }));

      expect(screen.getByText('SaaS App')).toBeInTheDocument();
      expect(screen.queryByText('OnPrem App')).not.toBeInTheDocument();
    });
  });
});
