import { describe, it, expect } from 'vitest';
import { parseSSEChunk } from './parseSSE';

describe('parseSSEChunk', () => {
  it('should parse a token event', () => {
    const chunk = 'event: token\ndata: {"content":"Hello"}\n\n';
    const events = parseSSEChunk(chunk);
    expect(events).toEqual([{ type: 'token', content: 'Hello' }]);
  });

  it('should parse a done event', () => {
    const chunk = 'event: done\ndata: {"messageId":"msg-1","tokensUsed":42}\n\n';
    const events = parseSSEChunk(chunk);
    expect(events).toEqual([{ type: 'done', messageId: 'msg-1', tokensUsed: 42 }]);
  });

  it('should parse an error event', () => {
    const chunk = 'event: error\ndata: {"code":"llm_error","message":"Service unavailable"}\n\n';
    const events = parseSSEChunk(chunk);
    expect(events).toEqual([{ type: 'error', code: 'llm_error', message: 'Service unavailable' }]);
  });

  it('should ignore ping events', () => {
    const chunk = 'event: ping\ndata: {}\n\n';
    const events = parseSSEChunk(chunk);
    expect(events).toEqual([]);
  });

  it('should parse multiple events in one chunk', () => {
    const chunk = 'event: token\ndata: {"content":"Hello "}\n\nevent: token\ndata: {"content":"world"}\n\n';
    const events = parseSSEChunk(chunk);
    expect(events).toEqual([
      { type: 'token', content: 'Hello ' },
      { type: 'token', content: 'world' },
    ]);
  });

  it('should handle empty chunk', () => {
    const events = parseSSEChunk('');
    expect(events).toEqual([]);
  });

  it('should handle chunk with only whitespace', () => {
    const events = parseSSEChunk('\n\n');
    expect(events).toEqual([]);
  });

  it('should handle partial chunks without double newline', () => {
    const chunk = 'event: token\ndata: {"content":"partial"}';
    const events = parseSSEChunk(chunk);
    expect(events).toEqual([]);
  });
});
