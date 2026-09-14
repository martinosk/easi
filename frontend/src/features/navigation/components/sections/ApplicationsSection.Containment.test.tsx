import { screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { ComponentId } from '../../../../api/types';
import { renderWithProviders } from '../../../../test/helpers';
import { buildComponent } from '../../../../test/helpers/entityBuilders';
import { ApplicationsSection } from './ApplicationsSection';

const multiSelect = {
  isMultiSelected: () => false,
  handleItemClick: () => 'single' as const,
  handleContextMenu: () => false,
  handleDragStart: () => false,
  selectedItems: [],
};

const props = {
  currentView: null,
  selectedNodeId: null,
  isExpanded: true,
  onToggle: vi.fn(),
  onComponentContextMenu: vi.fn(),
  editingState: null,
  setEditingState: vi.fn(),
  onRenameSubmit: vi.fn(),
  editInputRef: { current: null },
  multiSelect,
};

describe('ApplicationsSection containment annotation', () => {
  it('annotates parts with their parent and parents with their part count', () => {
    const components = [
      buildComponent({
        id: 'quoting' as ComponentId,
        name: 'Quoting',
        partOf: { id: 'crm' as ComponentId, name: 'CRM Suite', kind: 'composition' },
      }),
      buildComponent({
        id: 'crm' as ComponentId,
        name: 'CRM Suite',
        parts: [
          { id: 'quoting' as ComponentId, name: 'Quoting', kind: 'composition' },
          { id: 'billing' as ComponentId, name: 'Billing', kind: 'aggregation' },
        ],
      }),
      buildComponent({ id: 'standalone' as ComponentId, name: 'Standalone' }),
    ];

    renderWithProviders(<ApplicationsSection {...props} components={components} />, { withRouter: false });

    const annotations = screen.getAllByTestId('containment-annotation').map((node) => node.textContent);
    expect(annotations).toEqual(['in CRM Suite', '2 parts']);
  });
});
