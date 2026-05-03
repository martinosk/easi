import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { RelatedLink } from '../../../utils/xRelated';
import { HandleCreatePicker } from './HandleCreatePicker';

const entry = (overrides: Partial<RelatedLink> = {}): RelatedLink => ({
  href: '/api/v1/components',
  methods: ['POST'],
  title: 'Component (related)',
  targetType: 'component',
  relationType: 'component-relation',
  ...overrides,
});

describe('HandleCreatePicker', () => {
  it('returns null when there are no entries', () => {
    const { container } = render(
      <HandleCreatePicker x={10} y={20} entries={[]} onSelect={() => {}} onClose={() => {}} />,
    );
    expect(container.firstChild).toBeNull();
  });

  it('renders Triggers and Serves variants for component-relation entries', () => {
    render(
      <HandleCreatePicker
        x={0}
        y={0}
        entries={[entry({ relationType: 'component-relation', title: 'Component (related)' })]}
        onSelect={() => {}}
        onClose={() => {}}
      />,
    );
    expect(screen.getByRole('menuitem', { name: 'Component (Triggers)' })).toBeTruthy();
    expect(screen.getByRole('menuitem', { name: 'Component (Serves)' })).toBeTruthy();
  });

  it('renders one menu item per non-component-relation entry, labelled by title', () => {
    const entries = [
      entry({ relationType: 'capability-realization', title: 'Component (realization)', targetType: 'component' }),
      entry({ relationType: 'capability-parent', title: 'Capability (child of)', targetType: 'capability' }),
    ];
    render(<HandleCreatePicker x={0} y={0} entries={entries} onSelect={() => {}} onClose={() => {}} />);
    expect(screen.getByRole('menuitem', { name: 'Component (realization)' })).toBeTruthy();
    expect(screen.getByRole('menuitem', { name: 'Capability (child of)' })).toBeTruthy();
  });

  it('invokes onSelect with the chosen entry and Triggers when Triggers variant clicked', () => {
    const onSelect = vi.fn();
    const onClose = vi.fn();
    render(
      <HandleCreatePicker
        x={0}
        y={0}
        entries={[entry({ relationType: 'component-relation', title: 'Component (related)' })]}
        onSelect={onSelect}
        onClose={onClose}
      />,
    );
    fireEvent.click(screen.getByRole('menuitem', { name: 'Component (Triggers)' }));
    expect(onSelect).toHaveBeenCalledWith(
      expect.objectContaining({
        relationSubType: 'Triggers',
        entry: expect.objectContaining({ relationType: 'component-relation', title: 'Component (Triggers)' }),
      }),
    );
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('invokes onSelect without a sub-type for non-component-relation entries', () => {
    const onSelect = vi.fn();
    const target = entry({ relationType: 'capability-realization', title: 'Component (realization)' });
    render(<HandleCreatePicker x={0} y={0} entries={[target]} onSelect={onSelect} onClose={() => {}} />);
    fireEvent.click(screen.getByRole('menuitem', { name: 'Component (realization)' }));
    expect(onSelect).toHaveBeenCalledWith({ entry: target });
  });

  it('calls onClose on Escape', () => {
    const onClose = vi.fn();
    render(<HandleCreatePicker x={0} y={0} entries={[entry()]} onSelect={() => {}} onClose={onClose} />);
    fireEvent.keyDown(document, { key: 'Escape' });
    expect(onClose).toHaveBeenCalled();
  });

  it('calls onClose on outside mousedown', () => {
    const onClose = vi.fn();
    render(<HandleCreatePicker x={0} y={0} entries={[entry()]} onSelect={() => {}} onClose={onClose} />);
    const outside = document.createElement('div');
    document.body.appendChild(outside);
    fireEvent.mouseDown(outside);
    expect(onClose).toHaveBeenCalled();
    document.body.removeChild(outside);
  });

  it('renders a Cancel control that calls onClose without onSelect', () => {
    const onSelect = vi.fn();
    const onClose = vi.fn();
    render(<HandleCreatePicker x={0} y={0} entries={[entry()]} onSelect={onSelect} onClose={onClose} />);
    fireEvent.click(screen.getByRole('menuitem', { name: /cancel/i }));
    expect(onSelect).not.toHaveBeenCalled();
    expect(onClose).toHaveBeenCalled();
  });
});
