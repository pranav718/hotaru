package raft

import (
	"fmt"
	"net"
	"net/rpc"
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

func (rn *RaftNode) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) error {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	fmt.Printf("[Node %d] Received RequestVote from Candidate %d in Term %d\n", rn.id, args.CandidateId, args.Term)
	reply.Term = rn.currentTerm
	reply.VoteGranted = false
	return nil
}

func (rn *RaftNode) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) error {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	if len(args.Entries) == 0 {
		fmt.Printf("[Node %d] Received Heartbeat (AppendEntries) from Leader %d in Term %d\n", rn.id, args.LeaderId, args.Term)
	} else {
		fmt.Printf("[Node %d] Received Replication (AppendEntries) from Leader %d with %d entries in Term %d\n", rn.id, args.LeaderId, len(args.Entries), args.Term)
	}
	reply.Term = rn.currentTerm
	reply.Success = true
	return nil
}

func (rn *RaftNode) StartServer() error {
	myPort, ok := rn.peerPorts[rn.id]
	if !ok {
		return fmt.Errorf("node %d has no port assigned in peerPorts configuration", rn.id)
	}
	rpcServer := rpc.NewServer()
	err := rpcServer.Register(rn)
	if err != nil {
		return fmt.Errorf("registering RPC service: %v", err)
	}
	l, err := net.Listen("tcp", myPort)
	if err != nil {
		return fmt.Errorf("listening on port %s: %v", myPort, err)
	}
	rn.listener = l
	go func() {
		for {
			conn, err := rn.listener.Accept()
			if err != nil {
				return
			}
			go rpcServer.ServeConn(conn)
		}
	}()
	fmt.Printf("[Node %d] RPC server listening on port %s\n", rn.id, myPort)
	return nil
}

func (rn *RaftNode) StopServer() {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	if rn.listener != nil {
		rn.listener.Close()
	}
}

func (rn *RaftNode) sendRequestVote(peerId int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	port, ok := rn.peerPorts[peerId]
	if !ok {
		return false
	}
	client, err := rpc.Dial("tcp", port)
	if err != nil {
		fmt.Printf("[Node %d] Error dialing peer %d at port %s: %v\n", rn.id, peerId, port, err)
		return false
	}
	defer client.Close()
	err = client.Call("RaftNode.RequestVote", args, reply)
	if err != nil {
		fmt.Printf("[Node %d] RPC Error calling RequestVote on peer %d: %v\n", rn.id, peerId, err)
		return false
	}
	return true
}

func (rn *RaftNode) sendAppendEntries(peerId int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	port, ok := rn.peerPorts[peerId]
	if !ok {
		return false
	}
	client, err := rpc.Dial("tcp", port)
	if err != nil {
		fmt.Printf("[Node %d] Error dialing peer %d at port %s: %v\n", rn.id, peerId, port, err)
		return false
	}
	defer client.Close()
	err = client.Call("RaftNode.AppendEntries", args, reply)
	if err != nil {
		fmt.Printf("[Node %d] RPC Error calling AppendEntries on peer %d: %v\n", rn.id, peerId, err)
		return false
	}
	return true
}

func (rn *RaftNode) TestTriggerSendRPC(peerId int, isAppendEntries bool) {
	if isAppendEntries {
		args := &AppendEntriesArgs{
			Term:     rn.currentTerm,
			LeaderId: rn.id,
		}
		var reply AppendEntriesReply
		fmt.Printf("[Node %d] Sending stub AppendEntries to peer %d\n", rn.id, peerId)
		ok := rn.sendAppendEntries(peerId, args, &reply)
		fmt.Printf("[Node %d] Peer %d AppendEntries response: status=%t, replyTerm=%d, success=%t\n", rn.id, peerId, ok, reply.Term, reply.Success)
	} else {
		args := &RequestVoteArgs{
			Term:        rn.currentTerm,
			CandidateId: rn.id,
		}
		var reply RequestVoteReply
		fmt.Printf("[Node %d] Sending stub RequestVote to peer %d\n", rn.id, peerId)
		ok := rn.sendRequestVote(peerId, args, &reply)
		fmt.Printf("[Node %d] Peer %d RequestVote response: status=%t, replyTerm=%d, voteGranted=%t\n", rn.id, peerId, ok, reply.Term, reply.VoteGranted)
	}
}
