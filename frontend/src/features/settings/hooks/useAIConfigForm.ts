import { useEffect, useState } from 'react';
import type { AIConfigurationResponse, LLMProvider, TestConnectionRequest, TestConnectionResponse } from '../../../api/assistant/types';
import { useTestAIConnection, useUpdateAIConfiguration } from './useAIConfiguration';

interface FormFields {
  provider: LLMProvider;
  endpoint: string;
  apiKey: string;
  model: string;
  maxTokens: number;
  temperature: number;
  systemPromptOverride: string;
}

const DEFAULT_FIELDS: FormFields = {
  provider: 'openai',
  endpoint: '',
  apiKey: '',
  model: '',
  maxTokens: 4096,
  temperature: 0.3,
  systemPromptOverride: '',
};

function deriveFieldsFromConfig(config: AIConfigurationResponse): FormFields {
  return {
    provider: config.provider || DEFAULT_FIELDS.provider,
    endpoint: config.endpoint || '',
    apiKey: '',
    model: config.model || '',
    maxTokens: config.maxTokens || DEFAULT_FIELDS.maxTokens,
    temperature: config.temperature ?? DEFAULT_FIELDS.temperature,
    systemPromptOverride: config.systemPromptOverride || '',
  };
}

function toSaveRequest(fields: FormFields) {
  return {
    ...fields,
    apiKey: fields.apiKey || undefined,
    systemPromptOverride: fields.systemPromptOverride || null,
  };
}

function isApiKeyRequired(config: AIConfigurationResponse | undefined, apiKey: string): boolean {
  return config?.apiKeyStatus !== 'configured' && !apiKey;
}

function isConnectionTestable(
  fields: Pick<FormFields, 'provider' | 'model' | 'apiKey'>,
  config: AIConfigurationResponse | undefined
): boolean {
  return !!fields.provider && !!fields.model && (!!fields.apiKey || config?.apiKeyStatus === 'configured');
}

export function useAIConfigForm(config: AIConfigurationResponse | undefined) {
  const updateMutation = useUpdateAIConfiguration();
  const testMutation = useTestAIConnection();

  const [fields, setFields] = useState<FormFields>(DEFAULT_FIELDS);
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [showApiKeyInput, setShowApiKeyInput] = useState(false);
  const [testResult, setTestResult] = useState<TestConnectionResponse | null>(null);

  useEffect(() => {
    if (config) {
      setFields(deriveFieldsFromConfig(config));
      setShowApiKeyInput(config.apiKeyStatus !== 'configured');
    }
  }, [config]);

  const updateField = <K extends keyof FormFields>(key: K, value: FormFields[K]) =>
    setFields((prev) => ({ ...prev, [key]: value }));

  const handleSave = async () => {
    setTestResult(null);
    try {
      await updateMutation.mutateAsync(toSaveRequest(fields));
      updateField('apiKey', '');
      setShowApiKeyInput(false);
    } catch {
      // Error is handled by the mutation's onError callback (toast notification)
    }
  };

  const handleTestConnection = async () => {
    setTestResult(null);
    const req: TestConnectionRequest = {
      provider: fields.provider,
      endpoint: fields.endpoint,
      model: fields.model,
      apiKey: fields.apiKey || undefined,
    };
    try {
      const result = await testMutation.mutateAsync(req);
      setTestResult(result);
    } catch (err) {
      setTestResult({ success: false, error: err instanceof Error ? err.message : 'Connection test failed' });
    }
  };

  const needsApiKey = isApiKeyRequired(config, fields.apiKey);
  const isSaveDisabled = !fields.provider || !fields.model || needsApiKey;

  return {
    fields,
    updateField,
    advanced: { showAdvanced, setShowAdvanced },
    apiKeyInput: { showApiKeyInput, setShowApiKeyInput },
    testResult,
    handleSave,
    handleTestConnection,
    isSaveDisabled,
    isSaving: updateMutation.isPending,
    isTesting: testMutation.isPending,
    isTestable: isConnectionTestable(fields, config),
  };
}
