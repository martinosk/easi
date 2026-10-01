import { ErrorScreen } from '../../components/shared/ErrorScreen';
import { LoadingScreen } from '../../components/shared/LoadingScreen';
import { useAppInitialization } from '../../hooks/useAppInitialization';
import CanvasContainer from './CanvasContainer';

export default function CanvasView() {
  const { isLoading, error } = useAppInitialization();

  if (isLoading) return <LoadingScreen fill="region" />;
  if (error) return <ErrorScreen fill="region" error={error.message} onRetry={() => window.location.reload()} />;
  return <CanvasContainer />;
}
