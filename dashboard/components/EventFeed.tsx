"use client";

import { useState, useRef, useEffect } from "react";
import { AnimatePresence } from "framer-motion";
import { EventItem } from "./EventItem";
import { RaftEvent } from "@/types/raft";
import { Pause, Play } from "lucide-react";

interface EventFeedProps {
  events: RaftEvent[];
}

export function EventFeed({ events: initialEvents }: EventFeedProps) {
  const [isPaused, setIsPaused] = useState(false);
  const [frozenEvents, setFrozenEvents] = useState<RaftEvent[]>([]);
  const scrollRef = useRef<HTMLDivElement>(null);

  const togglePause = () => {
    if (!isPaused) {
      setFrozenEvents(initialEvents);
    }
    setIsPaused(!isPaused);
  };

  const displayedEvents = isPaused ? frozenEvents : initialEvents;
  const latestTimestamp = displayedEvents[0]?.timestamp;

  useEffect(() => {
    if (!isPaused && scrollRef.current) {
      if (scrollRef.current.scrollTop <= 50) {
        scrollRef.current.scrollTo({ top: 0, behavior: "smooth" });
      }
    }
  }, [latestTimestamp, displayedEvents.length, isPaused]);

  return (
    <div className="rounded-xl border border-zinc-800 bg-zinc-900/40 p-4 space-y-3 backdrop-blur-md">
      <div className="flex items-center justify-between font-mono text-xs">
        <div>
          <h2 className="font-semibold text-zinc-200 tracking-wide">
            live event stream
          </h2>
        </div>

        <button
          onClick={togglePause}
          className="flex items-center gap-1.5 px-2.5 py-1 rounded bg-zinc-800/80 hover:bg-zinc-700/80 text-[#C9F27D] text-[11px] transition-colors border border-zinc-700/50"
        >
          {isPaused ? (
            <>
              <Play className="h-3 w-3 text-[#C9F27D]" />
              <span>resume</span>
            </>
          ) : (
            <>
              <Pause className="h-3 w-3 text-[#C9F27D]" />
              <span>pause</span>
            </>
          )}
        </button>
      </div>

      <div
        ref={scrollRef}
        data-lenis-prevent
        className={`max-h-[360px] min-h-[260px] overflow-y-auto space-y-2 pr-1 scrollbar-thin ${displayedEvents.length === 0 ? "flex items-center justify-center" : ""}`}
      >
        {displayedEvents.length === 0 ? (
          <div className="font-mono text-xs text-zinc-500 text-center">
            waiting for cluster events...
          </div>
        ) : (
          <AnimatePresence initial={false}>
            {displayedEvents.map((evt, idx) => (
              <EventItem key={`${evt.timestamp}-${idx}`} event={evt} />
            ))}
          </AnimatePresence>
        )}
      </div>
    </div>
  );
}
