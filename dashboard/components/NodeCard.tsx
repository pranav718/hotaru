"use client";

import { motion } from "framer-motion";
import { ClusterNodeStatus } from "@/types/raft";
import { cn } from "@/lib/utils";

interface NodeCardProps {
  node: ClusterNodeStatus;
}

export function NodeCard({ node }: NodeCardProps) {
  const roleStyles = {
    Leader: "border-[#C9F27D]/40 bg-[#C9F27D]/5 text-[#C9F27D]",
    Candidate: "border-amber-500/40 bg-amber-950/20 text-amber-400 animate-pulse",
    Follower: "border-zinc-800 bg-zinc-900/60 text-zinc-300",
    Offline: "border-red-900/40 bg-red-950/20 text-red-400 opacity-60",
  };

  const badgeStyles = {
    Leader: "text-[#C9F27D]",
    Candidate: "text-amber-400",
    Follower: "text-zinc-500",
    Offline: "text-red-400",
  };

  return (
    <motion.div
      layout
      whileHover={{ y: -3, transition: { duration: 0.2 } }}
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      transition={{ type: "spring", stiffness: 300, damping: 25 }}
      className={cn("relative rounded-xl border p-5 transition-colors duration-300 backdrop-blur-md overflow-hidden", roleStyles[node.state] || roleStyles.Follower)}
    >
      <div className="flex items-center justify-between mb-4">
        <div>
          <h3 className="font-mono text-sm font-semibold tracking-wide text-zinc-100">node-{node.id}</h3>
          <p className="text-[11px] font-mono text-zinc-500">port: {8000 + node.id}</p>
        </div>

        <motion.div
          layout
          key={node.state}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          className={cn("text-xs font-mono font-medium tracking-wide", badgeStyles[node.state])}
        >
          <span>{node.state.toLowerCase()}</span>
        </motion.div>
      </div>

      <div className="grid grid-cols-2 gap-2.5 pt-3 border-t border-zinc-800/60 font-mono text-xs">
        {[
          { label: "term", val: node.term },
          { label: "commit index", val: node.commit_index },
          { label: "last applied", val: node.last_applied },
          { label: "log size", val: node.log_length },
        ].map(({ label, val }) => (
          <div key={label} className="bg-zinc-950/40 rounded-lg p-2.5 border border-zinc-800/40">
            <span className="text-zinc-500 block text-[10px] tracking-wider mb-0.5">{label}</span>
            <span className="text-zinc-200 font-semibold">{val}</span>
          </div>
        ))}
      </div>
    </motion.div>
  );
}
