import { useState, useEffect } from 'react';
import { apiClient, ApiRequestTiming } from '../services/api';

export function useApiTiming() {
  const [lastTiming, setLastTiming] = useState<ApiRequestTiming | null>(null);

  useEffect(() => {
    // Check for timing updates every 100ms
    const interval = setInterval(() => {
      const currentTiming = apiClient.getLastRequestTiming();
      if (currentTiming && currentTiming.timestamp !== lastTiming?.timestamp) {
        setLastTiming(currentTiming);
      }
    }, 100);

    return () => clearInterval(interval);
  }, [lastTiming?.timestamp]);

  return lastTiming;
}