import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';

// Types for API responses
export interface Article {
  id: string;
  title: string;
  content: string;
  url: string;
  category: string;
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
  success: boolean;
  data: {
    article: Article;
    processing_mode: string;
  };
  message: string;
}

export interface ArticleChunk {
  id: string;
  article_id: string;
  content: string;
  chunk_index: number;
  character_count: number;
  token_count: number;
  created_at: string;
  pinecone_id?: string | null;
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
  service: string;
  status: string;
  timestamp: string;
  uptime: string;
  version: string;
  dependencies: {
    database: {
      status: string;
      response_time: string;
    };
    redis: {
      status: string;
      response_time: string;
    };
    rabbitmq: {
      status: string;
    };
    python_ml_service: {
      status: string;
      response_time: string;
      response: {
        status: string;
        timestamp: string;
        version: string;
      };
    };
    query_rag_worker: {
      status: string;
      response_time: string;
    };
  };
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
    // Transform Go backend response to match expected interface
    return {
      articles: response.data.data || [],
      total: response.data.total || 0
    };
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
    const response = await this.client.get(`/api/v1/articles/${articleId}/chunks`);
    // Handle potential wrapped response format
    if (response.data.success && response.data.data) {
      return response.data.data.chunks || response.data.data || [];
    }
    return response.data || [];
  }

  async getAllChunks(page: number = 1, limit: number = 10): Promise<{ chunks: ArticleChunk[], total: number }> {
    const response = await this.client.get(`/api/v1/chunks?page=${page}&limit=${limit}`);
    // Transform Go backend response to match expected interface
    if (response.data.success && response.data.data) {
      return {
        chunks: response.data.data.chunks || [],
        total: response.data.data.total || 0
      };
    }
    return { chunks: [], total: 0 };
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