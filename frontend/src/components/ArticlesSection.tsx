import React, { useState } from "react";
import { useArticles, useDeleteArticle } from "../hooks/useArticles";
import { useArticleChunks } from "../hooks/useChunks";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "./ui/card";
import { Button } from "./ui/button";

export const ArticlesSection: React.FC = () => {
    const [selectedArticleId, setSelectedArticleId] = useState<string | null>(null);
    const { data: articlesData, isLoading, error } = useArticles(1, 10);
    const deleteArticleMutation = useDeleteArticle();
    const articles = articlesData?.articles || [];
    
    const { data: chunks, isLoading: chunksLoading, error: chunksError } = useArticleChunks(selectedArticleId || '');
    
    const handleLoadChunks = (articleId: string) => {
        setSelectedArticleId(selectedArticleId === articleId ? null : articleId);
    };

    const refreshPage = () => {
    window.location.reload();
};

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <h2 className="text-ios-title font-bold">Articles</h2>
                <Button variant="outline" size="sm" onClick={() => refreshPage()}>
                    Refresh
                </Button>
            </div>

            <Card>
                <CardHeader>
                    <CardTitle>Article Management</CardTitle>
                    <CardDescription>
                        Browse and manage your uploaded articles
                    </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    {isLoading && <div>Loading articles from backend...</div>}
                    {error && <div>Error: {error?.message}</div>}

                    <div className="grid gap-4">
                        {articles.map(article => (
                            <Card key={article.id} className="p-4">
                                <div className="flex items-start justify-between">
                                    <div className="space-y-1">
                                        <h3 className="font-medium">{article.title}</h3>
                                        <p className="text-sm text-muted-foreground">{article.category}</p>
                                    </div>
                                    <div className="flex space-x-2">
                                        <Button
                                            size="sm"
                                            variant="outline"
                                            onClick={() => handleLoadChunks(article.id)}
                                            disabled={chunksLoading && selectedArticleId === article.id}
                                        >
                                            {chunksLoading && selectedArticleId === article.id 
                                                ? 'Loading...' 
                                                : selectedArticleId === article.id 
                                                    ? 'Hide Chunks' 
                                                    : 'Load Chunks'
                                            }
                                        </Button>
                                        <Button
                                            size="sm"
                                            variant="destructive"
                                            onClick={() => deleteArticleMutation.mutate(article.id)}
                                            disabled={deleteArticleMutation.isPending}
                                        >
                                            {deleteArticleMutation.isPending ? 'Deleting...' : 'Delete'}
                                        </Button>
                                    </div>
                                </div>
                                
                                {/* Display chunks if this article is selected */}
                                {selectedArticleId === article.id && (
                                    <div className="mt-4 pt-4 border-t border-border">

                                        <h4 className="font-medium mb-3">Article Chunks</h4>
                                        {chunksLoading && <div className="text-sm text-muted-foreground">Loading chunks...</div>}
                                        {chunksError && <div className="text-sm text-red-500">Error loading chunks: {chunksError.message}</div>}

                                        {chunks && chunks.length > 0 && (
                                            <div className="space-y-3">
                                                {chunks.map((chunk, _) => (
                                                    <div key={chunk.id} className="bg-muted/30 rounded-md p-3">
                                                        <div className="flex items-center justify-between mb-2">
                                                            <span className="text-xs font-medium text-muted-foreground">
                                                                Chunk {chunk.chunk_index + 1}
                                                            </span>
                                                            <div className="flex items-center space-x-2 text-xs text-muted-foreground">
                                                                <span>{chunk.character_count} chars</span>
                                                                <span>{chunk.token_count} tokens</span>
                                                            </div>
                                                        </div>
                                                        <p className="text-sm text-foreground leading-relaxed">
                                                            {chunk.content.length > 300 
                                                                ? `${chunk.content.substring(0, 300)}...` 
                                                                : chunk.content
                                                            }
                                                        </p>
                                                    </div>
                                                ))}
                                            </div>
                                        )}
                                        {chunks && chunks.length === 0 && (
                                            <div className="text-sm text-muted-foreground">No chunks found for this article.</div>
                                        )}
                                    </div>
                                )}
                            </Card>
                        ))}
                    </div>
                </CardContent>
            </Card>
        </div>
    );
};