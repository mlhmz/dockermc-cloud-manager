import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../../api/client';

export const healthQueryKey = ['health'] as const;

export function useHealth() {
  return useQuery({
    queryKey: healthQueryKey,
    queryFn: () => apiClient.getHealth(),
    refetchInterval: 30000, // Refetch every 30 seconds
  });
}
