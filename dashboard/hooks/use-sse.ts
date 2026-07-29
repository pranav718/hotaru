"use client";

import { useEffect, useState, useRef } from "react";
import { RaftEvent } from "@/types/raft";

export function useSSE(endpoints: string[]) {
  const [events, setEvents] = useState<RaftEvent[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const sourcesRef = useRef<EventSource[]>([]);

  useEffect(() => {
    sourcesRef.current.forEach((src) => src.close());
    sourcesRef.current = [];

    let connectedCount = 0;

    endpoints.forEach((url) => {
      try {
        const es = new EventSource(url);

        es.onopen = () => {
          connectedCount++;
          setIsConnected(true);
        };

        es.onmessage = (event) => {
          try {
            const parsed: RaftEvent = JSON.parse(event.data);
            setEvents((prev) => [parsed, ...prev.slice(0, 199)]);
          } catch {
            // ignore malformed events
          }
        };

        es.onerror = () => {
          connectedCount = Math.max(0, connectedCount - 1);
          if (connectedCount === 0) setIsConnected(false);
        };

        sourcesRef.current.push(es);
      } catch {
        // fail silently for unreachable nodes
      }
    });

    return () => {
      sourcesRef.current.forEach((src) => src.close());
      sourcesRef.current = [];
    };
  }, [endpoints]);

  return { events, isConnected };
}
