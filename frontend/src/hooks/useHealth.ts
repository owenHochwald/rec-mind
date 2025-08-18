import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../services/api';

export const healthKeys = {
  all: ['health'] as const,
  status: () => [...healthKeys.all, 'status'] as const,
};

export function useHealth() {
  return useQuery({
    queryKey: healthKeys.status(),
    queryFn: () => apiClient.getHealth(),
    refetchInterval: 30 * 1000, // Refetch every 30 seconds
    refetchIntervalInBackground: true,
    staleTime: 15 * 1000, // Consider data stale after 15 seconds
  });
}
