import React, { useState } from 'react';
import { Button } from './ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card';
import { Input } from './ui/input';
import { useSearchImmediateMutation, useSearchAsync, useSearchJobStatus } from '../hooks/useSearch';

export const SearchSection: React.FC = () => {
    const [query, setQuery] = useState('');
    const [topK, setTopK] = useState(10);
    const [scoreThreshold, setScoreThreshold] = useState(0.7);
    const [searchMode, setSearchMode] = useState<'immediate' | 'async'>('immediate');
    const [jobId, setJobId] = useState<string | null>(null);

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
    const immediateResults = searchImmediateMutation.data;
    const asyncResults = jobStatusQuery.data?.status === 'completed' ? jobStatusQuery.data.results : null;
    const searchResults = immediateResults || asyncResults;

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

                        {/* Job Status for Async Search */}
                        {jobId && searchMode === 'async' && (
                            <div className="bg-muted/30 rounded-md p-3">
                                <div className="flex items-center space-x-2 mb-2">
                                    <div className="h-2 w-2 bg-blue-500 rounded-full animate-pulse" />
                                    <span className="text-sm font-medium">Job Status</span>
                                </div>
                                <p className="text-xs text-muted-foreground">Job ID: {jobId}</p>
                                <p className="text-xs text-muted-foreground">
                                    Status: {jobStatusQuery.data?.status || 'pending'}
                                </p>
                                {jobStatusQuery.data?.status === 'completed' && (
                                    <p className="text-xs text-accent-green">Search completed!</p>
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
                            `Found ${searchResults.total_results} results in ${searchResults.search_time_ms}ms` :
                            'Results will appear here with similarity scores'
                        }
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    {searchResults?.results && searchResults.results.length > 0 ? (
                        <div className="space-y-4">
                            {searchResults.results.map((result, index) => (
                                <div key={`${result.article_id}-${index}`} className="border border-border rounded-md p-4">
                                    <div className="flex items-start justify-between mb-2">
                                        <h3 className="font-medium text-sm">{result.title}</h3>
                                        <div className="flex items-center space-x-2">
                                            <span className="text-xs bg-muted px-2 py-1 rounded">
                                                {result.category}
                                            </span>
                                            <span className="text-xs bg-accent-green/10 text-accent-green px-2 py-1 rounded">
                                                {(result.similarity_score * 100).toFixed(1)}% match
                                            </span>
                                        </div>
                                    </div>
                                    <p className="text-sm text-muted-foreground mb-2">
                                        {result.content.length > 200 
                                            ? `${result.content.substring(0, 200)}...` 
                                            : result.content
                                        }
                                    </p>
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
                                jobId
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
                    {jobId && (
                        <div>
                            <h4 className="font-medium mb-2">Job Polling Details</h4>
                            <pre className="text-xs bg-muted p-3 rounded overflow-auto max-h-32">
                                {JSON.stringify({
                                    jobId,
                                    status: jobStatusQuery.data?.status || 'pending',
                                    isPolling: jobStatusQuery.isFetching,
                                    error: jobStatusQuery.error?.message
                                }, null, 2)}
                            </pre>
                        </div>
                    )}
                </CardContent>
            </Card>
        </div>
    );
};