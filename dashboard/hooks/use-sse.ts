"use client";

import { useEffect, useState, useRef } from "react";
import { RaftEvent } from "@/types/raft";

export function useSSE(endpoints: string[]) {
  const [events, setEvents] = useState<RaftEvent[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const sourcesRef = useRef<EventSource[]>([]);
  const endpointKey = endpoints.join(",");

  useEffect(() => {
    sourcesRef.current.forEach((src) => src.close());
    sourcesRef.current = [];

    let activeCount = 0;

    endpoints.forEach((url) => {
      try {
        const es = new EventSource(url);

        es.onopen = () => {
          activeCount++;
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
          // keep connected status true if at least one node is alive
        };

        sourcesRef.current.push(es);
      } catch {
        // ignore unreachable nodes
      }
    });

    return () => {
      sourcesRef.current.forEach((src) => src.close());
      sourcesRef.current = [];
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [endpointKey]);

  return { events, isConnected };
}
