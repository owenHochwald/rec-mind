import React, { useState } from 'react';
import { Button } from './ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card';
import { Input } from './ui/input';
import { useSearchImmediateMutation, useSearchAsync, useSearchJobStatus } from '../hooks/useSearch';
import { useArticles } from '../hooks/useArticles';
import { SearchResult } from '../services/api';

export const SearchSection: React.FC = () => {
    const [query, setQuery] = useState('');
    const [topK, setTopK] = useState(10);
    const [scoreThreshold, setScoreThreshold] = useState(0.5);
    const [searchMode, setSearchMode] = useState<'immediate' | 'async'>('immediate');
    const [jobId, setJobId] = useState<string | null>(null);
    const [selectedArticle, setSelectedArticle] = useState<SearchResult | null>(null);

    const searchImmediateMutation = useSearchImmediateMutation();
    const searchAsyncMutation = useSearchAsync();
    const jobStatusQuery = useSearchJobStatus(jobId || '');

    const handleSearch = async () => {
        if (!query.trim()) return;

        try {
            if (searchMode === 'immediate') {
                searchImmediateMutation.mutate({
                    query,
                    topK,
                    scoreThreshold
                });
            } else {
                const result = await searchAsyncMutation.mutateAsync({
                    query,
                    topK,
                    scoreThreshold
                });
                setJobId(result.job_id);
            }
        } catch (error) {
            console.error('Search failed:', error);
        }
    };

    const isSearching = searchImmediateMutation.isPending || searchAsyncMutation.isPending;
    // Both immediate and async return job_id, so we always poll for results
    const immediateJobId = searchImmediateMutation.data?.job_id;
    const asyncJobId = searchAsyncMutation.data?.job_id;
    const activeJobId = immediateJobId || asyncJobId || jobId;
    
    // Get results from job polling (works for both immediate and async)
    const pollingQuery = useSearchJobStatus(activeJobId || '');
    const searchResults = pollingQuery.data?.status === 'completed' ? pollingQuery.data.results : null;

    // Helper function to format search time
    const formatSearchTime = (timeMs: number): string => {
        if (timeMs >= 1000) {
            const seconds = timeMs / 1000;
            return `${seconds.toFixed(2)}s`;
        } else {
            return `${Math.round(timeMs)}ms`;
        }
    };

    // Get full article data for popup
    const { data: articlesData } = useArticles(1, 100); // Get more articles for popup lookup
    const articles = articlesData?.articles || [];

    const getFullArticle = (articleId: string) => {
        return articles.find(article => article.id === articleId);
    };

    // Function to highlight matched chunks in article content
    const highlightMatchedChunks = (content: string, matchedChunks: SearchResult['chunk_matches']) => {
        if (!matchedChunks || matchedChunks.length === 0) return content;
        
        let highlightedContent = content;
        
        // Sort chunks by score (highest first) to highlight best matches first
        const sortedChunks = [...matchedChunks].sort((a, b) => b.score - a.score);
        
        sortedChunks.forEach((chunk, index) => {
            // Use chunk preview content to find and highlight in full content
            if (chunk.content_preview) {
                // Remove ellipsis and find the exact text in content
                const previewText = chunk.content_preview.replace(/\.\.\./g, '').trim();
                if (previewText.length > 10) {
                    const regex = new RegExp(previewText.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'gi');
                    highlightedContent = highlightedContent.replace(regex, 
                        `<mark class="bg-accent-green/30 px-1 rounded" data-score="${(chunk.score * 100).toFixed(1)}%">$&</mark>`
                    );
                }
            }
        });
        
        return highlightedContent;
    };

    return (
        <div className="space-y-6">
            <div>
                <h2 className="text-ios-title font-bold">Search Recommendations</h2>
                <p className="text-muted-foreground mt-2">Find similar articles using semantic search</p>
            </div>

            <Card>
                <CardHeader>
                    <CardTitle>Search Query</CardTitle>
                    <CardDescription>
                        Enter your search query and configure search parameters
                    </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="space-y-4">
                        <div>
                            <label className="text-sm font-medium">Search Query</label>
                            <Input 
                                placeholder="What are you looking for?"
                                value={query}
                                onChange={(e) => setQuery(e.target.value)}
                                onKeyPress={(e) => e.key === 'Enter' && handleSearch()}
                            />
                        </div>

                        <div className="grid grid-cols-2 gap-4">
                            <div>
                                <label className="text-sm font-medium">Top Results</label>
                                <Input 
                                    type="number" 
                                    value={topK} 
                                    onChange={(e) => setTopK(Number(e.target.value))}
                                    min="1" 
                                    max="50" 
                                />
                            </div>
                            <div>
                                <label className="text-sm font-medium">Score Threshold</label>
                                <Input 
                                    type="number" 
                                    step="0.1" 
                                    value={scoreThreshold}
                                    onChange={(e) => setScoreThreshold(Number(e.target.value))}
                                    min="0" 
                                    max="1" 
                                />
                            </div>
                        </div>

                        <div className="flex space-x-4">
                            <label className="flex items-center space-x-2">
                                <input 
                                    type="radio" 
                                    name="searchMode" 
                                    value="immediate" 
                                    checked={searchMode === 'immediate'}
                                    onChange={(e) => setSearchMode('immediate')}
                                />
                                <span className="text-sm">Immediate</span>
                            </label>
                            <label className="flex items-center space-x-2">
                                <input 
                                    type="radio" 
                                    name="searchMode" 
                                    value="async" 
                                    checked={searchMode === 'async'}
                                    onChange={(e) => setSearchMode('async')}
                                />
                                <span className="text-sm">Async (with polling)</span>
                            </label>
                        </div>

                        <Button 
                            className="w-full"
                            onClick={handleSearch}
                            disabled={isSearching || !query.trim()}
                        >
                            {isSearching ? 'Searching...' : 'Search'}
                        </Button>

                        {/* Job Status for Active Search */}
                        {activeJobId && (
                            <div className="bg-muted/30 rounded-md p-3">
                                <div className="flex items-center space-x-2 mb-2">
                                    <div className="h-2 w-2 bg-blue-500 rounded-full animate-pulse" />
                                    <span className="text-sm font-medium">Job Status</span>
                                </div>
                                <p className="text-xs text-muted-foreground">Job ID: {activeJobId}</p>
                                <p className="text-xs text-muted-foreground">
                                    Status: {pollingQuery.data?.status || 'pending'}
                                </p>
                                {pollingQuery.data?.status === 'completed' && (
                                    <p className="text-xs text-accent-green">Search completed!</p>
                                )}
                                {pollingQuery.data?.error && (
                                    <p className="text-xs text-red-500">Error: {pollingQuery.data.error}</p>
                                )}
                            </div>
                        )}
                    </div>

                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>Search Results</CardTitle>
                    <CardDescription>
                        {searchResults?.results ? 
                            `Found ${searchResults.total_results} result${searchResults.total_results !== 1 ? 's' : ''} in ${formatSearchTime(searchResults.search_time_ms)}` :
                            'Results will appear here with similarity scores'
                        }
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    {searchResults?.results && searchResults.results.length > 0 ? (
                        <div className="space-y-4">
                            {searchResults.results.map((result, index) => (
                                <div key={`${result.article_id}-${index}`} className="border border-border rounded-md p-4 hover:border-accent-green/50 cursor-pointer transition-colors" onClick={() => setSelectedArticle(result)}>
                                    <div className="flex items-start justify-between mb-2">
                                        <h3 className="font-medium text-sm hover:text-accent-green">{result.title}</h3>
                                        <div className="flex items-center space-x-2">
                                            <span className="text-xs bg-muted px-2 py-1 rounded">
                                                {result.category}
                                            </span>
                                            <span className="text-xs bg-accent-green/10 text-accent-green px-2 py-1 rounded">
                                                {(result.hybrid_score * 100).toFixed(1)}% match
                                            </span>
                                        </div>
                                    </div>
                                    <div className="text-sm text-muted-foreground mb-2">
                                        {result.chunk_matches && result.chunk_matches.length > 0 ? (
                                            <div className="space-y-2">
                                                {result.chunk_matches.slice(0, 2).map((chunk, idx) => (
                                                    <div key={chunk.chunk_id} className="bg-muted/20 p-2 rounded text-xs">
                                                        <div className="flex items-center justify-between mb-1">
                                                            <span className="font-medium">Chunk {chunk.chunk_index + 1}</span>
                                                            <span className="text-accent-green">{(chunk.score * 100).toFixed(1)}%</span>
                                                        </div>
                                                        <p>{chunk.content_preview}</p>
                                                    </div>
                                                ))}
                                                {result.chunk_matches.length > 2 && (
                                                    <p className="text-xs text-muted-foreground">
                                                        +{result.chunk_matches.length - 2} more chunks matched
                                                    </p>
                                                )}
                                            </div>
                                        ) : (
                                            <p>No content preview available</p>
                                        )}
                                    </div>
                                    <a 
                                        href={result.url} 
                                        target="_blank" 
                                        rel="noopener noreferrer"
                                        className="text-xs text-blue-500 hover:underline"
                                    >
                                        View Source
                                    </a>
                                </div>
                            ))}
                        </div>
                    ) : searchResults && searchResults.results.length === 0 ? (
                        <div className="text-center py-8">
                            <p className="text-sm text-muted-foreground">No results found for your query.</p>
                            <p className="text-xs text-muted-foreground mt-1">Try adjusting your search terms or lowering the score threshold.</p>
                        </div>
                    ) : (
                        <div className="text-center py-8">
                            <p className="text-sm text-muted-foreground">No search results yet</p>
                            <p className="text-xs text-muted-foreground mt-1">Enter a query above to start searching</p>
                        </div>
                    )}

                    {/* Error Messages */}
                    {searchImmediateMutation.error && (
                        <div className="mt-4 p-3 bg-red-500/5 border border-red-500/20 rounded-md">
                            <p className="text-sm text-red-500">
                                Search failed: {searchImmediateMutation.error.message}
                            </p>
                        </div>
                    )}
                    {searchAsyncMutation.error && (
                        <div className="mt-4 p-3 bg-red-500/5 border border-red-500/20 rounded-md">
                            <p className="text-sm text-red-500">
                                Async search failed: {searchAsyncMutation.error.message}
                            </p>
                        </div>
                    )}
                </CardContent>
            </Card>

            {/* Search Debug Information */}
            <Card>
                <CardHeader>
                    <CardTitle>Search Request & Response Details</CardTitle>
                    <CardDescription>
                        Debug information for search operations with TanStack Query
                    </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    {/* Request Details */}
                    <div>
                        <h4 className="font-medium mb-2">Last Request</h4>
                        <pre className="text-xs bg-muted p-3 rounded overflow-auto max-h-32">
                            {JSON.stringify({
                                query,
                                topK,
                                scoreThreshold,
                                searchMode,
                                activeJobId
                            }, null, 2)}
                        </pre>
                    </div>

                    {/* Response Details */}
                    <div>
                        <h4 className="font-medium mb-2">Response Data</h4>
                        <pre className="text-xs bg-muted p-3 rounded overflow-auto max-h-32">
                            {searchResults ? 
                                JSON.stringify(searchResults, null, 2) : 
                                'No search results yet'
                            }
                        </pre>
                    </div>

                    {/* Mutation Status */}
                    <div>
                        <h4 className="font-medium mb-2">Search Status</h4>
                        <div className="grid grid-cols-2 gap-4 text-sm">
                            <div>
                                <span className="font-medium">Immediate Search:</span>
                                <div className="text-muted-foreground">
                                    Status: {searchImmediateMutation.isPending ? 'Loading' : searchImmediateMutation.error ? 'Error' : searchImmediateMutation.data ? 'Success' : 'Idle'}
                                </div>
                            </div>
                            <div>
                                <span className="font-medium">Async Search:</span>
                                <div className="text-muted-foreground">
                                    Status: {searchAsyncMutation.isPending ? 'Loading' : searchAsyncMutation.error ? 'Error' : jobId ? 'Job Created' : 'Idle'}
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Job Polling Status */}
                    {activeJobId && (
                        <div>
                            <h4 className="font-medium mb-2">Job Polling Details</h4>
                            <pre className="text-xs bg-muted p-3 rounded overflow-auto max-h-32">
                                {JSON.stringify({
                                    activeJobId,
                                    status: pollingQuery.data?.status || 'pending',
                                    isPolling: pollingQuery.isFetching,
                                    error: pollingQuery.error?.message
                                }, null, 2)}
                            </pre>
                        </div>
                    )}
                </CardContent>
            </Card>

            {/* Article Popup Modal */}
            {selectedArticle && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
                    <div className="bg-background border border-border rounded-lg max-w-4xl w-full max-h-[90vh] overflow-hidden">
                        <div className="flex items-center justify-between p-4 border-b border-border">
                            <div>
                                <h2 className="text-lg font-bold">{selectedArticle.title}</h2>
                                <div className="flex items-center space-x-3 mt-1">
                                    <span className="text-xs bg-muted px-2 py-1 rounded">{selectedArticle.category}</span>
                                    <span className="text-xs bg-accent-green/10 text-accent-green px-2 py-1 rounded">
                                        {(selectedArticle.hybrid_score * 100).toFixed(1)}% match
                                    </span>
                                    <span className="text-xs text-muted-foreground">
                                        {selectedArticle.matched_chunks} chunk{selectedArticle.matched_chunks !== 1 ? 's' : ''} matched
                                    </span>
                                </div>
                            </div>
                            <Button 
                                variant="ghost" 
                                size="sm"
                                onClick={() => setSelectedArticle(null)}
                            >
                                ✕
                            </Button>
                        </div>
                        
                        <div className="p-4 overflow-y-auto max-h-[70vh]">
                            {(() => {
                                const fullArticle = getFullArticle(selectedArticle.article_id);
                                if (!fullArticle) {
                                    return (
                                        <div className="text-center py-8">
                                            <p className="text-muted-foreground">Article content not available</p>
                                            <p className="text-xs text-muted-foreground mt-1">
                                                Article ID: {selectedArticle.article_id}
                                            </p>
                                        </div>
                                    );
                                }
                                
                                const highlightedContent = highlightMatchedChunks(fullArticle.content, selectedArticle.chunk_matches);
                                
                                return (
                                    <div className="space-y-4">
                                        {/* Article Metadata */}
                                        <div className="bg-muted/30 p-3 rounded-md text-sm">
                                            <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
                                                <div>
                                                    <span className="font-medium">Published:</span> {new Date(fullArticle.created_at).toLocaleDateString()}
                                                </div>
                                                <div>
                                                    <a 
                                                        href={fullArticle.url} 
                                                        target="_blank" 
                                                        rel="noopener noreferrer"
                                                        className="text-blue-500 hover:underline"
                                                    >
                                                        View Source →
                                                    </a>
                                                </div>
                                            </div>
                                        </div>
                                        
                                        {/* Matched Chunks Summary */}
                                        <div className="bg-accent-green/5 border border-accent-green/20 p-3 rounded-md">
                                            <h4 className="font-medium text-accent-green mb-2">Matched Chunks ({selectedArticle.chunk_matches?.length || 0})</h4>
                                            <div className="grid gap-2">
                                                {selectedArticle.chunk_matches?.map((chunk) => (
                                                    <div key={chunk.chunk_id} className="text-xs bg-background/50 p-2 rounded">
                                                        <div className="flex justify-between items-center mb-1">
                                                            <span>Chunk {chunk.chunk_index + 1}</span>
                                                            <span className="text-accent-green font-medium">{(chunk.score * 100).toFixed(1)}% match</span>
                                                        </div>
                                                        <p className="text-muted-foreground">{chunk.content_preview}</p>
                                                    </div>
                                                ))}
                                            </div>
                                        </div>
                                        
                                        {/* Full Article Content with Highlights */}
                                        <div className="prose prose-sm max-w-none">
                                            <div 
                                                className="text-sm leading-relaxed whitespace-pre-wrap"
                                                dangerouslySetInnerHTML={{ 
                                                    __html: highlightedContent 
                                                }}
                                            />
                                        </div>
                                    </div>
                                );
                            })()}
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};