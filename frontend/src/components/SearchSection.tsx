import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "./ui/card";

export const SearchSection: React.FC = () => {
    // TODO: Practice TanStack Query for search functionality
    // const [query, setQuery] = useState('');
    // const [topK, setTopK] = useState(10);
    // const [scoreThreshold, setScoreThreshold] = useState(0.7);
    // const [searchMode, setSearchMode] = useState('immediate');
    // const [jobId, setJobId] = useState<string | null>(null);
    // 
    // const searchImmediateMutation = useSearchImmediateMutation();
    // const searchAsyncMutation = useSearchAsync();
    // const jobStatusQuery = useSearchJobStatus(jobId || '');

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
                    <p className="text-sm text-muted-foreground">TODO: Use searchImmediateMutation or searchAsyncMutation</p>
                    <p className="text-sm text-muted-foreground">TODO: Use useSearchJobStatus for polling async jobs</p>
                    <p className="text-sm text-muted-foreground">TODO: Display results from mutation.data</p>
                    <p className="text-sm text-muted-foreground">TODO: Show loading states with mutation.isPending</p>

                    {/* Sample search interface for practice:
          <div className="space-y-4">
            <div>
              <label className="text-sm font-medium">Search Query</label>
              <Input placeholder="What are you looking for?" />
            </div>
            
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="text-sm font-medium">Top Results</label>
                <Input type="number" defaultValue="10" min="1" max="50" />
              </div>
              <div>
                <label className="text-sm font-medium">Score Threshold</label>
                <Input type="number" step="0.1" defaultValue="0.7" min="0" max="1" />
              </div>
            </div>
            
            <div className="flex space-x-4">
              <label className="flex items-center space-x-2">
                <input type="radio" name="searchMode" value="immediate" defaultChecked />
                <span className="text-sm">Immediate</span>
              </label>
              <label className="flex items-center space-x-2">
                <input type="radio" name="searchMode" value="async" />
                <span className="text-sm">Async (with polling)</span>
              </label>
            </div>
            
            <Button 
              className="w-full"
              onClick={() => {
                if (searchMode === 'immediate') {
                  searchImmediateMutation.mutate({ query, topK, scoreThreshold });
                } else {
                  searchAsyncMutation.mutate({ query, topK, scoreThreshold });
                }
              }}
              disabled={searchImmediateMutation.isPending || searchAsyncMutation.isPending}
            >
              {(searchImmediateMutation.isPending || searchAsyncMutation.isPending) ? 'Searching...' : 'Search'}
            </Button>
          </div>
          */}
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>Search Results</CardTitle>
                    <CardDescription>
                        Results will appear here with similarity scores
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <p className="text-sm text-muted-foreground text-center py-8">No search results yet</p>
                </CardContent>
            </Card>
        </div>
    );
};