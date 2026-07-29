"use client";

import { useState } from "react";
import { AnimatePresence } from "framer-motion";
import { EventItem } from "./EventItem";
import { RaftEvent } from "@/types/raft";
import { Radio, Pause, Play, Trash2 } from "lucide-react";

interface EventFeedProps {
  events: RaftEvent[];
}

export function EventFeed({ events: initialEvents }: EventFeedProps) {
  const [isPaused, setIsPaused] = useState(false);
  const [frozenEvents, setFrozenEvents] = useState<RaftEvent[]>([]);

  const togglePause = () => {
    if (!isPaused) {
      setFrozenEvents(initialEvents);
    }
    setIsPaused(!isPaused);
  };

  const displayedEvents = isPaused ? frozenEvents : initialEvents;

  return (
    <div className="rounded-xl border border-zinc-800 bg-zinc-900/40 p-4 space-y-3 backdrop-blur-md">
      <div className="flex items-center justify-between font-mono text-xs">
        <div className="flex items-center gap-2">
          <Radio className="h-4 w-4 text-purple-400 animate-pulse" />
          <h2 className="font-semibold text-zinc-200 uppercase tracking-wide">
            live event stream
          </h2>
          <span className="text-[10px] text-zinc-500 bg-zinc-800 px-2 py-0.5 rounded-full">
            {displayedEvents.length} events
          </span>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={togglePause}
            className="flex items-center gap-1 px-2.5 py-1 rounded bg-zinc-800 hover:bg-zinc-700 text-zinc-300 text-[11px] transition-colors"
          >
            {isPaused ? (
              <>
                <Play className="h-3 w-3 text-emerald-400" />
                <span>resume</span>
              </>
            ) : (
              <>
                <Pause className="h-3 w-3 text-amber-400" />
                <span>pause</span>
              </>
            )}
          </button>
        </div>
      </div>

      <div className="max-h-[360px] overflow-y-auto space-y-2 pr-1 scrollbar-thin">
        {displayedEvents.length === 0 ? (
          <div className="py-8 text-center font-mono text-xs text-zinc-600 italic">
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
