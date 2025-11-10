import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../api/client';
import { serversQueryKey } from '../queries/useServers';
import { serverQueryKey } from '../queries/useServer';

export function useStartServer() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => apiClient.startServer(id),
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: serversQueryKey });
      queryClient.invalidateQueries({ queryKey: serverQueryKey(id) });
    },
  });
}
