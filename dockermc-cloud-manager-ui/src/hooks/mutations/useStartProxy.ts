import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../api/client';
import { proxyQueryKey } from '../queries/useProxy';

export function useStartProxy() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => apiClient.startProxy(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: proxyQueryKey });
    },
  });
}
