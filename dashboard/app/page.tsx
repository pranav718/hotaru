"use client";

import { useSSE } from "@/hooks/use-sse";
import { useClusterState } from "@/hooks/use-cluster-state";
import { ClusterTopology } from "@/components/ClusterTopology";
import { LogBar } from "@/components/LogBar";
import { EventFeed } from "@/components/EventFeed";
import { ControlPanel } from "@/components/ControlPanel";
import { DisplayLogEntry } from "@/types/raft";
import { Flame, ExternalLink, GitBranch, Wifi, WifiOff } from "lucide-react";

const SSE_ENDPOINTS = [
  "http://127.0.0.1:8010/events",
  "http://127.0.0.1:8011/events",
  "http://127.0.0.1:8012/events",
];

export default function DashboardPage() {
  const { events, isConnected } = useSSE(SSE_ENDPOINTS);
  const { nodes } = useClusterState();

  const getNodeLogs = (nodeId: number): DisplayLogEntry[] => {
    const nodeStatus = nodes[nodeId];
    const commitIdx = nodeStatus ? nodeStatus.commit_index : 0;

    const appends = events.filter(
      (e) => e.node_id === nodeId && e.type === "log_append"
    );

    return appends.map((e) => {
      const idx = Number(e.data?.index || 0);
      return {
        index: idx,
        term: e.term,
        command: String(e.data?.command || "entry"),
        isCommitted: idx <= commitIdx,
      };
    });
  };

  const hasActiveCluster =
    isConnected || Object.values(nodes).some((n) => n.state !== "Offline");

  return (
    <main className="flex-1 max-w-7xl mx-auto w-full px-4 py-8 space-y-8">
      <header className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 pb-6 border-b border-zinc-800">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br from-amber-500/20 to-orange-600/20 border border-amber-500/30 text-amber-400">
            <Flame className="h-5 w-5" />
          </div>
          <div>
            <h1 className="font-mono text-lg font-bold tracking-tight text-zinc-100 flex items-center gap-2">
              hotaru <span className="text-zinc-600 text-xs font-normal">v1.0</span>
            </h1>
            <p className="text-xs font-mono text-zinc-400">
              distributed raft consensus & key-value engine visualizer
            </p>
          </div>
        </div>

        <div className="flex items-center gap-4 text-xs font-mono">
          <div className="flex items-center gap-2 px-3 py-1.5 rounded-full border border-zinc-800 bg-zinc-900/80">
            {hasActiveCluster ? (
              <>
                <Wifi className="h-3.5 w-3.5 text-emerald-400 animate-pulse" />
                <span className="text-emerald-400 font-medium">cluster online</span>
              </>
            ) : (
              <>
                <WifiOff className="h-3.5 w-3.5 text-zinc-500" />
                <span className="text-zinc-500">disconnected (run main --dashboard)</span>
              </>
            )}
          </div>

          <a
            href="https://medium.com/@knightkun/raft-consensus-explained-simply-then-built-in-go-f642531b6527?sharedUserId=knightkun"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-1.5 text-zinc-400 hover:text-zinc-200 transition-colors"
          >
            <span>blog</span>
            <ExternalLink className="h-3 w-3" />
          </a>

          <a
            href="https://github.com/pranav718/hotaru"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-1.5 text-zinc-400 hover:text-zinc-200 transition-colors"
          >
            <GitBranch className="h-3.5 w-3.5 text-amber-400" />
            <span>github</span>
          </a>
        </div>
      </header>

      <section>
        <ClusterTopology nodes={nodes} />
      </section>

      <section className="space-y-3">
        <h2 className="font-mono text-sm font-semibold tracking-wide text-zinc-200 uppercase">
          log propagation & commit state
        </h2>
        <div className="space-y-3">
          {[0, 1, 2].map((nodeId) => (
            <LogBar
              key={nodeId}
              nodeId={nodeId}
              commitIndex={nodes[nodeId]?.commit_index || 0}
              entries={getNodeLogs(nodeId)}
            />
          ))}
        </div>
      </section>

      <section className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <EventFeed events={events} />
        <ControlPanel />
      </section>

      <footer className="pt-8 border-t border-zinc-800/80 flex flex-col sm:flex-row items-center justify-between text-xs font-mono text-zinc-400 gap-2">
        <div>hotaru raft engine &copy; 2026. built from scratch in go.</div>
        <div>
          read the raft paper breakdown on{" "}
          <a
            href="https://medium.com/@knightkun/raft-consensus-explained-simply-then-built-in-go-f642531b6527?sharedUserId=knightkun"
            target="_blank"
            rel="noopener noreferrer"
            className="text-amber-400 hover:underline"
          >
            medium
          </a>
        </div>
      </footer>
    </main>
  );
}
