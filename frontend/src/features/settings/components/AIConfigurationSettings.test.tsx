import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MantineProvider } from '@mantine/core';
import React from 'react';
import { AIConfigurationSettings } from './AIConfigurationSettings';
import { assistantConfigApi } from '../../../api/assistant/assistantConfigApi';
import type { AIConfigurationResponse } from '../../../api/assistant/types';

vi.mock('../../../api/assistant/assistantConfigApi', () => ({
  assistantConfigApi: {
    getConfig: vi.fn(),
    updateConfig: vi.fn(),
    testConnection: vi.fn(),
  },
}));

vi.mock('react-hot-toast', () => ({
  default: { error: vi.fn(), success: vi.fn() },
}));

function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });
}

function renderWithProviders(ui: React.ReactElement) {
  const client = createQueryClient();
  return render(
    <QueryClientProvider client={client}>
      <MantineProvider>
        {ui}
      </MantineProvider>
    </QueryClientProvider>
  );
}

const unconfiguredResponse: AIConfigurationResponse = {
  id: '',
  provider: 'openai',
  endpoint: '',
  apiKeyStatus: 'not_configured',
  model: '',
  maxTokens: 4096,
  temperature: 0.3,
  status: 'not_configured',
  updatedAt: new Date().toISOString(),
  _links: {
    self: { href: '/api/v1/assistant-config', method: 'GET' },
    update: { href: '/api/v1/assistant-config', method: 'PUT' },
  },
};

const configuredResponse: AIConfigurationResponse = {
  id: 'config-1',
  provider: 'openai',
  endpoint: 'https://api.openai.com',
  apiKeyStatus: 'configured',
  model: 'gpt-4o',
  maxTokens: 4096,
  temperature: 0.3,
  status: 'configured',
  updatedAt: new Date().toISOString(),
  _links: {
    self: { href: '/api/v1/assistant-config', method: 'GET' },
    update: { href: '/api/v1/assistant-config', method: 'PUT' },
    test: { href: '/api/v1/assistant-config/test', method: 'POST' },
  },
};

async function renderWithConfig(config: AIConfigurationResponse) {
  vi.mocked(assistantConfigApi.getConfig).mockResolvedValue(config);
  renderWithProviders(<AIConfigurationSettings />);
}

async function renderAndWaitForSaveButton(config: AIConfigurationResponse) {
  await renderWithConfig(config);
  await waitFor(() => {
    expect(screen.getByRole('button', { name: 'Save' })).toBeInTheDocument();
  });
}

async function renderAndWaitForTestButton(config: AIConfigurationResponse) {
  await renderWithConfig(config);
  await waitFor(() => {
    expect(screen.getByRole('button', { name: 'Test Connection' })).toBeInTheDocument();
  });
}

describe('AIConfigurationSettings', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders loading state initially', () => {
    vi.mocked(assistantConfigApi.getConfig).mockImplementation(
      () => new Promise(() => {})
    );

    renderWithProviders(<AIConfigurationSettings />);

    expect(screen.getByText(/loading ai configuration/i)).toBeInTheDocument();
  });

  it('renders error state when loading fails', async () => {
    vi.mocked(assistantConfigApi.getConfig).mockRejectedValue(
      new Error('Network error')
    );

    renderWithProviders(<AIConfigurationSettings />);

    await waitFor(() => {
      expect(screen.getByText(/network error/i)).toBeInTheDocument();
    });
  });

  it('renders the form after loading', async () => {
    await renderWithConfig(unconfiguredResponse);

    await waitFor(() => {
      expect(screen.getByText('AI Assistant Configuration')).toBeInTheDocument();
    });

    expect(screen.getByLabelText(/^Provider/)).toBeInTheDocument();
    expect(screen.getByLabelText(/base url override/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/^API Key/)).toBeInTheDocument();
    expect(screen.getByLabelText(/^Model/)).toBeInTheDocument();
  });

  it('renders the data residency banner', async () => {
    await renderWithConfig(unconfiguredResponse);

    await waitFor(() => {
      expect(screen.getByText(/data handling requirements/i)).toBeInTheDocument();
    });
  });

  it('renders provider dropdown with OpenAI and Anthropic options', async () => {
    await renderWithConfig(unconfiguredResponse);

    await waitFor(() => {
      expect(screen.getByLabelText(/^Provider/)).toBeInTheDocument();
    });

    const select = screen.getByLabelText(/^Provider/) as HTMLSelectElement;
    expect(select.options).toHaveLength(2);
    expect(select.options[0].text).toBe('OpenAI');
    expect(select.options[1].text).toBe('Anthropic');
  });

  it('disables save when required fields are missing', async () => {
    await renderAndWaitForSaveButton(unconfiguredResponse);
    expect(screen.getByRole('button', { name: 'Save' })).toBeDisabled();
  });

  it('does not show test connection button when not configured', async () => {
    await renderAndWaitForSaveButton(unconfiguredResponse);
    expect(screen.queryByRole('button', { name: 'Test Connection' })).not.toBeInTheDocument();
  });

  it('shows test connection button when configured', async () => {
    await renderAndWaitForTestButton(configuredResponse);
  });

  it('populates form fields from configured response', async () => {
    await renderWithConfig(configuredResponse);

    await waitFor(() => {
      const modelInput = screen.getByLabelText(/^Model/) as HTMLInputElement;
      expect(modelInput.value).toBe('gpt-4o');
    });

    const providerSelect = screen.getByLabelText(/^Provider/) as HTMLSelectElement;
    expect(providerSelect.value).toBe('openai');
  });

  it('shows API key configured status when already set', async () => {
    await renderWithConfig(configuredResponse);

    await waitFor(() => {
      expect(screen.getByText('API key configured')).toBeInTheDocument();
    });

    expect(screen.getByRole('button', { name: 'Change' })).toBeInTheDocument();
  });

  it('shows API key input when Change button is clicked', async () => {
    await renderWithConfig(configuredResponse);

    await waitFor(() => {
      expect(screen.getByText('API key configured')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole('button', { name: 'Change' }));

    expect(screen.getByLabelText(/^API Key/)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Cancel' })).toBeInTheDocument();
  });

  it('toggles advanced settings', async () => {
    await renderWithConfig(unconfiguredResponse);

    await waitFor(() => {
      expect(screen.getByText(/advanced settings/i)).toBeInTheDocument();
    });

    expect(screen.queryByLabelText('Max Tokens')).not.toBeInTheDocument();

    fireEvent.click(screen.getByText(/advanced settings/i));

    expect(screen.getByLabelText('Max Tokens')).toBeInTheDocument();
    expect(screen.getByLabelText('Temperature')).toBeInTheDocument();
    expect(screen.getByLabelText('System Prompt Override')).toBeInTheDocument();
  });

  it('calls updateConfig on save', async () => {
    vi.mocked(assistantConfigApi.updateConfig).mockResolvedValue(configuredResponse);
    await renderWithConfig(unconfiguredResponse);

    await waitFor(() => {
      expect(screen.getByLabelText(/^Model/)).toBeInTheDocument();
    });

    fireEvent.change(screen.getByLabelText(/^Model/), { target: { value: 'gpt-4o' } });
    fireEvent.change(screen.getByLabelText(/^API Key/), { target: { value: 'sk-test-key' } });

    fireEvent.click(screen.getByRole('button', { name: 'Save' }));

    await waitFor(() => {
      expect(assistantConfigApi.updateConfig).toHaveBeenCalledWith(
        expect.objectContaining({
          provider: 'openai',
          model: 'gpt-4o',
          apiKey: 'sk-test-key',
        })
      );
    });
  });

  it('calls testConnection when test button is clicked', async () => {
    vi.mocked(assistantConfigApi.testConnection).mockResolvedValue({
      success: true,
      model: 'gpt-4o',
      latencyMs: 150,
    });
    await renderAndWaitForTestButton(configuredResponse);

    fireEvent.click(screen.getByRole('button', { name: 'Test Connection' }));

    await waitFor(() => {
      expect(screen.getByText(/connection successful/i)).toBeInTheDocument();
    });
    expect(screen.getByText(/150ms/)).toBeInTheDocument();
  });

  it('shows test connection failure', async () => {
    vi.mocked(assistantConfigApi.testConnection).mockResolvedValue({
      success: false,
      error: 'Invalid API key',
    });
    await renderAndWaitForTestButton(configuredResponse);

    fireEvent.click(screen.getByRole('button', { name: 'Test Connection' }));

    await waitFor(() => {
      expect(screen.getByText(/connection failed/i)).toBeInTheDocument();
    });
    expect(screen.getByText(/invalid api key/i)).toBeInTheDocument();
  });

  it('disables save when model is empty even with API key configured', async () => {
    await renderAndWaitForSaveButton({ ...configuredResponse, model: '' });
    expect(screen.getByRole('button', { name: 'Save' })).toBeDisabled();
  });

  it('omits apiKey from save request when not changed on configured instance', async () => {
    vi.mocked(assistantConfigApi.updateConfig).mockResolvedValue(configuredResponse);
    await renderWithConfig(configuredResponse);

    await waitFor(() => {
      const modelInput = screen.getByLabelText(/^Model/) as HTMLInputElement;
      expect(modelInput.value).toBe('gpt-4o');
    });

    fireEvent.click(screen.getByRole('button', { name: 'Save' }));

    await waitFor(() => {
      expect(assistantConfigApi.updateConfig).toHaveBeenCalled();
    });

    const callArg = vi.mocked(assistantConfigApi.updateConfig).mock.calls[0][0];
    expect(callArg.apiKey).toBeUndefined();
  });

  it('updates model placeholder when provider changes to Anthropic', async () => {
    await renderWithConfig(unconfiguredResponse);

    await waitFor(() => {
      expect(screen.getByLabelText(/^Provider/)).toBeInTheDocument();
    });

    fireEvent.change(screen.getByLabelText(/^Provider/), { target: { value: 'anthropic' } });

    const modelInput = screen.getByLabelText(/^Model/) as HTMLInputElement;
    expect(modelInput.placeholder).toContain('claude');
  });
});
