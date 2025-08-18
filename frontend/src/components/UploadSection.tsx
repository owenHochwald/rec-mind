import { useState } from "react";
import { useUploadArticle } from "../hooks/useArticles";
import { Button } from "./ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "./ui/card";
import { Input } from "./ui/input";
import { title } from "process";

export const UploadSection: React.FC = () => {
    const uploadMutation = useUploadArticle();
    const [formData, setFormData] = useState({ title: '', content: '', url: '', category: '' });
    const [isSync, setIsSync] = useState(false);

    return (
        <div className="space-y-6">
            <div>
                <h2 className="text-ios-title font-bold">Upload Article</h2>
                <p className="text-muted-foreground mt-2">Add new articles to your knowledge base</p>
            </div>

            <Card>
                <CardHeader>
                    <CardTitle>Article Details</CardTitle>
                    <CardDescription>
                        Fill in the article information for processing
                    </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="space-y-4">
                        {/* Title */}
                        <div>
                            <label className="text-sm font-medium">Title</label>
                            <Input
                                placeholder="Enter article title..."
                                value={formData.title}
                                onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                            />
                        </div>
                        {/* Content */}
                        <div>
                            <label className="text-sm font-medium">Content</label>
                            <textarea
                                className="min-h-[200px] w-full rounded-ios border border-input bg-background px-3 py-2 text-sm"
                                placeholder="Paste article content..."
                                value={formData.content}
                                onChange={(e) => setFormData({ ...formData, content: e.target.value })}
                            />
                        </div>
                        <div className="grid grid-cols-2 gap-4">
                            {/* URL */}
                            <div>
                                <label className="text-sm font-medium">URL</label>
                                <Input
                                    type="url"
                                    placeholder="https://..."
                                    value={formData.url}
                                    onChange={(e) => setFormData({ ...formData, url: e.target.value })}
                                />
                            </div>
                            {/* Category */}
                            <div>
                                <label className="text-sm font-medium">Category</label>
                                <Input
                                    placeholder="e.g., technology"
                                    value={formData.category}
                                    onChange={(e) => setFormData({ ...formData, category: e.target.value })} />
                            </div>
                        </div>
                        {/* Sync Option */}
                        <div className="flex items-center space-x-2">
                            <input
                                type="checkbox"
                                id="sync"
                                checked={isSync}
                                onChange={(e) => setIsSync(e.target.checked)}
                            />
                            <label htmlFor="sync" className="text-sm">Sync Processing (wait for ML completion)</label>
                        </div>
                        <Button
                            className="w-full"
                            onClick={() => uploadMutation.mutate({ article: formData, isSync })}
                            disabled={uploadMutation.isPending}
                        >
                            {uploadMutation.isPending ? 'Uploading...' : 'Upload Article'}
                        </Button>
                    </div>

                </CardContent>
            </Card>
        </div>
    );
};