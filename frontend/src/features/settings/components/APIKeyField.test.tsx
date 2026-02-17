import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MantineProvider } from '@mantine/core';
import React from 'react';
import { APIKeyField } from './APIKeyField';

function renderWithMantine(ui: React.ReactElement) {
  return render(<MantineProvider>{ui}</MantineProvider>);
}

describe('APIKeyField', () => {
  it('shows configured status and Change button when API key is configured', () => {
    renderWithMantine(
      <APIKeyField
        apiKeyStatus="configured"
        apiKey=""
        onApiKeyChange={vi.fn()}
        showInput={false}
        onShowInput={vi.fn()}
      />
    );

    expect(screen.getByText('API key configured')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Change' })).toBeInTheDocument();
    expect(screen.queryByLabelText('API Key')).not.toBeInTheDocument();
  });

  it('calls onShowInput when Change is clicked', () => {
    const onShowInput = vi.fn();
    renderWithMantine(
      <APIKeyField
        apiKeyStatus="configured"
        apiKey=""
        onApiKeyChange={vi.fn()}
        showInput={false}
        onShowInput={onShowInput}
      />
    );

    fireEvent.click(screen.getByRole('button', { name: 'Change' }));
    expect(onShowInput).toHaveBeenCalledWith(true);
  });

  it('shows input field when not configured', () => {
    renderWithMantine(
      <APIKeyField
        apiKeyStatus="not_configured"
        apiKey=""
        onApiKeyChange={vi.fn()}
        showInput={true}
        onShowInput={vi.fn()}
      />
    );

    expect(screen.getByLabelText('API Key')).toBeInTheDocument();
  });

  it('shows input with Cancel button when configured and showInput is true', () => {
    renderWithMantine(
      <APIKeyField
        apiKeyStatus="configured"
        apiKey=""
        onApiKeyChange={vi.fn()}
        showInput={true}
        onShowInput={vi.fn()}
      />
    );

    expect(screen.getByLabelText('API Key')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Cancel' })).toBeInTheDocument();
  });

  it('calls onApiKeyChange when input value changes', () => {
    const onApiKeyChange = vi.fn();
    renderWithMantine(
      <APIKeyField
        apiKeyStatus="not_configured"
        apiKey=""
        onApiKeyChange={onApiKeyChange}
        showInput={true}
        onShowInput={vi.fn()}
      />
    );

    fireEvent.change(screen.getByLabelText('API Key'), { target: { value: 'sk-new-key' } });
    expect(onApiKeyChange).toHaveBeenCalledWith('sk-new-key');
  });

  it('hides input and clears value when Cancel is clicked', () => {
    const onShowInput = vi.fn();
    const onApiKeyChange = vi.fn();
    renderWithMantine(
      <APIKeyField
        apiKeyStatus="configured"
        apiKey="partial-key"
        onApiKeyChange={onApiKeyChange}
        showInput={true}
        onShowInput={onShowInput}
      />
    );

    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }));
    expect(onShowInput).toHaveBeenCalledWith(false);
    expect(onApiKeyChange).toHaveBeenCalledWith('');
  });
});
