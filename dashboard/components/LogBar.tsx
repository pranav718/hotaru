"use client";

import { LogEntry } from "./LogEntry";
import { DisplayLogEntry } from "@/types/raft";
import { GitCommitHorizontal } from "lucide-react";

interface LogBarProps {
  nodeId: number;
  commitIndex: number;
  entries: DisplayLogEntry[];
}

export function LogBar({ nodeId, commitIndex, entries }: LogBarProps) {
  return (
    <div className="rounded-xl border border-zinc-800 bg-zinc-900/40 p-4 space-y-3 backdrop-blur-md">
      <div className="flex items-center justify-between font-mono text-xs">
        <div className="flex items-center gap-2">
          <GitCommitHorizontal className="h-4 w-4 text-emerald-400" />
          <span className="font-semibold text-zinc-200">node-{nodeId} log</span>
        </div>
        <div className="flex items-center gap-3 text-zinc-400 text-[11px]">
          <span>commit index: <strong className="text-emerald-400">{commitIndex}</strong></span>
          <span>total entries: <strong className="text-zinc-200">{entries.length}</strong></span>
        </div>
      </div>

      <div className="flex items-center gap-2 overflow-x-auto pb-2 pt-1 scrollbar-thin">
        {entries.length === 0 ? (
          <div className="py-3 px-4 font-mono text-xs text-zinc-600 italic">
            no entries in log
          </div>
        ) : (
          entries.map((entry) => (
            <LogEntry key={`${nodeId}-${entry.index}`} entry={entry} />
          ))
        )}
      </div>
    </div>
  );
}
