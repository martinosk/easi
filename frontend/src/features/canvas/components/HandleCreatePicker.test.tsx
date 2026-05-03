import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { RelatedLink } from '../../../utils/xRelated';
import { HandleCreatePicker } from './HandleCreatePicker';

const entry = (overrides: Partial<RelatedLink> = {}): RelatedLink => ({
  href: '/api/v1/components',
  methods: ['POST'],
  title: 'Component (triggers)',
  targetType: 'component',
  relationType: 'component-triggers',
  ...overrides,
});

describe('HandleCreatePicker', () => {
  it('returns null when there are no entries', () => {
    const { container } = render(
      <HandleCreatePicker x={10} y={20} entries={[]} onSelect={() => {}} onClose={() => {}} />,
    );
    expect(container.firstChild).toBeNull();
  });

  it('renders one menu item per entry, labelled by the backend-supplied title', () => {
    const entries = [
      entry({ relationType: 'component-triggers', title: 'Component (triggers)' }),
      entry({ relationType: 'component-serves', title: 'Component (serves)' }),
      entry({ relationType: 'capability-parent', title: 'Capability (child of)', targetType: 'capability' }),
      entry({ relationType: 'origin-acquired-via', title: 'Acquired Entity (acquired-via)', targetType: 'acquiredEntity' }),
    ];
    render(<HandleCreatePicker x={0} y={0} entries={entries} onSelect={() => {}} onClose={() => {}} />);

    expect(screen.getByRole('menuitem', { name: 'Component (triggers)' })).toBeTruthy();
    expect(screen.getByRole('menuitem', { name: 'Component (serves)' })).toBeTruthy();
    expect(screen.getByRole('menuitem', { name: 'Capability (child of)' })).toBeTruthy();
    expect(screen.getByRole('menuitem', { name: 'Acquired Entity (acquired-via)' })).toBeTruthy();
  });

  it('invokes onSelect with the chosen entry, no string mutation, no synthesized variants', () => {
    const onSelect = vi.fn();
    const onClose = vi.fn();
    const triggers = entry({ relationType: 'component-triggers', title: 'Component (triggers)' });
    render(<HandleCreatePicker x={0} y={0} entries={[triggers]} onSelect={onSelect} onClose={onClose} />);

    fireEvent.click(screen.getByRole('menuitem', { name: 'Component (triggers)' }));

    expect(onSelect).toHaveBeenCalledWith({ entry: triggers });
    expect(onClose).toHaveBeenCalledTimes(1);
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

  it('does not render a Cancel control — outside-click handles cancellation', () => {
    render(<HandleCreatePicker x={0} y={0} entries={[entry()]} onSelect={() => {}} onClose={() => {}} />);
    expect(screen.queryByRole('menuitem', { name: /cancel/i })).toBeNull();
  });
});
