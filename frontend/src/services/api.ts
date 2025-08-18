import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';

// Types for API responses
export interface Article {
  id: string;
  title: string;
  content: string;
  url: string;
  category: string;
  published_at: string;
  created_at: string;
  updated_at: string;
}

export interface UploadArticleRequest {
  title: string;
  content: string;
  url: string;
  category: string;
  published_at: string; // ISO date string
}

export interface UploadArticleResponse {
    article: {};
    message: string;
    processing_mode: 'sync' | 'async';
}

export interface ArticleChunk {
  id: string;
  article_id: string;
  content: string;
  chunk_index: number;
  created_at: string;
}

export interface SearchResult {
  article_id: string;
  title: string;
  content: string;
  similarity_score: number;
  url: string;
  category: string;
}

export interface SearchResponse {
  results: SearchResult[];
  total_results: number;
  query: string;
  search_time_ms: number;
}

export interface HealthStatus {
  status: string;
  database: string;
  redis: string;
  rabbitmq: string;
  timestamp: string;
}

// Request timing interface
export interface ApiRequestTiming {
  startTime: number;
  endTime: number;
  duration: number;
  url: string;
  method: string;
}

class ApiClient {
  private client: AxiosInstance;
  public lastRequestTiming: ApiRequestTiming | null = null;

  constructor() {
    this.client = axios.create({
      baseURL: 'http://localhost:8080',
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Add request timing interceptors
    this.client.interceptors.request.use((config) => {
      config.metadata = { startTime: performance.now() };
      return config;
    });

    this.client.interceptors.response.use(
      (response: AxiosResponse) => {
        const endTime = performance.now();
        const startTime = response.config.metadata?.startTime || endTime;
        
        this.lastRequestTiming = {
          startTime,
          endTime,
          duration: endTime - startTime,
          url: response.config.url || '',
          method: response.config.method?.toUpperCase() || 'GET'
        };
        
        return response;
      },
      (error) => {
        // Still capture timing on errors
        if (error.config) {
          const endTime = performance.now();
          const startTime = error.config.metadata?.startTime || endTime;
          
          this.lastRequestTiming = {
            startTime,
            endTime, 
            duration: endTime - startTime,
            url: error.config.url || '',
            method: error.config.method?.toUpperCase() || 'GET'
          };
        }
        return Promise.reject(error);
      }
    );
  }

  // Health check
  async getHealth(): Promise<HealthStatus> {
    const response = await this.client.get<HealthStatus>('/health');
    return response.data;
  }

  // Article endpoints
  async getArticles(page: number = 1, limit: number = 10): Promise<{ articles: Article[], total: number }> {
    const response = await this.client.get(`/api/v1/articles?page=${page}&limit=${limit}`);
    return response.data;
  }

  async getArticle(id: string): Promise<Article> {
    const response = await this.client.get<Article>(`/api/v1/articles/${id}`);
    return response.data;
  }

  async deleteArticle(id: string): Promise<void> {
    await this.client.delete(`/api/v1/articles/${id}`);
  }

  async uploadArticle(article: UploadArticleRequest, isSync: boolean = false): Promise<UploadArticleResponse> {
    const processing = isSync ? 'sync' : 'async';
    const response = await this.client.post<UploadArticleResponse>(
      `/api/upload?processing=${processing}`, 
      article
    );
    
    return response.data;
  }

  // Article chunks endpoints  
  async getArticleChunks(articleId: string): Promise<ArticleChunk[]> {
    const response = await this.client.get<ArticleChunk[]>(`/api/v1/articles/${articleId}/chunks`);
    return response.data;
  }

  async getAllChunks(page: number = 1, limit: number = 10): Promise<{ chunks: ArticleChunk[], total: number }> {
    const response = await this.client.get(`/api/v1/chunks?page=${page}&limit=${limit}`);
    return response.data;
  }

  // Search endpoints
  async searchImmediate(query: string, topK: number = 10, scoreThreshold: number = 0.7): Promise<SearchResponse> {
    const response = await this.client.post<SearchResponse>('/api/v1/search/immediate', {
      query,
      top_k: topK,
      score_threshold: scoreThreshold
    });
    return response.data;
  }

  async searchAsync(query: string, topK: number = 10, scoreThreshold: number = 0.7): Promise<{ job_id: string }> {
    const response = await this.client.post<{ job_id: string }>('/api/v1/search/recommendations', {
      query,
      top_k: topK,
      score_threshold: scoreThreshold
    });
    return response.data;
  }

  async getSearchJobStatus(jobId: string): Promise<{ status: string, results?: SearchResponse }> {
    const response = await this.client.get<{ status: string, results?: SearchResponse }>(`/api/v1/search/jobs/${jobId}`);
    return response.data;
  }

  // Chunk endpoints
  async createChunk(chunk: { article_id: string; content: string; chunk_index: number }): Promise<ArticleChunk> {
    const response = await this.client.post<ArticleChunk>('/api/v1/chunks', chunk);
    return response.data;
  }

  async createChunksBatch(chunks: { article_id: string; content: string; chunk_index: number }[]): Promise<ArticleChunk[]> {
    const response = await this.client.post<ArticleChunk[]>('/api/v1/chunks/batch', { chunks });
    return response.data;
  }

  async getChunk(id: string): Promise<ArticleChunk> {
    const response = await this.client.get<ArticleChunk>(`/api/v1/chunks/${id}`);
    return response.data;
  }

  async deleteChunk(id: string): Promise<void> {
    await this.client.delete(`/api/v1/chunks/${id}`);
  }

  async deleteArticleChunks(articleId: string): Promise<void> {
    await this.client.delete(`/api/v1/articles/${articleId}/chunks`);
  }

  // Search health endpoint
  async getSearchHealth(): Promise<HealthStatus> {
    const response = await this.client.get<HealthStatus>('/api/v1/search/health');
    return response.data;
  }

  // Helper to get last request timing info
  getLastRequestTiming(): ApiRequestTiming | null {
    return this.lastRequestTiming;
  }
}

export const apiClient = new ApiClient();

declare module 'axios' {
  export interface AxiosRequestConfig {
    metadata?: {
      startTime: number;
    };
  }
}