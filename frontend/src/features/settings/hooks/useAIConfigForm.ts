import { useState, useEffect } from 'react';
import {
  useUpdateAIConfiguration,
  useTestAIConnection,
} from './useAIConfiguration';
import type {
  LLMProvider,
  AIConfigurationResponse,
  TestConnectionResponse,
} from '../../../api/assistant/types';

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
  provider: 'openai', endpoint: '', apiKey: '', model: '',
  maxTokens: 4096, temperature: 0.3, systemPromptOverride: '',
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
    await updateMutation.mutateAsync(toSaveRequest(fields));
    updateField('apiKey', '');
    setShowApiKeyInput(false);
  };

  const handleTestConnection = async () => {
    setTestResult(null);
    const result = await testMutation.mutateAsync();
    setTestResult(result);
  };

  const needsApiKey = config?.apiKeyStatus !== 'configured' && !fields.apiKey;
  const isSaveDisabled = !fields.provider || !fields.model || needsApiKey;

  return {
    fields, updateField,
    advanced: { showAdvanced, setShowAdvanced },
    apiKeyInput: { showApiKeyInput, setShowApiKeyInput },
    testResult,
    handleSave,
    handleTestConnection,
    isSaveDisabled,
    isSaving: updateMutation.isPending,
    isTesting: testMutation.isPending,
    isConfigured: config?.status === 'configured',
  };
}
