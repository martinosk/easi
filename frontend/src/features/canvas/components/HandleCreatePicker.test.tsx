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

  it('renders one menu item per entry, labelled by title', () => {
    const entries = [
      entry({ relationType: 'component-relation', title: 'Component (related)' }),
      entry({
        relationType: 'capability-realization',
        title: 'Component (realization)',
        targetType: 'component',
      }),
    ];
    render(<HandleCreatePicker x={0} y={0} entries={entries} onSelect={() => {}} onClose={() => {}} />);
    expect(screen.getByRole('menuitem', { name: 'Component (related)' })).toBeTruthy();
    expect(screen.getByRole('menuitem', { name: 'Component (realization)' })).toBeTruthy();
  });

  it('invokes onSelect with the chosen entry and then closes', () => {
    const onSelect = vi.fn();
    const onClose = vi.fn();
    const target = entry({ relationType: 'capability-realization', title: 'Component (realization)' });
    render(
      <HandleCreatePicker x={0} y={0} entries={[target]} onSelect={onSelect} onClose={onClose} />,
    );
    fireEvent.click(screen.getByRole('menuitem', { name: 'Component (realization)' }));
    expect(onSelect).toHaveBeenCalledWith(target);
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

  it('renders a Cancel control that calls onClose without onSelect', () => {
    const onSelect = vi.fn();
    const onClose = vi.fn();
    render(<HandleCreatePicker x={0} y={0} entries={[entry()]} onSelect={onSelect} onClose={onClose} />);
    fireEvent.click(screen.getByRole('button', { name: /cancel/i }));
    expect(onSelect).not.toHaveBeenCalled();
    expect(onClose).toHaveBeenCalled();
  });

  it('positions the popover at the supplied coordinates', () => {
    render(<HandleCreatePicker x={123} y={456} entries={[entry()]} onSelect={() => {}} onClose={() => {}} />);
    const popover = screen.getByTestId('handle-create-picker');
    expect(popover.style.left).toBe('123px');
    expect(popover.style.top).toBe('456px');
  });
});
