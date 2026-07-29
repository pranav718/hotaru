package raft

import "time"

type RaftEvent struct {
	Type      string                 `json:"type"`
	NodeID    int                    `json:"node_id"`
	Term      int                    `json:"term"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

type ClusterNodeStatus struct {
	ID          int      `json:"id"`
	State       string   `json:"state"`
	Term        int      `json:"term"`
	CommitIndex int      `json:"commit_index"`
	LastApplied int      `json:"last_applied"`
	LogLength   int      `json:"log_length"`
	LeaderID    int      `json:"leader_id"`
	Peers       []int    `json:"peers"`
}

func (rn *RaftNode) emitEvent(eventType string, data map[string]interface{}) {
	rn.eventMu.RLock()
	defer rn.eventMu.RUnlock()

	if rn.eventBus == nil {
		return
	}

	event := RaftEvent{
		Type:      eventType,
		NodeID:    rn.id,
		Term:      rn.currentTerm,
		Data:      data,
		Timestamp: time.Now(),
	}

	select {
	case rn.eventBus <- event:
	default:
	}
}

func (rn *RaftNode) GetEventBus() <-chan RaftEvent {
	return rn.eventBus
}

func (rn *RaftNode) GetClusterStatus() ClusterNodeStatus {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	peers := make([]int, len(rn.peers))
	copy(peers, rn.peers)

	return ClusterNodeStatus{
		ID:          rn.id,
		State:       rn.state.String(),
		Term:        rn.currentTerm,
		CommitIndex: rn.commitIndex,
		LastApplied: rn.lastApplied,
		LogLength:   rn.getLastLogIndex(),
		LeaderID:    rn.leaderId,
		Peers:       peers,
	}
}
