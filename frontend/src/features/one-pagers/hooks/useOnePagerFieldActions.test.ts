import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { BuiltInField, CustomField, OnePagerConfiguration } from '../types';
import { useOnePagerFieldActions } from './useOnePagerFieldActions';

vi.mock('../api/onePagersApi', () => ({
  onePagersApi: {
    changeFieldRequirement: vi.fn(),
    changeBuiltInFieldRequirement: vi.fn(),
    reorderFields: vi.fn(),
    includeBuiltInField: vi.fn(),
    excludeBuiltInField: vi.fn(),
    renameCustomField: vi.fn(),
    retireCustomField: vi.fn(),
    reactivateCustomField: vi.fn(),
    addSelectionOption: vi.fn(),
    retireSelectionOption: vi.fn(),
    setNumberFieldBounds: vi.fn(),
  },
}));

vi.mock('react-hot-toast', () => ({
  default: { success: vi.fn(), error: vi.fn() },
}));

import { onePagersApi } from '../api/onePagersApi';

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

function buildConfiguration(overrides: Partial<OnePagerConfiguration> = {}): OnePagerConfiguration {
  return {
    id: 'config-1',
    subjectType: 'application',
    builtInFields: [],
    customFields: [],
    displayOrder: [],
    version: 1,
    createdAt: '2026-01-01T00:00:00Z',
    modifiedAt: '2026-01-01T00:00:00Z',
    modifiedBy: 'admin@example.com',
    _links: {
      self: { href: '/api/v1/one-pagers/configurations/application', method: 'GET' },
      'x-impact-preview': { href: '/api/v1/one-pagers/configurations/application/impact-preview', method: 'GET' },
    },
    ...overrides,
  };
}

function buildCustomField(overrides: Partial<CustomField> = {}): CustomField {
  return {
    id: 'field-1',
    name: 'Contract link',
    type: 'link',
    required: false,
    helpText: '',
    active: true,
    _links: {
      'x-set-requirement': {
        href: '/api/v1/one-pagers/configurations/application/custom-fields/field-1/requirement',
        method: 'PUT',
      },
    },
    ...overrides,
  };
}

function buildBuiltInField(overrides: Partial<BuiltInField> = {}): BuiltInField {
  return {
    id: 'experts',
    label: 'Experts',
    included: true,
    required: false,
    _links: {
      'x-set-requirement': {
        href: '/api/v1/one-pagers/configurations/application/built-in-fields/experts/requirement',
        method: 'PUT',
      },
    },
    ...overrides,
  };
}

describe('useOnePagerFieldActions', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    vi.clearAllMocks();
    queryClient = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  function renderActions(
    configuration: OnePagerConfiguration,
    onRequireConfirmationNeeded = vi.fn(),
    onRequireBuiltInConfirmationNeeded = vi.fn(),
  ) {
    const onRename = vi.fn();
    const hook = renderHook(
      () =>
        useOnePagerFieldActions(
          'application',
          configuration,
          onRename,
          onRequireConfirmationNeeded,
          onRequireBuiltInConfirmationNeeded,
        ),
      { wrapper: createWrapper(queryClient) },
    );
    return { ...hook, onRename, onRequireConfirmationNeeded, onRequireBuiltInConfirmationNeeded };
  }

  type RenderedActions = ReturnType<typeof renderActions>;

  it.each([
    {
      label: 'a field',
      field: buildCustomField({ required: false }) as CustomField | BuiltInField,
      toggle: (hooks: RenderedActions, field: CustomField | BuiltInField) =>
        hooks.result.current.fieldActions.onToggleRequired(field as CustomField, true),
      confirmedVia: (hooks: RenderedActions) => hooks.onRequireConfirmationNeeded,
      untouchedApi: () => onePagersApi.changeFieldRequirement,
    },
    {
      label: 'a built-in field',
      field: buildBuiltInField({ required: false }) as CustomField | BuiltInField,
      toggle: (hooks: RenderedActions, field: CustomField | BuiltInField) =>
        hooks.result.current.fieldActions.onToggleBuiltInRequired(field as BuiltInField, true),
      confirmedVia: (hooks: RenderedActions) => hooks.onRequireBuiltInConfirmationNeeded,
      untouchedApi: () => onePagersApi.changeBuiltInFieldRequirement,
    },
  ])('intercepts flipping $label to required instead of mutating immediately', ({ field, toggle, confirmedVia, untouchedApi }) => {
    const hooks = renderActions(buildConfiguration());

    act(() => {
      toggle(hooks, field);
    });

    expect(confirmedVia(hooks)).toHaveBeenCalledWith(field);
    expect(untouchedApi()).not.toHaveBeenCalled();
  });

  it('unrequires a built-in field immediately without a confirmation dialog', async () => {
    const configuration = buildConfiguration();
    const field = buildBuiltInField({ required: true });
    const onRequireBuiltIn = vi.fn();
    vi.mocked(onePagersApi.changeBuiltInFieldRequirement).mockResolvedValue(configuration);
    const { result } = renderActions(configuration, vi.fn(), onRequireBuiltIn);

    act(() => {
      result.current.fieldActions.onToggleBuiltInRequired(field, false);
    });

    await waitFor(() =>
      expect(onePagersApi.changeBuiltInFieldRequirement).toHaveBeenCalledWith(field, { required: false, version: 1 }),
    );
    expect(onRequireBuiltIn).not.toHaveBeenCalled();
  });

  it.each([
    {
      label: 'flipping a field to optional',
      configuration: buildConfiguration(),
      field: buildCustomField({ required: true }),
      toggleTo: false,
    },
    {
      label: 'the x-impact-preview link is absent (defensive fallback)',
      configuration: buildConfiguration({ _links: { self: { href: '/x', method: 'GET' } } }),
      field: buildCustomField({ required: false }),
      toggleTo: true,
    },
  ])('mutates directly with no interception when $label', async ({ configuration, field, toggleTo }) => {
    vi.mocked(onePagersApi.changeFieldRequirement).mockResolvedValue(configuration);
    const { result, onRequireConfirmationNeeded } = renderActions(configuration);

    act(() => {
      result.current.fieldActions.onToggleRequired(field, toggleTo);
    });

    await waitFor(() =>
      expect(onePagersApi.changeFieldRequirement).toHaveBeenCalledWith(field, { required: toggleTo, version: 1 }),
    );
    expect(onRequireConfirmationNeeded).not.toHaveBeenCalled();
  });

  it('onSetBounds mutates with the field and current configuration version', async () => {
    const configuration = buildConfiguration();
    const field = buildCustomField({
      type: 'number',
      _links: { 'x-set-bounds': { href: '/api/v1/one-pagers/configurations/application/custom-fields/field-1/bounds', method: 'PUT' } },
    });
    vi.mocked(onePagersApi.setNumberFieldBounds).mockResolvedValue(configuration);
    const { result } = renderActions(configuration);

    act(() => {
      result.current.fieldActions.onSetBounds(field, 0, 5);
    });

    await waitFor(() =>
      expect(onePagersApi.setNumberFieldBounds).toHaveBeenCalledWith(field, { min: 0, max: 5, version: 1 }),
    );
  });

  it('confirmRequireField mutates with required true and reports completion', async () => {
    const configuration = buildConfiguration();
    const field = buildCustomField({ required: false });
    vi.mocked(onePagersApi.changeFieldRequirement).mockResolvedValue(configuration);
    const { result } = renderActions(configuration);
    const onDone = vi.fn();

    act(() => {
      result.current.confirmRequireField(field, onDone);
    });

    await waitFor(() =>
      expect(onePagersApi.changeFieldRequirement).toHaveBeenCalledWith(field, { required: true, version: 1 }),
    );
    await waitFor(() => expect(onDone).toHaveBeenCalled());
  });
});
