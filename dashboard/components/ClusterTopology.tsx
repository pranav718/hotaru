"use client";

import { motion } from "framer-motion";
import { NodeCard } from "./NodeCard";
import { ClusterNodeStatus } from "@/types/raft";
import { Activity, Radio } from "lucide-react";

interface ClusterTopologyProps {
  nodes: Record<number, ClusterNodeStatus>;
}

export function ClusterTopology({ nodes }: ClusterTopologyProps) {
  const nodeList = Object.values(nodes).sort((a, b) => a.id - b.id);
  const leaderNode = nodeList.find((n) => n.state === "Leader");
  const followerCount = nodeList.filter((n) => n.state === "Follower").length;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Activity className="h-4 w-4 text-[#C9F27D]" />
          <h2 className="font-mono text-sm font-semibold tracking-wide text-zinc-200 uppercase">
            cluster topology & quorum
          </h2>
        </div>
        <div className="flex items-center gap-3 font-mono text-xs text-zinc-500">
          {leaderNode && (
            <div className="flex items-center gap-1.5 text-[#C9F27D] bg-[#C9F27D]/10 border border-[#C9F27D]/20 px-2.5 py-0.5 rounded-full">
              <Radio className="h-3 w-3 animate-pulse text-[#C9F27D]" />
              <span>heartbeat active ({followerCount} peers)</span>
            </div>
          )}
          <div>
            active:{" "}
            <span className="text-zinc-300 font-semibold">
              {nodeList.filter((n) => n.state !== "Offline").length} /{" "}
              {nodeList.length || 3}
            </span>
          </div>
        </div>
      </div>

      <div className="relative">
        {leaderNode && (
          <motion.div
            animate={{ opacity: [0.2, 0.5, 0.2] }}
            transition={{ duration: 2.2, repeat: Infinity, ease: "easeInOut" }}
            className="absolute -inset-1 rounded-2xl bg-gradient-to-r from-[#C9F27D]/15 via-amber-500/5 to-cyan-500/10 blur-xl pointer-events-none"
          />
        )}

        <div className="relative grid grid-cols-1 md:grid-cols-3 gap-4">
          {nodeList.length === 0
            ? [0, 1, 2].map((id) => (
                <NodeCard
                  key={id}
                  node={{
                    id,
                    state: "Offline",
                    term: 0,
                    commit_index: 0,
                    last_applied: 0,
                    log_length: 0,
                    leader_id: -1,
                    peers: [],
                  }}
                />
              ))
            : nodeList.map((node) => <NodeCard key={node.id} node={node} />)}
        </div>
      </div>
    </div>
  );
}
