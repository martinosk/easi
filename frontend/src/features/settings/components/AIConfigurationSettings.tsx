import { Button, NativeSelect, Stack, Text, TextInput, UnstyledButton } from '@mantine/core';
import type { AIConfigurationResponse, LLMProvider, TestConnectionResponse } from '../../../api/assistant/types';
import { useAIConfigForm } from '../hooks/useAIConfigForm';
import { useAIConfiguration } from '../hooks/useAIConfiguration';
import { AdvancedSettings } from './AdvancedSettings';
import classes from './AIConfigurationSettings.module.css';
import { APIKeyField } from './APIKeyField';
import {
  SettingsSection,
  SettingsSectionError,
  SettingsSectionFooter,
  SettingsSectionHeader,
  SettingsSectionLoading,
} from './SettingsSection';

const PROVIDER_DEFAULTS: Record<LLMProvider, string> = {
  openai: 'https://api.openai.com',
  anthropic: 'https://api.anthropic.com',
};

function TestResultBanner({ result }: { result: TestConnectionResponse }) {
  const outcomeClass = result.success ? classes.testResultSuccess : classes.testResultFailure;
  return (
    <Text mt="sm" className={`${classes.testResult} ${outcomeClass}`}>
      {result.success
        ? `Connection successful. Model: ${result.model}. Latency: ${result.latencyMs}ms.`
        : `Connection failed: ${result.error}`}
    </Text>
  );
}

function AIConfigForm({ config }: { config: AIConfigurationResponse | undefined }) {
  const form = useAIConfigForm(config);

  return (
    <Stack gap="md">
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

      <TextInput
        id="ai-endpoint"
        label="Base URL override (optional)"
        value={form.fields.endpoint}
        onChange={(e) => form.updateField('endpoint', e.currentTarget.value)}
        placeholder={PROVIDER_DEFAULTS[form.fields.provider]}
        description="Leave empty to use the default provider endpoint"
      />

      <APIKeyField
        apiKeyStatus={config?.apiKeyStatus}
        apiKey={form.fields.apiKey}
        onApiKeyChange={(v) => form.updateField('apiKey', v)}
        showInput={form.apiKeyInput.showApiKeyInput}
        onShowInput={form.apiKeyInput.setShowApiKeyInput}
      />

      <TextInput
        id="ai-model"
        label="Model"
        withAsterisk
        value={form.fields.model}
        onChange={(e) => form.updateField('model', e.currentTarget.value)}
        placeholder={form.fields.provider === 'anthropic' ? 'claude-sonnet-4-5-20250929' : 'gpt-4o'}
      />

      <UnstyledButton
        component="button"
        type="button"
        className={classes.advancedToggle}
        onClick={() => form.advanced.setShowAdvanced(!form.advanced.showAdvanced)}
      >
        {form.advanced.showAdvanced ? '▼' : '▶'} Advanced Settings
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

      <SettingsSectionFooter>
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
      </SettingsSectionFooter>
    </Stack>
  );
}

export function AIConfigurationSettings() {
  const { data: config, isLoading, error } = useAIConfiguration();

  if (isLoading) return <SettingsSectionLoading message="Loading AI configuration..." />;
  if (error) return <SettingsSectionError error={error} fallback="Failed to load AI configuration" />;

  return (
    <SettingsSection>
      <SettingsSectionHeader
        title="AI Assistant Configuration"
        description="Configure the LLM provider for your organization's architecture assistant."
      />
      <Text className={classes.banner}>
        Architecture data will be sent to the configured LLM endpoint. Ensure compliance with your organization&apos;s
        data handling requirements.
      </Text>
      <AIConfigForm config={config} />
    </SettingsSection>
  );
}
