import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../../api/client';

export const serverQueryKey = (id: string) => ['servers', id] as const;

export function useServer(id: string) {
  return useQuery({
    queryKey: serverQueryKey(id),
    queryFn: () => apiClient.getServer(id),
    enabled: !!id,
  });
}
