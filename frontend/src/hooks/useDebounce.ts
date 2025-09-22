import { useState, useEffect } from 'react';

export interface UseDebounceReturn<T> {
  debouncedValue: T;
  setDebouncedValue: (value: T) => void;
}

export function useDebounce<T>(value: T, delay: number): UseDebounceReturn<T> {
  const [debouncedValue, setDebouncedValue] = useState<T>(value);

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);

    return () => {
      clearTimeout(handler);
    };
  }, [value, delay]);

  return {
    debouncedValue,
    setDebouncedValue,
  };
}

export default useDebounce;
