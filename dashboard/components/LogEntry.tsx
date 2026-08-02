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
      initial={{ scale: 0.95, opacity: 0 }}
      animate={{ scale: 1, opacity: 1 }}
      className={cn(
        "relative flex flex-col justify-between min-w-[100px] px-3 py-2 rounded-lg border font-mono text-xs shrink-0",
        entry.isCommitted
          ? "bg-[#C9F27D]/10 border-[#C9F27D]/30 text-[#C9F27D]"
          : "bg-zinc-900/60 border-zinc-800 text-zinc-400"
      )}
    >
      <div className="flex items-center justify-between w-full text-[10px] text-zinc-400 border-b border-zinc-800/60 pb-1 mb-1">
        <span>#{entry.index}</span>
        <span className="text-zinc-500">T{entry.term}</span>
      </div>

      <div className="text-[11px] font-semibold text-zinc-100 whitespace-nowrap px-0.5">
        {entry.command || "noop"}
      </div>

      <div className="mt-1.5 flex items-center justify-between">
        <span className="text-[9px] text-zinc-500">
          {entry.isCommitted ? "commit" : "pending"}
        </span>
      </div>
    </motion.div>
  );
}
