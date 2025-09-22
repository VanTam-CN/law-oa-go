import { useState, useCallback } from 'react';

export interface UseErrorReturn {
  error: Error | null;
  setError: (error: Error | null) => void;
  clearError: () => void;
  handleError: (error: unknown, message?: string) => void;
}

export function useError(): UseErrorReturn {
  const [error, setError] = useState<Error | null>(null);

  const clearError = useCallback(() => {
    setError(null);
  }, []);

  const handleError = useCallback((error: unknown, message?: string) => {
    if (error instanceof Error) {
      setError(error);
    } else {
      setError(new Error(message || 'An unknown error occurred'));
    }
  }, []);

  return {
    error,
    setError,
    clearError,
    handleError,
  };
}

export default useError;
