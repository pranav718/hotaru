export type NodeRole = "Follower" | "Candidate" | "Leader" | "Offline";

export interface RaftEvent {
  type:
    | "state_change"
    | "election_started"
    | "vote_granted"
    | "vote_denied"
    | "heartbeat"
    | "log_append"
    | "log_commit"
    | "peer_staged"
    | "peer_added"
    | "peer_removed"
    | "snapshot_installed";
  node_id: number;
  term: number;
  data?: Record<string, unknown>;
  timestamp: string;
}

export interface ClusterNodeStatus {
  id: number;
  state: NodeRole;
  term: number;
  commit_index: number;
  last_applied: number;
  log_length: number;
  leader_id: number;
  peers: number[];
}

export interface DisplayLogEntry {
  index: number;
  term: number;
  command: string;
  isCommitted: boolean;
}
