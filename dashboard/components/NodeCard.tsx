"use client";

import { motion } from "framer-motion";
import { Crown, Shield, Zap, AlertCircle, HardDrive } from "lucide-react";
import { ClusterNodeStatus } from "@/types/raft";
import { cn } from "@/lib/utils";

interface NodeCardProps {
  node: ClusterNodeStatus;
}

export function NodeCard({ node }: NodeCardProps) {
  const isLeader = node.state === "Leader";
  const isCandidate = node.state === "Candidate";
  const isOffline = node.state === "Offline";

  const roleStyles = {
    Leader:
      "border-emerald-500/40 bg-emerald-950/20 text-emerald-400 shadow-[0_0_25px_rgba(16,185,129,0.15)]",
    Candidate:
      "border-amber-500/40 bg-amber-950/20 text-amber-400 shadow-[0_0_25px_rgba(245,158,11,0.15)] animate-pulse",
    Follower: "border-zinc-800 bg-zinc-900/60 text-zinc-300",
    Offline: "border-red-900/40 bg-red-950/20 text-red-400 opacity-60",
  };

  const badgeStyles = {
    Leader: "bg-emerald-500/10 text-emerald-400 border-emerald-500/30",
    Candidate: "bg-amber-500/10 text-amber-400 border-amber-500/30",
    Follower: "bg-zinc-800 text-zinc-400 border-zinc-700",
    Offline: "bg-red-500/10 text-red-400 border-red-500/30",
  };

  return (
    <motion.div
      layout
      whileHover={{ y: -3, transition: { duration: 0.2 } }}
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      transition={{ type: "spring", stiffness: 300, damping: 25 }}
      className={cn(
        "relative rounded-xl border p-5 transition-colors duration-300 backdrop-blur-md",
        roleStyles[node.state] || roleStyles.Follower
      )}
    >
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2.5">
          <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-zinc-800/80 border border-zinc-700/50">
            <HardDrive className="h-4 w-4 text-zinc-300" />
          </div>
          <div>
            <h3 className="font-mono text-sm font-semibold tracking-wide text-zinc-100">
              node-{node.id}
            </h3>
            <p className="text-[11px] font-mono text-zinc-500">
              port: {8000 + node.id}
            </p>
          </div>
        </div>

        <motion.div
          layout
          key={node.state}
          initial={{ scale: 0.8, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          transition={{ type: "spring", stiffness: 400, damping: 20 }}
          className={cn(
            "flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-mono font-medium border",
            badgeStyles[node.state]
          )}
        >
          {isLeader && <Crown className="h-3.5 w-3.5" />}
          {isCandidate && <Zap className="h-3.5 w-3.5" />}
          {!isLeader && !isCandidate && !isOffline && (
            <Shield className="h-3.5 w-3.5" />
          )}
          {isOffline && <AlertCircle className="h-3.5 w-3.5" />}
          <span>{node.state.toLowerCase()}</span>
        </motion.div>
      </div>

      <div className="grid grid-cols-2 gap-2.5 pt-3 border-t border-zinc-800/60 font-mono text-xs">
        <div className="bg-zinc-950/40 rounded-lg p-2.5 border border-zinc-800/40">
          <span className="text-zinc-500 block text-[10px] uppercase tracking-wider mb-0.5">
            term
          </span>
          <span className="text-zinc-200 font-semibold">{node.term}</span>
        </div>
        <div className="bg-zinc-950/40 rounded-lg p-2.5 border border-zinc-800/40">
          <span className="text-zinc-500 block text-[10px] uppercase tracking-wider mb-0.5">
            commit index
          </span>
          <span className="text-zinc-200 font-semibold">{node.commit_index}</span>
        </div>
        <div className="bg-zinc-950/40 rounded-lg p-2.5 border border-zinc-800/40">
          <span className="text-zinc-500 block text-[10px] uppercase tracking-wider mb-0.5">
            last applied
          </span>
          <span className="text-zinc-200 font-semibold">{node.last_applied}</span>
        </div>
        <div className="bg-zinc-950/40 rounded-lg p-2.5 border border-zinc-800/40">
          <span className="text-zinc-500 block text-[10px] uppercase tracking-wider mb-0.5">
            log size
          </span>
          <span className="text-zinc-200 font-semibold">{node.log_length}</span>
        </div>
      </div>
    </motion.div>
  );
}
