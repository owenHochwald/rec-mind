import { useQuery, useMutation } from '@tanstack/react-query';
import { apiClient, SearchResponse } from '../services/api';

// Query Keys
export const searchKeys = {
  all: ['search'] as const,
  results: (query: string, topK: number, scoreThreshold: number) => 
    [...searchKeys.all, 'results', { query, topK, scoreThreshold }] as const,
  job: (jobId: string) => [...searchKeys.all, 'job', jobId] as const,
};

// Queries
export function useSearchImmediate(query: string, topK: number = 10, scoreThreshold: number = 0.7) {
  return useQuery({
    queryKey: searchKeys.results(query, topK, scoreThreshold),
    queryFn: () => apiClient.searchImmediate(query, topK, scoreThreshold),
    enabled: false, // Manual triggering
    staleTime: 2 * 60 * 1000, // 2 minutes - search results stay fresh longer
  });
}

export function useSearchJobStatus(jobId: string) {
  return useQuery({
    queryKey: searchKeys.job(jobId),
    queryFn: () => apiClient.getSearchJobStatus(jobId),
    enabled: !!jobId,
    refetchInterval: (query) => {
      // Stop polling if job is completed or failed
      if (query.state.data?.status === 'completed' || query.state.data?.status === 'failed') {
        return false;
      }
      return 2000; // Poll every 2 seconds
    },
    refetchIntervalInBackground: true,
  });
}

// Mutations
export function useSearchAsync() {
  return useMutation({
    mutationFn: ({ query, topK, scoreThreshold }: { 
      query: string; 
      topK: number; 
      scoreThreshold: number; 
    }) => apiClient.searchAsync(query, topK, scoreThreshold),
  });
}

export function useSearchImmediateMutation() {
  return useMutation({
    mutationFn: ({ query, topK, scoreThreshold }: { 
      query: string; 
      topK: number; 
      scoreThreshold: number; 
    }) => apiClient.searchImmediate(query, topK, scoreThreshold),
  });
}

// Search health query
export function useSearchHealth() {
  return useQuery({
    queryKey: [...searchKeys.all, 'health'],
    queryFn: () => apiClient.getSearchHealth(),
    refetchInterval: 30 * 1000, // Check every 30 seconds
    staleTime: 15 * 1000, // Consider stale after 15 seconds
  });
}

// TODO: Practice implementing these hooks in your components:
//
// Immediate search example:
// const searchQuery = useSearchImmediate('AI technology', 10, 0.7);
// const handleSearch = () => searchQuery.refetch();
//
// Async search example:
// const searchAsyncMutation = useSearchAsync();
// const [jobId, setJobId] = useState<string | null>(null);
// const jobStatusQuery = useSearchJobStatus(jobId || '');
//
// const handleAsyncSearch = async () => {
//   try {
//     const result = await searchAsyncMutation.mutateAsync({
//       query: 'AI technology',
//       topK: 10,
//       scoreThreshold: 0.7
//     });
//     setJobId(result.job_id);
//   } catch (error) {
//     console.error('Search failed:', error);
//   }
// };