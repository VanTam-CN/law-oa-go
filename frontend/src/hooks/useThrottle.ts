import { useState, useEffect, useRef, useCallback } from "react";

export interface UseThrottleReturn<T> {
  throttledValue: T;
  setThrottledValue: (value: T) => void;
}

export function useThrottle<T>(value: T, delay: number): UseThrottleReturn<T> {
  const [throttledValue, setThrottledValueState] = useState<T>(value);
  const lastExecuted = useRef<number>(0);

  const setThrottledValue = useCallback(
    (newValue: T) => {
      const now = Date.now();
      if (now - lastExecuted.current >= delay) {
        setThrottledValueState(newValue);
        lastExecuted.current = now;
      }
    },
    [delay],
  );

  useEffect(() => {
    setThrottledValue(value);
  }, [value, setThrottledValue]);

  return {
    throttledValue,
    setThrottledValue,
  };
}

export default useThrottle;
