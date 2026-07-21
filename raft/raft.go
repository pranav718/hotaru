package raft

import (
	"fmt"
	"net"
	"sync"
	"time"
)

//nodestate represents 3 poissible states of a raft node
type NodeState int

const (
	Follower NodeState = iota
	Candidate
	Leader
)

func (s NodeState) String() string {

	switch s {
	case Follower:
		return "Follower"
	case Candidate:
		return "Candidate"
	case Leader:
		return "Leader"
	default:
		return "Unknown"
	}

}

type RequestVoteArgs struct {
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

type AppendEntriesArgs struct {
	Term         int
	LeaderId     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

type InstallSnapshotArgs struct {
	Term              int
	LeaderId          int
	LastIncludedIndex int
	LastIncludedTerm  int
	Data              []byte
}

type InstallSnapshotReply struct {
	Term int
}

type RaftNode struct {
	mu sync.Mutex //persistent state
	id int 
	currentTerm int
	votedFor int //candidateID that received vote and -1 if none
	log []LogEntry 
	state NodeState
	commitIndex int //index of highest log entrt known to be commited
	lastApplied int //index of highest log entry applied to state machine
	peers []int
	peerPorts map[int]string
	listener net.Listener
	lastContact     time.Time
	electionTimeout time.Duration
	nextIndex       map[int]int
	matchIndex      map[int]int
	killed          bool
	kvStore         *KVStore
	httpServer      *HTTPServer
	leaderId        int
	lastIncludedIndex int
	lastIncludedTerm  int
}

type LogEntry struct {
	Term int 
	Index int //position in log, 1-indexed
	Command string
}

//for new raft node, every node starts as a follower state
func NewRaftNode(id int, peers []int, ports map[int]string) *RaftNode {
	node := &RaftNode{
		id: id,
		currentTerm: 0,
		votedFor: -1,
		log: make([]LogEntry, 0),
		state: Follower,
		commitIndex: 0,
		lastApplied: 0,
		peers: peers,
		peerPorts: ports,
		kvStore: NewKVStore(),
		leaderId: -1,
		lastIncludedIndex: 0,
		lastIncludedTerm:  0,
	}
	node.httpServer = NewHTTPServer(node, httpAddrFromRPC(ports[id]))
	node.resetElectionTimeout()
	node.lastContact = time.Now()
	node.readPersist()
	go node.runElectionTimer()

	fmt.Printf("[node %d] created. state: %s, term: %d\n", node.id, node.state, node.currentTerm)
	return node
}

func (rn *RaftNode) GetId() int {
	return rn.id
}

func (rn *RaftNode) GetState() (NodeState, int) {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	return rn.state, rn.currentTerm
}

func (rn *RaftNode) GetLastApplied() int {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	return rn.lastApplied
}

func (rn *RaftNode) becomeFollower(term int) {
	oldState := rn.state
	oldTerm := rn.currentTerm
	rn.state = Follower
	rn.currentTerm = term
	rn.votedFor = -1
	rn.leaderId = -1
	rn.lastContact = time.Now()
	rn.persist()
	fmt.Printf("[Node %d] %s (term %d) → Follower (term %d)\n", rn.id, oldState, oldTerm, term)
}

func (rn *RaftNode) becomeCandidate() {
	rn.state = Candidate
	rn.currentTerm++ 
	rn.votedFor = rn.id 
	rn.leaderId = -1
	rn.persist()
	fmt.Printf("[Node %d] → Candidate (term %d)\n", rn.id, rn.currentTerm)
}

func (rn *RaftNode) becomeLeader() {
	rn.state = Leader
	rn.leaderId = rn.id
	rn.nextIndex = make(map[int]int)
	rn.matchIndex = make(map[int]int)
	for _, peerId := range rn.peers {
		rn.nextIndex[peerId] = rn.getLastLogIndex() + 1
		rn.matchIndex[peerId] = 0
	}
	fmt.Printf("[Node %d] → Leader (term %d)\n", rn.id, rn.currentTerm)
	go rn.runHeartbeatLoop()
}

func (rn *RaftNode) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) error {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	// 1. Term check: Candidate term must be >= ours
	if args.Term < rn.currentTerm {
		reply.Term = rn.currentTerm
		reply.VoteGranted = false
		return nil
	}

	// 2. Discover higher term: Step down immediately
	if args.Term > rn.currentTerm {
		rn.becomeFollower(args.Term)
	}

	upToDate := false
	myLastTerm := rn.getLastLogTerm()
	myLastIndex := rn.getLastLogIndex()

	if args.LastLogTerm > myLastTerm {
		upToDate = true
	} else if args.LastLogTerm == myLastTerm && args.LastLogIndex >= myLastIndex {
		upToDate = true
	}

	//3.grant vote if we havent voted for anyone or voted for this candidate
	if (rn.votedFor == -1 || rn.votedFor == args.CandidateId) && upToDate {
		rn.votedFor = args.CandidateId
		rn.lastContact = time.Now() //4.election timeout reset on granting vote
		rn.persist()
		reply.VoteGranted = true
		fmt.Printf("[Node %d] Granted vote to Candidate %d in Term %d\n", rn.id, args.CandidateId, rn.currentTerm)
	} else {
		reply.VoteGranted = false
		fmt.Printf("[Node %d] Denied vote to Candidate %d in Term %d (already voted for %d)\n", rn.id, args.CandidateId, rn.currentTerm, rn.votedFor)
	}

	reply.Term = rn.currentTerm
	return nil
}

func (rn *RaftNode) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) error {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if args.Term < rn.currentTerm {
		reply.Term = rn.currentTerm
		reply.Success = false
		return nil
	}

	if args.Term > rn.currentTerm || (args.Term == rn.currentTerm && rn.state == Candidate) {
		rn.becomeFollower(args.Term)
	}
	rn.leaderId = args.LeaderId

	//consistency check to check if we have matching entry at PrevLogIndex
	if args.PrevLogIndex > 0 {
		if rn.getLastLogIndex() < args.PrevLogIndex {
			reply.Term = rn.currentTerm
			reply.Success = false
			return nil
		}
		if rn.getLogTerm(args.PrevLogIndex) != args.PrevLogTerm {
			reply.Term = rn.currentTerm
			reply.Success = false
			return nil
		}
	}

	if len(args.Entries) == 0 {
		fmt.Printf("[Node %d] Received Heartbeat (AppendEntries) from Leader %d in Term %d\n", rn.id, args.LeaderId, args.Term)
	} else {
		fmt.Printf("[Node %d] Received Replication (AppendEntries) from Leader %d with %d entries in Term %d\n", rn.id, args.LeaderId, len(args.Entries), args.Term)
		for _, entry := range args.Entries {
			if entry.Index <= rn.getLastLogIndex() {
				if rn.getLogTerm(entry.Index) != entry.Term {
					rn.log = rn.log[:entry.Index-rn.lastIncludedIndex-1]
					rn.log = append(rn.log, entry)
					fmt.Printf("[Node %d] Appended entry locally (overwriting conflict): Index %d, Term %d, Command '%s'\n", rn.id, entry.Index, entry.Term, entry.Command)
				}
			} else {
				rn.log = append(rn.log, entry)
				fmt.Printf("[Node %d] Appended entry locally: Index %d, Term %d, Command '%s'\n", rn.id, entry.Index, entry.Term, entry.Command)
			}
		}
	}

	if args.LeaderCommit > rn.commitIndex {
		rn.commitIndex = args.LeaderCommit
		if rn.getLastLogIndex() < rn.commitIndex {
			rn.commitIndex = rn.getLastLogIndex()
		}
		rn.applyLogs()
	}

	rn.lastContact = time.Now()
	reply.Term = rn.currentTerm
	reply.Success = true
	rn.persist()
	return nil
}

func (rn *RaftNode) InstallSnapshot(args *InstallSnapshotArgs, reply *InstallSnapshotReply) error {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if args.Term < rn.currentTerm {
		reply.Term = rn.currentTerm
		return nil
	}

	if args.Term > rn.currentTerm {
		rn.becomeFollower(args.Term)
	}
	rn.leaderId = args.LeaderId

	if args.LastIncludedIndex <= rn.lastIncludedIndex {
		reply.Term = rn.currentTerm
		return nil
	}

	fmt.Printf("[Node %d] Installing snapshot from Leader %d up to index %d\n", rn.id, args.LeaderId, args.LastIncludedIndex)

	if err := rn.kvStore.Restore(args.Data); err != nil {
		fmt.Printf("[Node %d] Error restoring snapshot: %v\n", rn.id, err)
		return err
	}

	if err := rn.SaveSnapshot(args.Data); err != nil {
		fmt.Printf("[Node %d] Error saving snapshot: %v\n", rn.id, err)
		return err
	}

	if args.LastIncludedIndex <= rn.getLastLogIndex() && rn.getLogTerm(args.LastIncludedIndex) == args.LastIncludedTerm {
		rn.log = rn.getLogSlice(args.LastIncludedIndex + 1)
	} else {
		rn.log = make([]LogEntry, 0)
	}

	rn.lastIncludedIndex = args.LastIncludedIndex
	rn.lastIncludedTerm = args.LastIncludedTerm

	if args.LastIncludedIndex > rn.commitIndex {
		rn.commitIndex = args.LastIncludedIndex
	}
	if args.LastIncludedIndex > rn.lastApplied {
		rn.lastApplied = args.LastIncludedIndex
	}

	rn.persist()
	rn.lastContact = time.Now()
	reply.Term = rn.currentTerm
	return nil
}

func (rn *RaftNode) GetLeaderId() int {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	return rn.leaderId
}

func (rn *RaftNode) GetLeaderHTTPAddr() string {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	if rn.leaderId == -1 {
		return ""
	}
	rpcAddr, ok := rn.peerPorts[rn.leaderId]
	if !ok {
		return ""
	}
	return httpAddrFromRPC(rpcAddr)
}

func (rn *RaftNode) getLastLogIndex() int {
	return rn.lastIncludedIndex + len(rn.log)
}

func (rn *RaftNode) getLastLogTerm() int {
	if len(rn.log) == 0 {
		return rn.lastIncludedTerm
	}
	return rn.log[len(rn.log)-1].Term
}

func (rn *RaftNode) getLogTerm(index int) int {
	if index == rn.lastIncludedIndex {
		return rn.lastIncludedTerm
	}
	if index > rn.lastIncludedIndex && index <= rn.getLastLogIndex() {
		return rn.log[index-rn.lastIncludedIndex-1].Term
	}
	return 0
}

func (rn *RaftNode) getLogEntry(index int) LogEntry {
	return rn.log[index-rn.lastIncludedIndex-1]
}

func (rn *RaftNode) getLogSlice(fromIndex int) []LogEntry {
	if fromIndex <= rn.lastIncludedIndex {
		return rn.log
	}
	return rn.log[fromIndex-rn.lastIncludedIndex-1:]
}

func (rn *RaftNode) TakeSnapshot(index int) error {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if index <= rn.lastIncludedIndex {
		return fmt.Errorf("snapshot index %d is already included in a previous snapshot (up to %d)", index, rn.lastIncludedIndex)
	}

	if index > rn.lastApplied {
		return fmt.Errorf("cannot snapshot unapplied entries (index %d > lastApplied %d)", index, rn.lastApplied)
	}

	snapshotData, err := rn.kvStore.Snapshot()
	if err != nil {
		return fmt.Errorf("failed to take state machine snapshot: %v", err)
	}

	lastIncludedTerm := rn.getLogTerm(index)

	// trim the log array to keep only the entries after 'index'
	rn.log = rn.getLogSlice(index + 1)
	rn.lastIncludedIndex = index
	rn.lastIncludedTerm = lastIncludedTerm

	rn.persist()

	if err := rn.SaveSnapshot(snapshotData); err != nil {
		return err
	}

	fmt.Printf("[Node %d] Snapshot created up to index %d (term %d), log trimmed, %d entries remaining\n", rn.id, index, lastIncludedTerm, len(rn.log))
	return nil
}