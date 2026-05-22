import { Button, PasswordInput } from '@mantine/core';

interface APIKeyFieldProps {
  apiKeyStatus: string | undefined;
  apiKey: string;
  onApiKeyChange: (value: string) => void;
  showInput: boolean;
  onShowInput: (show: boolean) => void;
}

export function APIKeyField({ apiKeyStatus, apiKey, onApiKeyChange, showInput, onShowInput }: APIKeyFieldProps) {
  if (apiKeyStatus === 'configured' && !showInput) {
    return (
      <div className="ai-config-field">
        <label htmlFor="ai-api-key">
          API Key <span className="ai-config-required">*</span>
        </label>
        <div>
          <span className="ai-config-api-key-status configured">API key configured</span>{' '}
          <Button variant="subtle" size="xs" onClick={() => onShowInput(true)}>
            Change
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="ai-config-field">
      <label htmlFor="ai-api-key">
        API Key <span className="ai-config-required">*</span>
      </label>
      <div className="ai-config-api-key-row">
        <PasswordInput
          id="ai-api-key"
          value={apiKey}
          onChange={(e) => onApiKeyChange(e.currentTarget.value)}
          placeholder="sk-..."
          style={{ flex: 1 }}
        />
        {apiKeyStatus === 'configured' && (
          <Button
            variant="subtle"
            size="xs"
            onClick={() => {
              onShowInput(false);
              onApiKeyChange('');
            }}
          >
            Cancel
          </Button>
        )}
      </div>
    </div>
  );
}
