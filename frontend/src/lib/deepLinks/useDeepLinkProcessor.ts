import { useEffect, useRef } from 'react';
import { clearParams, getParamValue } from './registry';
import type { DeepLinkHandler } from './types';

export function useDeepLinkProcessor(handlers: DeepLinkHandler[], isReady: boolean = true): void {
  const processedRef = useRef(false);

  useEffect(() => {
    if (!isReady || processedRef.current) return;

    const paramsToProcess = handlers
      .map((handler) => ({ handler, value: getParamValue(handler.param) }))
      .filter(({ value }) => value !== null);

    if (paramsToProcess.length === 0) return;

    processedRef.current = true;

    paramsToProcess.forEach(({ handler, value }) => {
      handler.onFound(value!);
    });

    clearParams(paramsToProcess.map(({ handler }) => handler.param));
  }, [handlers, isReady]);
}
