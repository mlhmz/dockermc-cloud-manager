import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../api/client';
import { serversQueryKey } from '../queries/useServers';
import type { CreateServerRequest } from '@/schemas';

export function useCreateServer() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: CreateServerRequest) => apiClient.createServer(request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: serversQueryKey });
    },
  });
}
