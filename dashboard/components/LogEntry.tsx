"use client";

import { motion } from "framer-motion";
import { DisplayLogEntry } from "@/types/raft";
import { cn } from "@/lib/utils";

interface LogEntryProps {
  entry: DisplayLogEntry;
}

export function LogEntry({ entry }: LogEntryProps) {
  return (
    <motion.div
      initial={{ scale: 0.8, opacity: 0, x: -10 }}
      animate={{ scale: 1, opacity: 1, x: 0 }}
      exit={{ scale: 0.8, opacity: 0 }}
      transition={{ type: "spring", stiffness: 400, damping: 25 }}
      className={cn(
        "group relative flex flex-col items-center justify-between min-w-[75px] px-2.5 py-2 rounded-lg border font-mono text-xs cursor-pointer transition-all duration-200 hover:scale-105",
        entry.isCommitted
          ? "bg-[#C9F27D]/10 border-[#C9F27D]/40 text-[#C9F27D] shadow-[0_0_12px_rgba(201,242,125,0.12)]"
          : "bg-zinc-900/60 border-zinc-700/50 text-zinc-400"
      )}
    >
      <div className="flex items-center justify-between w-full text-[10px] text-zinc-400 border-b border-zinc-800/60 pb-1 mb-1">
        <span>#{entry.index}</span>
        <span className="text-zinc-500">T{entry.term}</span>
      </div>

      <div className="truncate max-w-[85px] text-[11px] font-semibold text-zinc-100">
        {entry.command || "noop"}
      </div>

      <div className="mt-1.5 flex items-center gap-1">
        <span
          className={cn(
            "h-1.5 w-1.5 rounded-full",
            entry.isCommitted ? "bg-[#C9F27D] animate-pulse" : "bg-amber-400"
          )}
        />
        <span className="text-[9px] uppercase tracking-wider text-zinc-500">
          {entry.isCommitted ? "commit" : "pend"}
        </span>
      </div>
    </motion.div>
  );
}
