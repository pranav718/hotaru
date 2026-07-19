package raft

import (
	"encoding/json"
	"fmt"
	"os"
)

type RaftState struct {
	CurrentTerm int
	VotedFor    int
	Log         []LogEntry
	LastIncludedIndex int
	LastIncludedTerm  int
}

func (rn *RaftNode) TestSetLogAndTerm(term int, log []LogEntry) {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	rn.currentTerm = term
	rn.log = log
	rn.persist()
}

func (rn *RaftNode) persist() {
	state := RaftState{
		CurrentTerm: rn.currentTerm,
		VotedFor:    rn.votedFor,
		Log:         rn.log,
		LastIncludedIndex: rn.lastIncludedIndex,
		LastIncludedTerm:  rn.lastIncludedTerm,
	}
	data, err := json.Marshal(state)
	if err != nil {
		fmt.Printf("[Node %d] Error marshaling state: %v\n", rn.id, err)
		return
	}
	filename := fmt.Sprintf("raft_state_%d.json", rn.id)
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		fmt.Printf("[Node %d] Error writing state file: %v\n", rn.id, err)
	}
}

func (rn *RaftNode) readPersist() {
	filename := fmt.Sprintf("raft_state_%d.json", rn.id)
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		fmt.Printf("[Node %d] Error reading state file: %v\n", rn.id, err)
		return
	}
	var state RaftState
	err = json.Unmarshal(data, &state)
	if err != nil {
		fmt.Printf("[Node %d] Error unmarshaling state: %v\n", rn.id, err)
		return
	}
	rn.currentTerm = state.CurrentTerm
	rn.votedFor = state.VotedFor
	rn.log = state.Log
	rn.lastIncludedIndex = state.LastIncludedIndex
	rn.lastIncludedTerm = state.LastIncludedTerm
	fmt.Printf("[Node %d] Loaded persisted state: Term %d, VotedFor %d, Log entries: %d, LastIncludedIndex: %d\n", rn.id, rn.currentTerm, rn.votedFor, len(rn.log), rn.lastIncludedIndex)

	snapshotFilename := fmt.Sprintf("raft_snapshot_%d.bin", rn.id)
	if snapshotData, err := os.ReadFile(snapshotFilename); err == nil {
		if err := rn.kvStore.Restore(snapshotData); err != nil {
			fmt.Printf("[Node %d] Error restoring snapshot: %v\n", rn.id, err)
		} else {
			fmt.Printf("[Node %d] Restored snapshot up to index %d\n", rn.id, rn.lastIncludedIndex)
		}
	}
}
