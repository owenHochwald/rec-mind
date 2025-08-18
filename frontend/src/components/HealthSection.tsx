import { useHealth } from "../hooks/useHealth";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "./ui/card";

export const HealthSection: React.FC = () => {
    const { data: healthStatus, isLoading, error } = useHealth();
    
    if (isLoading) return (<div>Health checks are loading...</div>);
    if (error) return (<div>Error: {error.message}</div>);

    // Helper function to get status color
    const getStatusColor = (status: string) => {
        switch (status?.toLowerCase()) {
            case 'healthy': return 'bg-green-500';
            case 'warning': return 'bg-yellow-500';
            case 'error': 
            case 'unhealthy': return 'bg-red-500';
            default: return 'bg-gray-500';
        }
    };

    const services = [
        {
            name: 'API Server',
            status: healthStatus?.status || 'unknown',
            details: `${healthStatus?.service} v${healthStatus?.version}`
        },
        {
            name: 'Database',
            status: healthStatus?.dependencies?.database?.status || 'unknown',
            details: healthStatus?.dependencies?.database?.response_time || 'No data'
        },
        {
            name: 'Redis Cache',
            status: healthStatus?.dependencies?.redis?.status || 'unknown',
            details: healthStatus?.dependencies?.redis?.response_time || 'No data'
        },
        {
            name: 'RabbitMQ',
            status: healthStatus?.dependencies?.rabbitmq?.status || 'unknown',
            details: 'Message Queue'
        },
        {
            name: 'Python ML Service',
            status: healthStatus?.dependencies?.python_ml_service?.status || 'unknown',
            details: healthStatus?.dependencies?.python_ml_service?.response_time || 'No data'
        },
        {
            name: 'Query RAG Worker',
            status: healthStatus?.dependencies?.query_rag_worker?.status || 'unknown',
            details: healthStatus?.dependencies?.query_rag_worker?.response_time || 'No data'
        }
    ];

    return (
        <div className="space-y-6">
            <div>
                <h2 className="text-ios-title font-bold">System Health</h2>
                <p className="text-muted-foreground mt-2">Monitor the status of all system components (auto-refreshes every 30s)</p>
            </div>
            
            {/* Overall Status */}
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center space-x-2">
                        <div className={`h-3 w-3 rounded-full ${getStatusColor(healthStatus?.status || 'unknown')}`} />
                        <span>System Overview</span>
                    </CardTitle>
                    <CardDescription>
                        {healthStatus?.service} - Uptime: {healthStatus?.uptime}
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <p className="text-sm text-muted-foreground">
                        Last updated: {new Date(healthStatus?.timestamp || '').toLocaleString()}
                    </p>
                </CardContent>
            </Card>

            {/* Service Status Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {services.map((service) => (
                    <Card key={service.name}>
                        <CardContent className="p-4">
                            <div className="flex items-center justify-between">
                                <div className="flex items-center space-x-3">
                                    <div className={`h-3 w-3 rounded-full ${getStatusColor(service.status)}`} />
                                    <div>
                                        <span className="font-medium">{service.name}</span>
                                        <p className="text-xs text-muted-foreground">{service.details}</p>
                                    </div>
                                </div>
                                <span className="text-sm text-muted-foreground capitalize">{service.status}</span>
                            </div>
                        </CardContent>
                    </Card>
                ))}
            </div>

            {/* Raw Health Data (for debugging) */}
            <Card>
                <CardHeader>
                    <CardTitle>Health Check Details</CardTitle>
                    <CardDescription>
                        Real-time system monitoring with TanStack Query auto-refresh
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <pre className="text-xs bg-muted p-3 rounded overflow-auto max-h-64">
                        {JSON.stringify(healthStatus, null, 2)}
                    </pre>
                </CardContent>
            </Card>
        </div>
    );
};