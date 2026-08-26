import { NumberInput, Paper, SimpleGrid, Stack, Textarea } from '@mantine/core';

interface AdvancedSettingsProps {
  maxTokens: number;
  onMaxTokensChange: (value: number) => void;
  temperature: number;
  onTemperatureChange: (value: number) => void;
  systemPromptOverride: string;
  onSystemPromptOverrideChange: (value: string) => void;
}

export function AdvancedSettings(props: AdvancedSettingsProps) {
  return (
    <Paper bg="gray.0" p="md" radius="sm">
      <Stack gap="md">
        <SimpleGrid cols={{ base: 1, sm: 2 }} spacing="md">
          <NumberInput
            id="ai-max-tokens"
            label="Max Tokens"
            min={256}
            max={32768}
            value={props.maxTokens}
            onChange={(v) => props.onMaxTokensChange(typeof v === 'number' ? v : 4096)}
            description="256 - 32,768"
          />
          <NumberInput
            id="ai-temperature"
            label="Temperature"
            min={0}
            max={2}
            step={0.1}
            decimalScale={1}
            value={props.temperature}
            onChange={(v) => props.onTemperatureChange(typeof v === 'number' ? v : 0.3)}
            description="0.0 - 2.0 (lower = more deterministic)"
          />
        </SimpleGrid>
        <Textarea
          id="ai-system-prompt"
          label="System Prompt Override"
          value={props.systemPromptOverride}
          onChange={(e) => props.onSystemPromptOverrideChange(e.currentTarget.value)}
          placeholder="Provide additional organizational context..."
          maxLength={2000}
          description="Provide additional organizational context. This is appended to the built-in system prompt as informational context."
          autosize
          minRows={3}
        />
      </Stack>
    </Paper>
  );
}
