import type { ReactElement } from 'react';
import { screen, waitFor } from '@testing-library/react';
import { renderWithProviders } from '../../../../test/helpers';
import { seedOnePagerCompleteness } from '../../../../test/mocks/onePagerCompleteness';
const render = (ui: ReactElement) => renderWithProviders(ui, { withRouter: false });
import { describe, expect, it, vi } from 'vitest';
import type { Component, ComponentId, HATEOASLinks } from '../../../../api/types';
import { ApplicationsSection } from './ApplicationsSection';

describe('ApplicationsSection', () => {
  const mockLinks: HATEOASLinks = { self: { href: '/test', method: 'GET' } };

  const createMockComponent = (overrides: Partial<Component> = {}): Component => ({
    id: 'comp-123' as ComponentId,
    name: 'Billing System',
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
});
