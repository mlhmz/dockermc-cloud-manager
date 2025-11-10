import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../api/client';
import { serversQueryKey } from '../queries/useServers';

export function useDeleteServer() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => apiClient.deleteServer(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: serversQueryKey });
    },
  });
}
