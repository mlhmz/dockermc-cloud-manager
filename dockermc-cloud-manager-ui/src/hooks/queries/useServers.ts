import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../../api/client';

export const serversQueryKey = ['servers'] as const;

export function useServers() {
  return useQuery({
    queryKey: serversQueryKey,
    queryFn: () => apiClient.listServers(),
    refetchInterval: 5000, // Refetch every 5 seconds
  });
}
