"use client";

import { useSSE } from "@/hooks/use-sse";
import { useClusterState } from "@/hooks/use-cluster-state";
import { ClusterTopology } from "@/components/ClusterTopology";
import { LogBar } from "@/components/LogBar";
import { EventFeed } from "@/components/EventFeed";
import { ControlPanel } from "@/components/ControlPanel";
import { DisplayLogEntry } from "@/types/raft";
import { ExternalLink } from "lucide-react";

import { InteractiveLink } from "@/components/InteractiveLink";

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



  return (
    <main className="flex-1 max-w-7xl mx-auto w-full px-4 py-8 space-y-8">
      <header className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 pb-6 border-b border-zinc-800/80">
        <div>
          <h1>
            <span className="font-serif italic font-normal text-3xl tracking-normal text-[#C9F27D]">
              hotaru
            </span>
          </h1>
          <p className="text-xs font-mono text-zinc-400">
            distributed raft consensus & key-value engine visualizer
          </p>
        </div>

        <div className="flex items-center gap-4 text-xs font-mono">
          <InteractiveLink href="https://medium.com/@knightkun/raft-consensus-explained-simply-then-built-in-go-f642531b6527?sharedUserId=knightkun">
            blog
          </InteractiveLink>

          <InteractiveLink
            href="https://github.com/pranav718/hotaru"
            showArrow={false}
            icon={
              <svg className="h-4 w-4" viewBox="0 0 24 24" fill="currentColor">
                <path d="M12 0C5.374 0 0 5.373 0 12c0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23A11.509 11.509 0 0 1 12 5.803c1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576C20.566 21.797 24 17.3 24 12c0-6.627-5.373-12-12-12z"/>
              </svg>
            }
          >
            github
          </InteractiveLink>
        </div>
      </header>

      <section>
        <ClusterTopology nodes={nodes} />
      </section>

      <section className="space-y-3">
        <h2 className="font-mono text-sm font-semibold tracking-wide text-zinc-200">
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
        <div>hotaru raft engine. built from scratch in go.</div>
        <div>
          read{" "}
          <a
            href="https://medium.com/@knightkun/raft-consensus-explained-simply-then-built-in-go-f642531b6527?sharedUserId=knightkun"
            target="_blank"
            rel="noopener noreferrer"
            className="text-[#C9F27D] hover:underline"
          >
            raft consensus, explained simply (then built in go)
          </a>
          {" "}on medium
        </div>
      </footer>
    </main>
  );
}
