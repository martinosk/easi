export type SSEEvent =
  | { type: 'token'; content: string }
  | { type: 'done'; messageId: string; tokensUsed: number }
  | { type: 'error'; code: string; message: string };

interface RawSSEBlock {
  eventType: string;
  data: string;
}

function isActionableBlock(eventType: string, data: string): boolean {
  return Boolean(eventType) && Boolean(data) && eventType !== 'ping';
}

function extractFields(block: string): RawSSEBlock | null {
  const trimmed = block.trim();
  if (!trimmed) return null;

  let eventType = '';
  let data = '';

  for (const line of trimmed.split('\n')) {
    if (line.startsWith('event: ')) eventType = line.slice(7);
    else if (line.startsWith('data: ')) data = line.slice(6);
  }

  if (!isActionableBlock(eventType, data)) return null;
  return { eventType, data };
}

function toSSEEvent({ eventType, data }: RawSSEBlock): SSEEvent | null {
  const parsed = JSON.parse(data);
  switch (eventType) {
    case 'token': return { type: 'token', content: parsed.content };
    case 'done': return { type: 'done', messageId: parsed.messageId, tokensUsed: parsed.tokensUsed };
    case 'error': return { type: 'error', code: parsed.code, message: parsed.message };
    default: return null;
  }
}

export function parseSSEChunk(chunk: string): SSEEvent[] {
  if (!chunk.includes('\n\n')) return [];

  const blocks = chunk.split('\n\n');
  const completedBlocks = chunk.endsWith('\n\n') ? blocks : blocks.slice(0, -1);

  const events: SSEEvent[] = [];
  for (const block of completedBlocks) {
    const raw = extractFields(block);
    if (!raw) continue;
    const event = toSSEEvent(raw);
    if (event) events.push(event);
  }
  return events;
}
