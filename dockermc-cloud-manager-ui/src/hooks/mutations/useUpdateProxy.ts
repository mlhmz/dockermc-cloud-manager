import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../api/client';
import { proxyQueryKey } from '../queries/useProxy';
import type { UpdateProxyRequest } from '@/schemas';

export function useUpdateProxy() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: UpdateProxyRequest) => apiClient.updateProxy(request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: proxyQueryKey });
    },
  });
}
