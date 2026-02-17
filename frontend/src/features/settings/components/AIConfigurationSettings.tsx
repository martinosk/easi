import { Button } from '@mantine/core';
import { useAIConfiguration } from '../hooks/useAIConfiguration';
import { useAIConfigForm } from '../hooks/useAIConfigForm';
import type { AIConfigurationResponse, LLMProvider, TestConnectionResponse } from '../../../api/assistant/types';
import { APIKeyField } from './APIKeyField';
import { AdvancedSettings } from './AdvancedSettings';
import './AIConfigurationSettings.css';

const PROVIDER_DEFAULTS: Record<LLMProvider, string> = {
  openai: 'https://api.openai.com',
  anthropic: 'https://api.anthropic.com',
};

function TestResultBanner({ result }: { result: TestConnectionResponse }) {
  const className = `ai-config-test-result ${result.success ? 'success' : 'failure'}`;
  if (result.success) {
    return (
      <div className={className}>
        Connection successful. Model: {result.model}. Latency: {result.latencyMs}ms.
      </div>
    );
  }
  return <div className={className}>Connection failed: {result.error}</div>;
}

function AIConfigForm({ config }: { config: AIConfigurationResponse | undefined }) {
  const form = useAIConfigForm(config);

  return (
    <div className="ai-config-form">
      <div className="ai-config-field">
        <label htmlFor="ai-provider">Provider <span className="ai-config-required">*</span></label>
        <select
          id="ai-provider"
          value={form.fields.provider}
          onChange={(e) => form.updateField('provider', e.target.value as LLMProvider)}
        >
          <option value="openai">OpenAI</option>
          <option value="anthropic">Anthropic</option>
        </select>
      </div>

      <div className="ai-config-field">
        <label htmlFor="ai-endpoint">Base URL override (optional)</label>
        <input
          id="ai-endpoint"
          type="text"
          value={form.fields.endpoint}
          onChange={(e) => form.updateField('endpoint', e.target.value)}
          placeholder={PROVIDER_DEFAULTS[form.fields.provider]}
        />
        <p className="ai-config-field-hint">
          Leave empty to use the default provider endpoint
        </p>
      </div>

      <APIKeyField
        apiKeyStatus={config?.apiKeyStatus}
        apiKey={form.fields.apiKey}
        onApiKeyChange={(v) => form.updateField('apiKey', v)}
        showInput={form.apiKeyInput.showApiKeyInput}
        onShowInput={form.apiKeyInput.setShowApiKeyInput}
      />

      <div className="ai-config-field">
        <label htmlFor="ai-model">Model <span className="ai-config-required">*</span></label>
        <input
          id="ai-model"
          type="text"
          value={form.fields.model}
          onChange={(e) => form.updateField('model', e.target.value)}
          placeholder={form.fields.provider === 'anthropic' ? 'claude-sonnet-4-5-20250929' : 'gpt-4o'}
        />
      </div>

      <button
        className="ai-config-advanced-toggle"
        onClick={() => form.advanced.setShowAdvanced(!form.advanced.showAdvanced)}
        type="button"
      >
        {form.advanced.showAdvanced ? '\u25BC' : '\u25B6'} Advanced Settings
      </button>

      {form.advanced.showAdvanced && (
        <AdvancedSettings
          maxTokens={form.fields.maxTokens}
          onMaxTokensChange={(v) => form.updateField('maxTokens', v)}
          temperature={form.fields.temperature}
          onTemperatureChange={(v) => form.updateField('temperature', v)}
          systemPromptOverride={form.fields.systemPromptOverride}
          onSystemPromptOverrideChange={(v) => form.updateField('systemPromptOverride', v)}
        />
      )}

      {form.testResult && <TestResultBanner result={form.testResult} />}

      <div className="ai-config-actions">
        {form.isConfigured && (
          <Button
            variant="outline"
            onClick={form.handleTestConnection}
            loading={form.isTesting}
            disabled={form.isSaving}
          >
            Test Connection
          </Button>
        )}
        <Button
          onClick={form.handleSave}
          loading={form.isSaving}
          disabled={form.isSaveDisabled}
        >
          Save
        </Button>
      </div>
    </div>
  );
}

export function AIConfigurationSettings() {
  const { data: config, isLoading, error } = useAIConfiguration();

  if (isLoading) {
    return (
      <div className="ai-config-settings">
        <div className="loading-state">
          <div className="loading-spinner" />
          <p>Loading AI configuration...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="ai-config-settings">
        <div className="error-message">
          {error instanceof Error ? error.message : 'Failed to load AI configuration'}
        </div>
      </div>
    );
  }

  return (
    <div className="ai-config-settings">
      <div className="ai-config-header">
        <h2 className="ai-config-title">AI Assistant Configuration</h2>
        <p className="ai-config-description">
          Configure the LLM provider for your organization&apos;s architecture assistant.
        </p>
      </div>

      <div className="ai-config-banner">
        Architecture data will be sent to the configured LLM endpoint.
        Ensure compliance with your organization&apos;s data handling requirements.
      </div>

      <AIConfigForm config={config} />
    </div>
  );
}
