import { act, renderHook } from '@testing-library/react';
import type React from 'react';
import { describe, expect, it, vi } from 'vitest';
import type { BusinessDomain, HATEOASLinks } from '../../../api/types';
import { useDomainContextMenu } from './useDomainContextMenu';

function domain(links: HATEOASLinks): BusinessDomain {
  return {
    id: 'bd-1',
    name: 'Customer Engagement',
    description: '',
    capabilityCount: 0,
    createdAt: '2026-01-01T00:00:00Z',
    _links: links,
  } as BusinessDomain;
}

function openMenuFor(target: BusinessDomain) {
  const { result } = renderHook(() => useDomainContextMenu({ onEdit: vi.fn(), onDelete: vi.fn() }));
  act(() => result.current.handleContextMenu({ clientX: 1, clientY: 2 } as React.MouseEvent, target));
  return result;
}

describe('useDomainContextMenu stewards entry', () => {
  it('offers Stewards when the domain carries x-stewardships and opens them for the domain', () => {
    const href = '/api/v1/stewardships?domainId=bd-1';
    const result = openMenuFor(domain({ 'x-stewardships': { href, method: 'GET' } }));

    const items = result.current.getContextMenuItems(result.current.contextMenu!);
    const stewards = items.find((item) => item.label === 'Stewards...');
    expect(stewards).toBeDefined();

    act(() => stewards!.onClick());

    expect(result.current.domainForStewards).toEqual({
      domainId: 'bd-1',
      domainName: 'Customer Engagement',
      stewardshipsHref: href,
    });
  });

  it('places Stewards next to Invite to Edit', () => {
    const result = openMenuFor(
      domain({
        'x-edit-grants': { href: '/api/v1/edit-grants', method: 'POST' },
        'x-stewardships': { href: '/api/v1/stewardships?domainId=bd-1', method: 'GET' },
      }),
    );

    const labels = result.current.getContextMenuItems(result.current.contextMenu!).map((item) => item.label);
    expect(labels.indexOf('Stewards...')).toBe(labels.indexOf('Invite to Edit...') + 1);
  });

  it('offers no Stewards without the link', () => {
    const result = openMenuFor(domain({}));

    const labels = result.current.getContextMenuItems(result.current.contextMenu!).map((item) => item.label);
    expect(labels).not.toContain('Stewards...');
  });
});
