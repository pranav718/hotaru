"use client";

import { useEffect, useState } from "react";
import { ClusterNodeStatus } from "@/types/raft";

const DEFAULT_NODE_PORTS = [8010, 8011, 8012];

export function useClusterState() {
  const [nodes, setNodes] = useState<Record<number, ClusterNodeStatus>>({});

  useEffect(() => {
    let isMounted = true;

    async function fetchClusterStatus() {
      const updated: Record<number, ClusterNodeStatus> = {};

      await Promise.all(
        DEFAULT_NODE_PORTS.map(async (port, index) => {
          try {
            const res = await fetch(`http://127.0.0.1:${port}/status`, {
              cache: "no-store",
            });
            if (res.ok) {
              const status: ClusterNodeStatus = await res.json();
              updated[status.id] = status;
            } else {
              updated[index] = {
                id: index,
                state: "Offline",
                term: 0,
                commit_index: 0,
                last_applied: 0,
                log_length: 0,
                leader_id: -1,
                peers: [],
              };
            }
          } catch {
            updated[index] = {
              id: index,
              state: "Offline",
              term: 0,
              commit_index: 0,
              last_applied: 0,
              log_length: 0,
              leader_id: -1,
              peers: [],
            };
          }
        })
      );

      if (isMounted) {
        setNodes(updated);
      }
    }

    fetchClusterStatus();
    const interval = setInterval(fetchClusterStatus, 1000);

    return () => {
      isMounted = false;
      clearInterval(interval);
    };
  }, []);

  return { nodes };
}
