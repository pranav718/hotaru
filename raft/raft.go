package raft

import (
	"fmt"
	"sync"
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
}

type LogEntry struct {
	Term int 
	Index int //position in log, 1-indexed
	Command string
}

//for new raft node, every node starts as a follower state
func NewRaftNode(id int, peers []int) *RaftNode {
	node := &RaftNode{
		id: id,
		currentTerm: 0,
		votedFor: -1,
		log: make([]LogEntry, 0),
		state: Follower,
		commitIndex: 0,
		lastApplied: 0,
		peers: peers,
	}

	fmt.Printf("[node %d] created. state: %s, term: %d\n", node.id, node.state, node.currentTerm)
	return node
}

func (rn *RaftNode) GetState() (NodeState, int) {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	return rn.state, rn.currentTerm
}

func (rn *RaftNode) becomeFollower(term int) {
	oldState := rn.state
	oldTerm := rn.currentTerm
	rn.state = Follower
	rn.currentTerm = term
	rn.votedFor = -1 
	fmt.Printf("[Node %d] %s (term %d) → Follower (term %d)\n",
		rn.id, oldState, oldTerm, term)
}

func (rn *RaftNode) becomeCandidate() {
	rn.state = Candidate
	rn.currentTerm++ 
	rn.votedFor = rn.id 
	fmt.Printf("[Node %d] → Candidate (term %d)\n", rn.id, rn.currentTerm)
}

func (rn *RaftNode) becomeLeader() {
	rn.state = Leader
	fmt.Printf("[Node %d] → Leader (term %d)\n", rn.id, rn.currentTerm)
}
