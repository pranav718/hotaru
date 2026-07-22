package raft

import (
	"fmt"
	"time"
)

func (rn *RaftNode) broadcastAppendEntries() {
	rn.mu.Lock()
	if rn.state != Leader {
		rn.mu.Unlock()
		return
	}
	term := rn.currentTerm
	myId := rn.id
	peers := rn.peers
	rn.mu.Unlock()

	for _, peerId := range peers {
		go func(pid int) {
			rn.mu.Lock()
			if rn.state != Leader || rn.currentTerm != term {
				rn.mu.Unlock()
				return
			}

			next := rn.nextIndex[pid]

			if next <= rn.lastIncludedIndex {
				snapshotData, err := rn.ReadSnapshot()
				if err != nil {
					fmt.Printf("[Node %d] Error reading snapshot for peer %d: %v\n", rn.id, pid, err)
					rn.mu.Unlock()
					return
				}

				args := InstallSnapshotArgs{
					Term:              term,
					LeaderId:          myId,
					LastIncludedIndex: rn.lastIncludedIndex,
					LastIncludedTerm:  rn.lastIncludedTerm,
					Data:              snapshotData,
				}
				rn.mu.Unlock()

				var reply InstallSnapshotReply
				ok := rn.sendInstallSnapshot(pid, &args, &reply)
				if !ok {
					return
				}

				rn.mu.Lock()
				defer rn.mu.Unlock()
				if rn.currentTerm != term || rn.state != Leader {
					return
				}

				if reply.Term > rn.currentTerm {
					rn.becomeFollower(reply.Term)
					return
				}

				rn.nextIndex[pid] = args.LastIncludedIndex + 1
				rn.matchIndex[pid] = args.LastIncludedIndex
				rn.updateLeaderCommit()
				return
			}

			var entries []LogEntry
			if rn.getLastLogIndex() >= next {
				entries = rn.getLogSlice(next)
			}

			prevLogIndex := next - 1
			prevLogTerm := 0
			if prevLogIndex > 0 {
				prevLogTerm = rn.getLogTerm(prevLogIndex)
			}

			args := AppendEntriesArgs{
				Term:         term,
				LeaderId:     myId,
				PrevLogIndex: prevLogIndex,
				PrevLogTerm:  prevLogTerm,
				Entries:      entries,
				LeaderCommit: rn.commitIndex,
			}
			rn.mu.Unlock()

			var reply AppendEntriesReply
			ok := rn.sendAppendEntries(pid, &args, &reply)
			if !ok {
				return
			}

			rn.mu.Lock()
			defer rn.mu.Unlock()
			if rn.currentTerm != term || rn.state != Leader {
				return
			}

			if reply.Term > rn.currentTerm {
				rn.becomeFollower(reply.Term)
				return
			}

			if reply.Success {
				rn.nextIndex[pid] = next + len(entries)
				rn.matchIndex[pid] = rn.nextIndex[pid] - 1
				rn.updateLeaderCommit()
			} else {
				rn.nextIndex[pid] = rn.nextIndex[pid] - 1
				if rn.nextIndex[pid] < 1 {
					rn.nextIndex[pid] = 1
				}
			}
		}(peerId)
	}
}

func (rn *RaftNode) runHeartbeatLoop() {
	rn.broadcastAppendEntries()
	for {
		time.Sleep(50 * time.Millisecond)
		rn.mu.Lock()
		if rn.killed || rn.state != Leader {
			rn.mu.Unlock()
			return
		}
		rn.mu.Unlock()
		rn.broadcastAppendEntries()
	}
}

func (rn *RaftNode) Propose(command string) (int, bool) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if rn.state != Leader {
		return 0, false
	}

	entry := LogEntry{
		Term:    rn.currentTerm,
		Index:   rn.getLastLogIndex() + 1,
		Command: command,
	}
	rn.log = append(rn.log, entry)
	rn.persist()
	fmt.Printf("[Node %d] Leader appended entry locally: Index %d, Term %d, Command '%s'\n", rn.id, entry.Index, entry.Term, entry.Command)
	return entry.Index, true
}

func (rn *RaftNode) ProposeAndCommit(command string) (string, error) {
	index, isLeader := rn.Propose(command)
	if !isLeader {
		return "", fmt.Errorf("not leader")
	}

	start := time.Now()
	for {
		rn.mu.Lock()
		killed := rn.killed
		applied := rn.lastApplied >= index
		var termMatches bool
		if applied && index <= rn.getLastLogIndex() {
			termMatches = rn.getLogTerm(index) == rn.currentTerm
		}
		rn.mu.Unlock()

		if killed {
			return "", fmt.Errorf("server stopped")
		}

		if applied {
			if !termMatches {
				return "", fmt.Errorf("command overwritten by a newer leader")
			}
			break
		}

		if time.Since(start) > 2*time.Second {
			return "", fmt.Errorf("timeout waiting for command to commit")
		}
		time.Sleep(10 * time.Millisecond)
	}

	return "OK", nil
}

func (rn *RaftNode) applyLogs() {
	for rn.commitIndex > rn.lastApplied {
		rn.lastApplied++
		entry := rn.getLogEntry(rn.lastApplied)
		result := rn.kvStore.Apply(entry.Command)
		fmt.Printf("[Node %d] Applied entry locally: Index %d, Term %d, Command '%s' -> Result: '%s'\n", rn.id, entry.Index, entry.Term, entry.Command, result)

		if entry.Type == EntryRemoveNode && entry.TargetID == rn.id && rn.state == Leader {
			fmt.Printf("[Node %d] Leader committed self-removal entry. Stepping down to Follower.\n", rn.id)
			rn.becomeFollower(rn.currentTerm)
		}
	}
}

func (rn *RaftNode) updateLeaderCommit() {
	for n := rn.getLastLogIndex(); n > rn.commitIndex; n-- {
		if rn.getLogTerm(n) != rn.currentTerm {
			continue
		}

		count := 1
		for _, peerId := range rn.peers {
			if rn.matchIndex[peerId] >= n {
				count++
			}
		}

		if count > (len(rn.peers)+1)/2 {
			rn.commitIndex = n
			rn.applyLogs()
			break
		}
	}
}

func (rn *RaftNode) VerifyLeadership() bool {
	rn.mu.Lock()
	if rn.state != Leader {
		rn.mu.Unlock()
		return false
	}
	term := rn.currentTerm
	peers := make([]int, len(rn.peers))
	copy(peers, rn.peers)
	rn.mu.Unlock()

	done := make(chan bool, len(peers))
	for _, peerId := range peers {
		go func(pid int) {
			rn.mu.Lock()
			args := AppendEntriesArgs{
				Term:         term,
				LeaderId:     rn.id,
				PrevLogIndex: 0,
				PrevLogTerm:  0,
				Entries:      []LogEntry{},
				LeaderCommit: 0,
			}
			rn.mu.Unlock()
			var reply AppendEntriesReply
			ok := rn.sendAppendEntries(pid, &args, &reply)
			done <- ok && reply.Success
		}(peerId)
	}

	responses := 1
	timeout := time.After(100 * time.Millisecond)
Loop:
	for i := 0; i < len(peers); i++ {
		select {
		case success := <-done:
			if success {
				responses++
			}
		case <-timeout:
			break Loop
		}
	}

	return responses > (len(peers)+1)/2
}

func (rn *RaftNode) QueryKey(key string) string {
	return rn.kvStore.Apply(fmt.Sprintf("GET %s", key))
}

func (rn *RaftNode) ProposeAddNode(peerID int, rpcAddr string) (int, bool) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if rn.state != Leader {
		return 0, false
	}

	if rn.hasUncommittedConfigChange() {
		fmt.Printf("[Node %d] Cannot propose AddNode %d: another configuration change is uncommitted\n", rn.id, peerID)
		return 0, false
	}

	entry := LogEntry{
		Term:      rn.currentTerm,
		Index:     rn.getLastLogIndex() + 1,
		Command:   fmt.Sprintf("ADD_NODE %d %s", peerID, rpcAddr),
		Type:      EntryAddNode,
		TargetID:  peerID,
		TargetRPC: rpcAddr,
	}
	rn.log = append(rn.log, entry)
	rn.addPeer(peerID, rpcAddr)
	rn.persist()
	fmt.Printf("[Node %d] Leader appended AddNode entry locally: Index %d, Peer %d (%s)\n", rn.id, entry.Index, peerID, rpcAddr)
	return entry.Index, true
}

func (rn *RaftNode) ProposeRemoveNode(peerID int) (int, bool) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if rn.state != Leader {
		return 0, false
	}

	if rn.hasUncommittedConfigChange() {
		fmt.Printf("[Node %d] Cannot propose RemoveNode %d: another configuration change is uncommitted\n", rn.id, peerID)
		return 0, false
	}

	entry := LogEntry{
		Term:     rn.currentTerm,
		Index:    rn.getLastLogIndex() + 1,
		Command:  fmt.Sprintf("REMOVE_NODE %d", peerID),
		Type:     EntryRemoveNode,
		TargetID: peerID,
	}
	rn.log = append(rn.log, entry)
	rn.removePeer(peerID)
	rn.persist()
	fmt.Printf("[Node %d] Leader appended RemoveNode entry locally: Index %d, Peer %d\n", rn.id, entry.Index, peerID)
	return entry.Index, true
}
