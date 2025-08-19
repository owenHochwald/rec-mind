import React, { useState } from 'react';
import { useUploadArticle } from "../hooks/useArticles";
import { Button } from "./ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "./ui/card";
import { Input } from "./ui/input";

export const UploadSection: React.FC = () => {
    const uploadMutation = useUploadArticle();
    const [formData, setFormData] = useState({
        title: '',
        content: '',
        url: '',
        category: '',
        published_at: ''
    });
    const [isSync, setIsSync] = useState(false);
    const [showSuccessMessage, setShowSuccessMessage] = useState(false);

    const handleInputChange = (field: string) => (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
        setFormData(prev => ({
            ...prev,
            [field]: e.target.value
        }));
    };

    const handleSubmit = async () => {
        if (!formData.title || !formData.content || !formData.url || !formData.category || !formData.published_at) {
            return;
        }

        try {
            await uploadMutation.mutateAsync({
                article: formData,
                isSync
            });
            
            // Show success message
            setShowSuccessMessage(true);
            setTimeout(() => setShowSuccessMessage(false), 8000); // Hide after 8 seconds
            
            // Reset form
            setFormData({
                title: '',
                content: '',
                url: '',
                category: '',
                published_at: ''
            });
        } catch (error) {
            console.error('Upload failed:', error);
        }
    };

    const isFormValid = formData.title && formData.content && formData.url && formData.category && formData.published_at;

    return (
        <div className="space-y-6">
            <div>
                <h2 className="text-ios-title font-bold">Upload Article</h2>
                <p className="text-muted-foreground mt-2">Add new articles to your knowledge base</p>
            </div>

            {/* Success Message */}
            {showSuccessMessage && (
                <Card className="border-accent-green/20 bg-accent-green/5">
                    <CardContent className="pt-4">
                        <div className="flex items-center space-x-2">
                            <div className="h-2 w-2 bg-accent-green rounded-full" />
                            <p className="text-sm font-medium text-accent-green">
                                Article uploaded successfully!
                            </p>
                        </div>
                        <p className="text-sm text-muted-foreground mt-1">
                            {isSync 
                                ? "Article processing completed. Embeddings and chunks have been generated."
                                : "Article saved to database. Embeddings and chunks are being generated in the background. Check back in a minute to see the processed chunks."
                            }
                        </p>
                    </CardContent>
                </Card>
            )}

            {/* Upload Error */}
            {uploadMutation.error && (
                <Card className="border-red-500/20 bg-red-500/5">
                    <CardContent className="pt-4">
                        <div className="flex items-center space-x-2">
                            <div className="h-2 w-2 bg-red-500 rounded-full" />
                            <p className="text-sm font-medium text-red-500">
                                Upload failed
                            </p>
                        </div>
                        <p className="text-sm text-muted-foreground mt-1">
                            {uploadMutation.error.message || "An error occurred while uploading the article."}
                        </p>
                    </CardContent>
                </Card>
            )}

            <Card>
                <CardHeader>
                    <CardTitle>Article Details</CardTitle>
                    <CardDescription>
                        Fill in the article information for processing and embedding generation
                    </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="space-y-4">
                        <div>
                            <label className="text-sm font-medium">Title</label>
                            <Input 
                                placeholder="Enter article title..."
                                value={formData.title}
                                onChange={handleInputChange('title')}
                            />
                        </div>
                        
                        <div>
                            <label className="text-sm font-medium">Content</label>
                            <textarea 
                                className="min-h-[200px] w-full rounded-ios border border-input bg-background px-3 py-2 text-sm resize-y"
                                placeholder="Paste article content..."
                                value={formData.content}
                                onChange={handleInputChange('content')}
                            />
                        </div>
                        
                        <div className="grid grid-cols-2 gap-4">
                            <div>
                                <label className="text-sm font-medium">URL</label>
                                <Input 
                                    type="url" 
                                    placeholder="https://..."
                                    value={formData.url}
                                    onChange={handleInputChange('url')}
                                />
                            </div>
                            <div>
                                <label className="text-sm font-medium">Category</label>
                                <Input 
                                    placeholder="e.g., technology"
                                    value={formData.category}
                                    onChange={handleInputChange('category')}
                                />
                            </div>
                        </div>
                        
                        <div>
                            <label className="text-sm font-medium">Published Date</label>
                            <Input 
                                type="datetime-local"
                                value={formData.published_at}
                                onChange={handleInputChange('published_at')}
                            />
                        </div>
                        
                        <div className="flex items-center space-x-2">
                            <input 
                                type="checkbox" 
                                id="sync" 
                                checked={isSync}
                                onChange={(e) => setIsSync(e.target.checked)}
                            />
                            <label htmlFor="sync" className="text-sm">
                                Sync Processing (wait for embeddings and chunks generation)
                            </label>
                        </div>
                        
                        <div className="pt-4">
                            <Button 
                                className="w-full"
                                onClick={handleSubmit}
                                disabled={uploadMutation.isPending || !isFormValid}
                            >
                                {uploadMutation.isPending 
                                    ? (isSync ? 'Uploading & Processing...' : 'Uploading...') 
                                    : 'Upload Article'
                                }
                            </Button>
                        </div>

                        {/* Processing Information */}
                        <div className="bg-muted/30 rounded-md p-3 text-sm text-muted-foreground">
                            <p className="font-medium mb-2">What happens when you upload:</p>
                            <ul className="space-y-1 text-xs">
                                <li>• Article is saved to PostgreSQL database</li>
                                <li>• Content is automatically chunked into smaller segments</li>
                                <li>• OpenAI embeddings are generated for semantic search</li>
                                <li>• Vector embeddings are stored in Pinecone for fast retrieval</li>
                                {isSync && <li>• <strong>Sync mode:</strong> Wait for all processing to complete</li>}
                                {!isSync && <li>• <strong>Async mode:</strong> Background processing (faster upload)</li>}
                            </ul>
                        </div>
                    </div>

                </CardContent>
            </Card>
        </div>
    );
};