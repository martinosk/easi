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
    <div className="ai-config-advanced-section">
      <div className="ai-config-row">
        <div className="ai-config-field">
          <label htmlFor="ai-max-tokens">Max Tokens</label>
          <input
            id="ai-max-tokens"
            type="number"
            min={256}
            max={32768}
            value={props.maxTokens}
            onChange={(e) => props.onMaxTokensChange(parseInt(e.target.value) || 4096)}
          />
          <p className="ai-config-field-hint">256 - 32,768</p>
        </div>

        <div className="ai-config-field">
          <label htmlFor="ai-temperature">Temperature</label>
          <input
            id="ai-temperature"
            type="number"
            min={0}
            max={2}
            step={0.1}
            value={props.temperature}
            onChange={(e) => props.onTemperatureChange(parseFloat(e.target.value) || 0.3)}
          />
          <p className="ai-config-field-hint">0.0 - 2.0 (lower = more deterministic)</p>
        </div>
      </div>

      <div className="ai-config-field">
        <label htmlFor="ai-system-prompt">System Prompt Override</label>
        <textarea
          id="ai-system-prompt"
          value={props.systemPromptOverride}
          onChange={(e) => props.onSystemPromptOverrideChange(e.target.value)}
          placeholder="Provide additional organizational context..."
          maxLength={2000}
        />
        <p className="ai-config-field-hint">
          Provide additional organizational context. This is appended to the built-in system prompt as informational context.
        </p>
      </div>
    </div>
  );
}
