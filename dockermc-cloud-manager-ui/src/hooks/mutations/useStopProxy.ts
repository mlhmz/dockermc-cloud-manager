import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../api/client';
import { proxyQueryKey } from '../queries/useProxy';

export function useStopProxy() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => apiClient.stopProxy(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: proxyQueryKey });
    },
  });
}
