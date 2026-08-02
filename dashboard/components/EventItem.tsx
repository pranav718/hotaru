"use client";

import { motion } from "framer-motion";
import { RefreshCw, Zap, CheckCircle2, XCircle, Activity, PlusCircle, GitCommit, Archive, UserPlus, UserMinus } from "lucide-react";
import { RaftEvent } from "@/types/raft";
import { cn } from "@/lib/utils";

interface EventItemProps {
  event: RaftEvent;
}

const META_MAP: Record<string, { icon: React.ElementType; color: string }> = {
  state_change: { icon: RefreshCw, color: "text-purple-400 border-purple-500/30 bg-purple-500/10" },
  election_started: { icon: Zap, color: "text-amber-400 border-amber-500/30 bg-amber-500/10" },
  vote_granted: { icon: CheckCircle2, color: "text-emerald-400 border-emerald-500/30 bg-emerald-500/10" },
  vote_denied: { icon: XCircle, color: "text-red-400 border-red-500/30 bg-red-500/10" },
  heartbeat: { icon: Activity, color: "text-blue-400 border-blue-500/30 bg-blue-500/10" },
  log_append: { icon: PlusCircle, color: "text-cyan-400 border-cyan-500/30 bg-cyan-500/10" },
  log_commit: { icon: GitCommit, color: "text-emerald-400 border-emerald-500/30 bg-emerald-500/10" },
  snapshot_installed: { icon: Archive, color: "text-indigo-400 border-indigo-500/30 bg-indigo-500/10" },
  peer_added: { icon: UserPlus, color: "text-teal-400 border-teal-500/30 bg-teal-500/10" },
  peer_staged: { icon: UserPlus, color: "text-teal-400 border-teal-500/30 bg-teal-500/10" },
  peer_removed: { icon: UserMinus, color: "text-rose-400 border-rose-500/30 bg-rose-500/10" },
};

export function EventItem({ event }: EventItemProps) {
  const meta = META_MAP[event.type] || { icon: Activity, color: "text-zinc-400 border-zinc-700 bg-zinc-800" };
  const Icon = meta.icon;
  const formattedTime = new Date(event.timestamp).toLocaleTimeString([], { hour12: false, hour: "2-digit", minute: "2-digit", second: "2-digit", fractionalSecondDigits: 3 });

  return (
    <motion.div initial={{ opacity: 0, y: -8 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0, scale: 0.95 }} transition={{ duration: 0.2 }} className="flex items-start gap-3 p-3 rounded-lg border border-zinc-800/80 bg-zinc-900/60 font-mono text-xs hover:border-zinc-700 transition-colors">
      <div className={cn("flex h-7 w-7 shrink-0 items-center justify-center rounded-md border", meta.color)}>
        <Icon className="h-3.5 w-3.5" />
      </div>
      <div className="flex-1 min-w-0 space-y-1">
        <div className="flex items-center justify-between gap-2">
          <div className="flex items-center gap-2">
            <span className="font-semibold text-zinc-200 text-[11px] tracking-wider">{event.type.replace(/_/g, " ").toLowerCase()}</span>
            <span className="text-[10px] text-zinc-500 bg-zinc-800 px-1.5 py-0.5 rounded">node-{event.node_id}</span>
            <span className="text-[10px] text-zinc-500">T{event.term}</span>
          </div>
          <span className="text-[10px] text-zinc-500 shrink-0">{formattedTime}</span>
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
