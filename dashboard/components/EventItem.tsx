"use client";

import { motion } from "framer-motion";
import { RaftEvent } from "@/types/raft";

interface EventItemProps {
  event: RaftEvent;
}

export function EventItem({ event }: EventItemProps) {
  const formattedTime = new Date(event.timestamp).toLocaleTimeString([], {
    hour12: false,
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    fractionalSecondDigits: 3,
  });

  return (
    <motion.div
      initial={{ opacity: 0, y: -8 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, scale: 0.95 }}
      transition={{ duration: 0.2 }}
      className="flex items-start p-3 rounded-lg border border-zinc-800/80 bg-zinc-900/60 font-mono text-xs hover:border-zinc-700 transition-colors"
    >
      <div className="flex-1 min-w-0 space-y-1">
        <div className="flex items-center justify-between gap-2">
          <div className="flex items-center gap-2">
            <span className="font-semibold text-zinc-200 text-[11px] tracking-wider">
              {event.type.replace(/_/g, " ").toLowerCase()}
            </span>
            <span className="text-[10px] text-zinc-500 bg-zinc-800 px-1.5 py-0.5 rounded">
              node-{event.node_id}
            </span>
            <span className="text-[10px] text-zinc-500">T{event.term}</span>
          </div>
          <span className="text-[10px] text-zinc-500 shrink-0">
            {formattedTime}
          </span>
        </div>
        {event.data && Object.keys(event.data).length > 0 && (
          <div className="text-[11px] text-zinc-400 bg-zinc-950/40 p-1.5 rounded border border-zinc-800/40 truncate">
            {JSON.stringify(event.data).replace(/[{}]/g, "")}
          </div>
        )}
      </div>
    </motion.div>
  );
}
