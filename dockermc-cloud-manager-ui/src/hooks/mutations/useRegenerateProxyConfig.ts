import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../api/client';
import { proxyQueryKey } from '../queries/useProxy';

export function useRegenerateProxyConfig() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => apiClient.regenerateProxyConfig(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: proxyQueryKey });
    },
  });
}
