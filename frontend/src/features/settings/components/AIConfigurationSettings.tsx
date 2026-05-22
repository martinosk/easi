import { Button, NativeSelect, TextInput, UnstyledButton } from '@mantine/core';
import type { AIConfigurationResponse, LLMProvider, TestConnectionResponse } from '../../../api/assistant/types';
import { useAIConfigForm } from '../hooks/useAIConfigForm';
import { useAIConfiguration } from '../hooks/useAIConfiguration';
import { AdvancedSettings } from './AdvancedSettings';
import { APIKeyField } from './APIKeyField';
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
        <NativeSelect
          id="ai-provider"
          label="Provider"
          withAsterisk
          data={[
            { value: 'openai', label: 'OpenAI' },
            { value: 'anthropic', label: 'Anthropic' },
          ]}
          value={form.fields.provider}
          onChange={(e) => form.updateField('provider', e.currentTarget.value as LLMProvider)}
        />
      </div>

      <div className="ai-config-field">
        <TextInput
          id="ai-endpoint"
          label="Base URL override (optional)"
          value={form.fields.endpoint}
          onChange={(e) => form.updateField('endpoint', e.currentTarget.value)}
          placeholder={PROVIDER_DEFAULTS[form.fields.provider]}
          description="Leave empty to use the default provider endpoint"
        />
      </div>

      <APIKeyField
        apiKeyStatus={config?.apiKeyStatus}
        apiKey={form.fields.apiKey}
        onApiKeyChange={(v) => form.updateField('apiKey', v)}
        showInput={form.apiKeyInput.showApiKeyInput}
        onShowInput={form.apiKeyInput.setShowApiKeyInput}
      />

      <div className="ai-config-field">
        <TextInput
          id="ai-model"
          label="Model"
          withAsterisk
          value={form.fields.model}
          onChange={(e) => form.updateField('model', e.currentTarget.value)}
          placeholder={form.fields.provider === 'anthropic' ? 'claude-sonnet-4-5-20250929' : 'gpt-4o'}
        />
      </div>

      <UnstyledButton
        component="button"
        type="button"
        className="ai-config-advanced-toggle"
        onClick={() => form.advanced.setShowAdvanced(!form.advanced.showAdvanced)}
      >
        {form.advanced.showAdvanced ? '\u25BC' : '\u25B6'} Advanced Settings
      </UnstyledButton>

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
        {form.isTestable && (
          <Button
            variant="outline"
            onClick={form.handleTestConnection}
            loading={form.isTesting}
            disabled={form.isSaving}
          >
            Test Connection
          </Button>
        )}
        <Button onClick={form.handleSave} loading={form.isSaving} disabled={form.isSaveDisabled}>
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
        Architecture data will be sent to the configured LLM endpoint. Ensure compliance with your organization&apos;s
        data handling requirements.
      </div>

      <AIConfigForm config={config} />
    </div>
  );
}
