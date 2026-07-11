import { useEffect, useState } from 'react';
import { createPortal } from 'react-dom';

export const CANVAS_COMMANDS_SLOT_ID = 'canvas-commands-slot';

export function CanvasCommandsPortal({ children }: { children: React.ReactNode }) {
  const [slot, setSlot] = useState<HTMLElement | null>(null);

  useEffect(() => {
    setSlot(document.getElementById(CANVAS_COMMANDS_SLOT_ID));
  }, []);

  if (!slot) return <>{children}</>;
  return createPortal(children, slot);
}
