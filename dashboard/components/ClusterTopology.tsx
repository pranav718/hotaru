"use client";

import { NodeCard } from "./NodeCard";
import { ClusterNodeStatus } from "@/types/raft";
import { Activity } from "lucide-react";

interface ClusterTopologyProps {
  nodes: Record<number, ClusterNodeStatus>;
}

export function ClusterTopology({ nodes }: ClusterTopologyProps) {
  const nodeList = Object.values(nodes).sort((a, b) => a.id - b.id);
  const leaderNode = nodeList.find((n) => n.state === "Leader");

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Activity className="h-4 w-4 text-emerald-400" />
          <h2 className="font-mono text-sm font-semibold tracking-wide text-zinc-200 uppercase">
            cluster topology & quorum
          </h2>
        </div>
        <div className="font-mono text-xs text-zinc-500">
          active nodes: <span className="text-zinc-300 font-semibold">{nodeList.filter(n => n.state !== "Offline").length} / {nodeList.length || 3}</span>
          {leaderNode && (
            <span className="ml-3 text-emerald-400">
              leader: node-{leaderNode.id}
            </span>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {nodeList.length === 0 ? (
          [0, 1, 2].map((id) => (
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
        ) : (
          nodeList.map((node) => <NodeCard key={node.id} node={node} />)
        )}
      </div>
    </div>
  );
}
