export type LLMProvider = 'openai' | 'anthropic';

export interface AIConfigurationResponse {
  id: string;
  provider: LLMProvider;
  endpoint: string;
  apiKeyStatus: 'configured' | 'not_configured';
  model: string;
  maxTokens: number;
  temperature: number;
  systemPromptOverride?: string | null;
  status: 'not_configured' | 'configured' | 'error';
  updatedAt: string;
  _links: Record<string, { href: string; method: string }>;
}

export interface UpdateAIConfigRequest {
  provider: LLMProvider;
  endpoint: string;
  apiKey?: string;
  model: string;
  maxTokens: number;
  temperature: number;
  systemPromptOverride?: string | null;
}

export interface TestConnectionResponse {
  success: boolean;
  model?: string;
  latencyMs?: number;
  error?: string;
}
