import type { ReactElement } from 'react';
import { screen } from '@testing-library/react';
import { renderWithProviders } from '../../../../test/helpers';
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
    it('should show the indicator when onePagerComplete is false', () => {
      const component = createMockComponent({ id: 'comp-123' as ComponentId, onePagerComplete: false });
      render(<ApplicationsSection {...defaultProps} components={[component]} />);

      expect(screen.getByTestId('one-pager-incomplete-comp-123')).toBeInTheDocument();
    });

    it('should not show the indicator when onePagerComplete is true', () => {
      const component = createMockComponent({ id: 'comp-123' as ComponentId, onePagerComplete: true });
      render(<ApplicationsSection {...defaultProps} components={[component]} />);

      expect(screen.queryByTestId('one-pager-incomplete-comp-123')).not.toBeInTheDocument();
    });

    it('should not show the indicator when onePagerComplete is absent', () => {
      const component = createMockComponent({ id: 'comp-123' as ComponentId });
      render(<ApplicationsSection {...defaultProps} components={[component]} />);

      expect(screen.queryByTestId('one-pager-incomplete-comp-123')).not.toBeInTheDocument();
    });
  });
});
