import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../services/api';

export const healthKeys = {
  all: ['health'] as const,
  status: () => [...healthKeys.all, 'status'] as const,
};

// Queries
export function useHealth() {
  return useQuery({
    queryKey: healthKeys.status(),
    queryFn: () => apiClient.getHealth(),
    refetchInterval: 30 * 1000, // Refetch every 30 seconds
    refetchIntervalInBackground: true,
    staleTime: 15 * 1000, // Consider data stale after 15 seconds
  });
}

// TODO: Practice implementing this hook in your components:
//
// Health monitoring example:
// const { data: healthStatus, isLoading, error } = useHealth();
//
// Usage in component:
// if (isLoading) return <div>Checking system health...</div>;
// if (error) return <div>Failed to get health status</div>;
// 
// return (
//   <div>
//     <div>API: {healthStatus?.status}</div>
//     <div>Database: {healthStatus?.database}</div>
//     <div>Redis: {healthStatus?.redis}</div>
//     <div>RabbitMQ: {healthStatus?.rabbitmq}</div>
//   </div>
// );