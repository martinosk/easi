import { zodResolver } from '@hookform/resolvers/zod';
import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import type { Component } from '../../../api/types';
import { type ConnectComponentsFormData, connectComponentsSchema } from '../../../lib/schemas';
import { useComponents } from '../../components/hooks/useComponents';
import { canContain, isContainmentKind } from '../utils/connectionKinds';
import { useConnectComponents } from './useConnectComponents';

export interface ConnectionDialogInput {
  isOpen: boolean;
  onClose: () => void;
  initialSource?: string;
  initialTarget?: string;
}

function defaultValues(initialSource?: string, initialTarget?: string): ConnectComponentsFormData {
  return {
    sourceComponentId: initialSource || '',
    targetComponentId: initialTarget || '',
    connectionKind: 'Triggers',
    name: '',
    description: '',
  };
}

function findComponent(components: Component[], id: string): Component | undefined {
  return components.find((c) => c.id === id);
}

function useConnectionForm({ isOpen, initialSource, initialTarget }: ConnectionDialogInput) {
  const form = useForm<ConnectComponentsFormData>({
    resolver: zodResolver(connectComponentsSchema),
    defaultValues: defaultValues(initialSource, initialTarget),
    mode: 'onChange',
  });
  const { reset } = form;

  useEffect(() => {
    if (isOpen) reset(defaultValues(initialSource, initialTarget));
  }, [isOpen, initialSource, initialTarget, reset]);

  return form;
}

function useSubmitWithBackendError(isOpen: boolean, submit: (data: ConnectComponentsFormData) => Promise<unknown>) {
  const [backendError, setBackendError] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen) setBackendError(null);
  }, [isOpen]);

  const onSubmit = async (data: ConnectComponentsFormData) => {
    setBackendError(null);
    try {
      await submit(data);
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to connect components');
    }
  };

  return { backendError, onSubmit };
}

export function useConnectionDialog(input: ConnectionDialogInput) {
  const { data: components = [] } = useComponents();
  const { connect, isPending } = useConnectComponents();
  const form = useConnectionForm(input);

  const sourceComponentId = form.watch('sourceComponentId');
  const targetComponentId = form.watch('targetComponentId');
  const connectionKind = form.watch('connectionKind');
  const source = findComponent(components, sourceComponentId);
  const target = findComponent(components, targetComponentId);
  const containmentAllowed = canContain(source, target);
  const isContainment = isContainmentKind(connectionKind);
  const hasSameSourceAndTarget = sourceComponentId !== '' && sourceComponentId === targetComponentId;

  const { setValue } = form;
  useEffect(() => {
    if (isContainment && !containmentAllowed) setValue('connectionKind', 'Triggers', { shouldValidate: true });
  }, [isContainment, containmentAllowed, setValue]);

  const { backendError, onSubmit } = useSubmitWithBackendError(input.isOpen, async (data) => {
    await connect(data, source);
    input.onClose();
  });

  return {
    form,
    componentOptions: components.map((c) => ({ value: c.id, label: c.name })),
    connectionKind,
    containmentAllowed,
    isContainment,
    hasSameSourceAndTarget,
    isPending,
    backendError,
    onSubmit,
  };
}
