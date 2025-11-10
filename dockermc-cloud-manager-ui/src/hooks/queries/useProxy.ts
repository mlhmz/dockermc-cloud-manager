import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../../api/client';

export const proxyQueryKey = ['proxy'] as const;

export function useProxy() {
  return useQuery({
    queryKey: proxyQueryKey,
    queryFn: () => apiClient.getProxy(),
    refetchInterval: 5000,
    retry: false, // Don't retry if proxy doesn't exist yet
  });
}
