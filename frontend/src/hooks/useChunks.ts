import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../services/api';

// Query Keys
export const chunkKeys = {
  all: ['chunks'] as const,
  lists: () => [...chunkKeys.all, 'list'] as const,
  list: (page?: number, limit?: number) => [...chunkKeys.lists(), { page, limit }] as const,
  details: () => [...chunkKeys.all, 'detail'] as const,
  detail: (id: string) => [...chunkKeys.details(), id] as const,
  articleChunks: (articleId: string) => [...chunkKeys.all, 'article', articleId] as const,
};

// Queries
export function useChunks(page: number = 1, limit: number = 10) {
  return useQuery({
    queryKey: chunkKeys.list(page, limit),
    queryFn: () => apiClient.getAllChunks(page, limit),
  });
}

export function useChunk(id: string) {
  return useQuery({
    queryKey: chunkKeys.detail(id),
    queryFn: () => apiClient.getChunk(id),
    enabled: !!id,
  });
}

export function useArticleChunks(articleId: string) {
  return useQuery({
    queryKey: chunkKeys.articleChunks(articleId),
    queryFn: () => apiClient.getArticleChunks(articleId),
    enabled: !!articleId,
  });
}

// Mutations
export function useCreateChunk() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (chunk: { article_id: string; content: string; chunk_index: number }) => 
      apiClient.createChunk(chunk),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: chunkKeys.all });
    },
  });
}

export function useCreateChunksBatch() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (chunks: { article_id: string; content: string; chunk_index: number }[]) => 
      apiClient.createChunksBatch(chunks),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: chunkKeys.all });
    },
  });
}

export function useDeleteChunk() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (id: string) => apiClient.deleteChunk(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: chunkKeys.all });
    },
  });
}

export function useDeleteArticleChunks() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (articleId: string) => apiClient.deleteArticleChunks(articleId),
    onSuccess: (_, articleId) => {
      queryClient.invalidateQueries({ queryKey: chunkKeys.articleChunks(articleId) });
      queryClient.invalidateQueries({ queryKey: chunkKeys.all });
    },
  });
}
