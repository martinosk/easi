import { screen, waitFor } from '@testing-library/react';
import type { ReactElement } from 'react';
import { describe, expect, it, vi } from 'vitest';
import type { Component, ComponentId, HATEOASLinks } from '../../../../api/types';
import { renderWithProviders } from '../../../../test/helpers';
import { seedOnePagerCompleteness } from '../../../../test/mocks/onePagerCompleteness';
import { ApplicationsSection } from './ApplicationsSection';

const render = (ui: ReactElement) => renderWithProviders(ui, { withRouter: false });

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

  describe('filters', () => {
    const user = () => import('@testing-library/user-event').then((m) => m.default);

    it('does not render the ownership and hosting summary', () => {
      render(<ApplicationsSection {...defaultProps} components={[createMockComponent()]} />);

      expect(screen.queryByTestId('statistics-summary')).not.toBeInTheDocument();
      expect(screen.queryByText('Ownership')).not.toBeInTheDocument();
      expect(screen.queryByText('Hosting')).not.toBeInTheDocument();
    });

    it('keeps the filters behind a filter icon next to the search box', () => {
      render(<ApplicationsSection {...defaultProps} components={[createMockComponent()]} />);

      expect(screen.getByLabelText('Toggle application filters')).toBeInTheDocument();
      expect(screen.queryByText('Filter by ownership')).not.toBeInTheDocument();
      expect(screen.queryByText('Filter by hosting')).not.toBeInTheDocument();
    });

    it.each([
      {
        dimension: 'ownership state',
        chip: 'Owned',
        matching: createMockComponent({ id: 'comp-owned' as ComponentId, name: 'Owned App', ownershipState: 'owned' }),
        other: createMockComponent({ id: 'comp-orphan' as ComponentId, name: 'Orphan App' }),
      },
      {
        dimension: 'hosting classification',
        chip: 'SaaS',
        matching: createMockComponent({ id: 'comp-saas' as ComponentId, name: 'SaaS App', hosting: 'saas' }),
        other: createMockComponent({ id: 'comp-onprem' as ComponentId, name: 'OnPrem App', hosting: 'on-premises' }),
      },
    ])('filters applications by $dimension', async ({ chip, matching, other }) => {
      render(<ApplicationsSection {...defaultProps} components={[matching, other]} />);
      const u = await user();

      await u.click(screen.getByLabelText('Toggle application filters'));
      await u.click(await screen.findByText(chip));

      expect(screen.getByText(matching.name)).toBeInTheDocument();
      expect(screen.queryByText(other.name)).not.toBeInTheDocument();
    });

    it('shows the filtered count in the section header', async () => {
      const saas = createMockComponent({ id: 'comp-saas' as ComponentId, name: 'SaaS App', hosting: 'saas' });
      const cloud = createMockComponent({ id: 'comp-cloud' as ComponentId, name: 'Cloud App', hosting: 'cloud' });
      render(<ApplicationsSection {...defaultProps} components={[saas, cloud]} />);
      const u = await user();
      const header = screen.getByRole('button', { name: /Applications/ });
      expect(header).toHaveTextContent('Applications2');

      await u.click(screen.getByLabelText('Toggle application filters'));
      await u.click(await screen.findByText('SaaS'));

      expect(header).toHaveTextContent('Applications1');
    });

    it('keeps the header count independent of the search text', async () => {
      render(
        <ApplicationsSection
          {...defaultProps}
          components={[createMockComponent(), createMockComponent({ id: 'comp-2' as ComponentId, name: 'Other' })]}
        />,
      );
      const u = await user();

      await u.type(screen.getByPlaceholderText('Search applications...'), 'Other');

      expect(screen.getByRole('button', { name: /Applications/ })).toHaveTextContent('Applications2');
    });

    it('shows a no-matches message when filters exclude every application', async () => {
      render(<ApplicationsSection {...defaultProps} components={[createMockComponent({ hosting: 'cloud' })]} />);
      const u = await user();

      await u.click(screen.getByLabelText('Toggle application filters'));
      await u.click(await screen.findByText('SaaS'));

      expect(screen.getByText('No matches')).toBeInTheDocument();
    });
  });
});
