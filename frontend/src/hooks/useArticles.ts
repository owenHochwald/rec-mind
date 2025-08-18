import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient, Article } from '../services/api';

// Query Keys
export const articleKeys = {
  all: ['articles'] as const,
  lists: () => [...articleKeys.all, 'list'] as const,
  list: (page?: number, limit?: number) => [...articleKeys.lists(), { page, limit }] as const,
  details: () => [...articleKeys.all, 'detail'] as const,
  detail: (id: string) => [...articleKeys.details(), id] as const,
  chunks: (id: string) => [...articleKeys.all, 'chunks', id] as const,
};

// Queries
export function useArticles(page: number = 1, limit: number = 10) {
  return useQuery({
    queryKey: articleKeys.list(page, limit),
    queryFn: () => apiClient.getArticles(page, limit),
    enabled: true, // Always fetch articles
  });
}

export function useArticle(id: string) {
  return useQuery({
    queryKey: articleKeys.detail(id),
    queryFn: () => apiClient.getArticle(id),
    enabled: !!id, // Only fetch if ID exists
  });  
}

export function useArticleChunks(articleId: string) {
  return useQuery({
    queryKey: articleKeys.chunks(articleId),
    queryFn: () => apiClient.getArticleChunks(articleId),
    enabled: !!articleId, 
  });
}

// Mutations
export function useDeleteArticle() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (id: string) => apiClient.deleteArticle(id),
    onSuccess: () => {
      // Invalidate and refetch articles list
      queryClient.invalidateQueries({ queryKey: articleKeys.lists() });
    },
  });
}

export function useUploadArticle() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: ({ article, isSync }: { article: any; isSync: boolean }) => 
      apiClient.uploadArticle(article, isSync),
    onSuccess: () => {
      // Invalidate and refetch articles list
      queryClient.invalidateQueries({ queryKey: articleKeys.lists() });
    },
  });
}