import React, { useState } from 'react';
import { Button } from './ui/button';
import { cn } from '../lib/utils';
import { HealthSection } from './HealthSection';
import { ArticlesSection } from './ArticlesSection';
import { UploadSection } from './UploadSection';
import { SearchSection } from './SearchSection';
import { useApiTiming } from '../hooks/useApiTiming';

const Dashboard: React.FC = () => {
    const [activeSection, setActiveSection] = useState<string>('articles');
    const lastTiming = useApiTiming();

    return (
        <div className="min-h-screen bg-background text-foreground">
            {/* Header */}
            <header className="border-b border-border/40 backdrop-blur-md bg-background/95 sticky top-0 z-50">
                <div className="flex h-16 items-center justify-between px-6">
                    <div className="flex items-center space-x-4">
                        <h1 className="text-ios-title font-bold">RecMind</h1>
                        <div className="hidden md:flex items-center space-x-1">
                            {/* Status indicator */}
                            <div className="h-2 w-2 bg-accent-green rounded-full animate-pulse" />
                            <span className="text-sm text-muted-foreground">Online</span>
                        </div>
                    </div>

                    <div className="flex items-center space-x-4">
                        {lastTiming && (
                            <div className="text-xs text-muted-foreground bg-muted/30 px-3 py-1 rounded-ios">
                                <div className="flex items-center space-x-2">
                                    <span className="font-medium">{lastTiming.description}:</span>
                                    <span className="text-accent-green">{Math.round(lastTiming.duration)}ms</span>
                                </div>
                            </div>
                        )}
                        {!lastTiming && (
                            <div className="text-xs text-muted-foreground bg-muted/30 px-3 py-1 rounded-ios">
                                Last Request: --
                            </div>
                        )}
                    </div>
                </div>
            </header>

            <div className="flex">
                {/* Sidebar Navigation */}
                <nav className="w-64 bg-secondary/30 backdrop-blur-md border-r border-border/40 min-h-[calc(100vh-4rem)]">
                    <div className="p-4 space-y-2">
                        {[
                            { id: 'articles', label: 'Articles' },
                            { id: 'upload', label: 'Upload' },
                            { id: 'search', label: 'Search' },
                            { id: 'health', label: 'Health' }
                        ].map((item) => (
                            <Button
                                key={item.id}
                                variant={activeSection === item.id ? "default" : "ghost"}
                                className={cn(
                                    "w-full justify-start text-left h-12",
                                    activeSection === item.id && "bg-primary text-primary-foreground"
                                )}
                                onClick={() => setActiveSection(item.id)}
                            >
                                {item.label}
                            </Button>
                        ))}
                    </div>
                </nav>

                {/* Main Content */}
                <main className="flex-1 p-6">
                    <div className="max-w-6xl">
                        {activeSection === 'articles' && <ArticlesSection />}
                        {activeSection === 'upload' && <UploadSection />}
                        {activeSection === 'search' && <SearchSection />}
                        {activeSection === 'health' && <HealthSection />}
                    </div>
                </main>
            </div>
        </div>
    );
};





export default Dashboard;