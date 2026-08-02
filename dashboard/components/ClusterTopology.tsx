"use client";

import { NodeCard } from "./NodeCard";
import { ClusterNodeStatus } from "@/types/raft";

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
        <h2 className="font-mono text-sm font-semibold tracking-wide text-zinc-200">
          cluster topology & quorum
        </h2>
        <div className="flex items-center gap-4 font-mono text-xs text-zinc-400">
          {leaderNode && (
            <span>
              heartbeat: <span className="text-[#C9F27D]">active</span>
            </span>
          )}
          <span>
            active: <span className="text-zinc-200">{nodeList.filter((n) => n.state !== "Offline").length} / {nodeList.length || 3}</span>
          </span>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
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
  );
}
