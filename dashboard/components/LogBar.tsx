"use client";

import { LogEntry } from "./LogEntry";
import { DisplayLogEntry } from "@/types/raft";

interface LogBarProps {
  nodeId: number;
  commitIndex: number;
  entries: DisplayLogEntry[];
}

export function LogBar({ nodeId, entries }: LogBarProps) {
  return (
    <div className="relative rounded-xl border border-zinc-800 bg-zinc-900/40 p-4 space-y-3 backdrop-blur-md overflow-hidden">
      <div className="absolute inset-0 bg-[linear-gradient(to_bottom,transparent_50%,rgba(0,0,0,0.25)_51%)] bg-[length:100%_4px] pointer-events-none opacity-30" />

      <div className="relative z-10 flex items-center justify-between font-mono text-xs">
        <div>
          <span className="font-semibold text-zinc-200">node-{nodeId} log</span>
        </div>
      </div>

      <div className="relative z-10 flex items-center gap-2 overflow-x-auto pb-2 pt-1 scrollbar-thin">
        {entries.length === 0 ? (
          <div className="py-1 font-mono text-xs text-zinc-500">
            no log entries
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
