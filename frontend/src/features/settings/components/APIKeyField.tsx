import { Button, Group, Input, PasswordInput, Stack } from '@mantine/core';
import classes from './APIKeyField.module.css';

interface APIKeyFieldProps {
  apiKeyStatus: string | undefined;
  apiKey: string;
  onApiKeyChange: (value: string) => void;
  showInput: boolean;
  onShowInput: (show: boolean) => void;
}

export function APIKeyField({ apiKeyStatus, apiKey, onApiKeyChange, showInput, onShowInput }: APIKeyFieldProps) {
  const isConfigured = apiKeyStatus === 'configured';

  return (
    <Stack gap="xs">
      <Input.Label htmlFor="ai-api-key" required>
        API Key
      </Input.Label>
      {isConfigured && !showInput ? (
        <Group gap="xs">
          <span className={classes.configuredStatus}>API key configured</span>
          <Button variant="subtle" size="xs" onClick={() => onShowInput(true)}>
            Change
          </Button>
        </Group>
      ) : (
        <Group gap="sm" align="flex-start" wrap="nowrap">
          <PasswordInput
            id="ai-api-key"
            value={apiKey}
            onChange={(e) => onApiKeyChange(e.currentTarget.value)}
            placeholder="sk-..."
            flex={1}
          />
          {isConfigured && (
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
        </Group>
      )}
    </Stack>
  );
}
