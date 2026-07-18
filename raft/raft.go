package raft

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"os"
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
		rn.nextIndex[peerId] = len(rn.log) + 1
		rn.matchIndex[peerId] = 0
	}
	fmt.Printf("[Node %d] → Leader (term %d)\n", rn.id, rn.currentTerm)
	go rn.runHeartbeatLoop()
}

func (rn *RaftNode) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) error {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	//1.candidate's term must be >= ours
	if args.Term < rn.currentTerm {
		reply.Term = rn.currentTerm
		reply.VoteGranted = false
		return nil
	}

	//2.step down if higher term 
	if args.Term > rn.currentTerm {
		rn.becomeFollower(args.Term)
	}

	//3.grant vote if we havent voted for anyone or voted for this candidate
	if rn.votedFor == -1 || rn.votedFor == args.CandidateId {
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
		if len(rn.log) < args.PrevLogIndex {
			reply.Term = rn.currentTerm
			reply.Success = false
			return nil
		}
		if rn.log[args.PrevLogIndex-1].Term != args.PrevLogTerm {
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
			if entry.Index <= len(rn.log) {
				if rn.log[entry.Index-1].Term != entry.Term {
					rn.log = rn.log[:entry.Index-1]
					rn.log = append(rn.log, entry)
					fmt.Printf("[Node %d] Appended entry locally (overwriting conflict): Index %d, Term %d, Command '%s'\n", rn.id, entry.Index, entry.Term, entry.Command)
				}
			} else {
				rn.log = append(rn.log, entry)
				fmt.Printf("[Node %d] Appended entry locally: Index %d, Term %d, Command '%s'\n", rn.id, entry.Index, entry.Term, entry.Command)
			}
		}
	}

	//update commitIndex if leader has committed new entries
	if args.LeaderCommit > rn.commitIndex {
		rn.commitIndex = args.LeaderCommit
		if len(rn.log) < rn.commitIndex {
			rn.commitIndex = len(rn.log)
		}
		rn.applyLogs()
	}

	rn.lastContact = time.Now()
	reply.Term = rn.currentTerm
	reply.Success = true
	rn.persist()
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

	if err := rn.httpServer.Start(); err != nil {
		return fmt.Errorf("starting HTTP server: %v", err)
	}
	return nil
}

func (rn *RaftNode) StopServer() {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	rn.killed = true
	if rn.listener != nil {
		rn.listener.Close()
	}
	if rn.httpServer != nil {
		rn.httpServer.Stop()
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

func (rn *RaftNode) resetElectionTimeout() {
	ms := 150 + rand.Intn(150)
	rn.electionTimeout = time.Duration(ms) * time.Millisecond
}

func (rn *RaftNode) runElectionTimer() {
	for {
		time.Sleep(10 * time.Millisecond)
		rn.mu.Lock()
		if rn.killed {
			rn.mu.Unlock()
			return
		}
		if rn.state == Leader {
			rn.mu.Unlock()
			continue
		}
		if time.Since(rn.lastContact) > rn.electionTimeout {
			fmt.Printf("[Node %d] Election timeout expired. Starting election...\n", rn.id)
			rn.becomeCandidate()
			rn.resetElectionTimeout()
			rn.lastContact = time.Now()

			go rn.startElection()
		}
		rn.mu.Unlock()
	}
}

func (rn *RaftNode) startElection() {
	rn.mu.Lock()
	if rn.state != Candidate {
		rn.mu.Unlock()
		return
	}
	term := rn.currentTerm
	myId := rn.id
	peers := rn.peers
	rn.mu.Unlock()

	votesGranted := 1 //self vote
	votesMutex := sync.Mutex{}

	for _, peerId := range peers {
		go func(pid int) {
			args := RequestVoteArgs{
				Term:        term,
				CandidateId: myId,
			}
			var reply RequestVoteReply
			
			ok := rn.sendRequestVote(pid, &args, &reply)
			if !ok {
				return
			}

			rn.mu.Lock()
			defer rn.mu.Unlock()

			//ignore response if no longer candidate or term is over
			if rn.currentTerm != term || rn.state != Candidate {
				return
			}

			//if peer has higher term, step down
			if reply.Term > rn.currentTerm {
				rn.becomeFollower(reply.Term)
				return
			}

			if reply.VoteGranted {
				votesMutex.Lock()
				votesGranted++
				// check for majority (including self)
				if votesGranted > (len(peers)+1)/2 && rn.state == Candidate {
					rn.becomeLeader()
				}
				votesMutex.Unlock()
			}
		}(peerId)
	}
}

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
			var entries []LogEntry
			if len(rn.log) >= next {
				entries = rn.log[next-1:]
			}

			prevLogIndex := next - 1
			prevLogTerm := 0
			if prevLogIndex > 0 {
				prevLogTerm = rn.log[prevLogIndex-1].Term
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
		Index:   len(rn.log) + 1,
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
		if applied && index-1 < len(rn.log) {
			termMatches = rn.log[index-1].Term == rn.currentTerm
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
		entry := rn.log[rn.lastApplied-1]
		result := rn.kvStore.Apply(entry.Command)
		fmt.Printf("[Node %d] Applied entry locally: Index %d, Term %d, Command '%s' -> Result: '%s'\n", rn.id, entry.Index, entry.Term, entry.Command, result)
	}
}

func (rn *RaftNode) updateLeaderCommit() {
	for n := len(rn.log); n > rn.commitIndex; n-- {
		if rn.log[n-1].Term != rn.currentTerm {
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

func (rn *RaftNode) TestSetLogAndTerm(term int, log []LogEntry) {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	rn.currentTerm = term
	rn.log = log
	rn.persist()
}

type RaftState struct {
	CurrentTerm int
	VotedFor    int
	Log         []LogEntry
}

func (rn *RaftNode) persist() {
	state := RaftState{
		CurrentTerm: rn.currentTerm,
		VotedFor:    rn.votedFor,
		Log:         rn.log,
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
	fmt.Printf("[Node %d] Loaded persisted state: Term %d, VotedFor %d, Log entries: %d\n", rn.id, rn.currentTerm, rn.votedFor, len(rn.log))
}

func (rn *RaftNode) QueryKey(key string) string {
	return rn.kvStore.Apply(fmt.Sprintf("GET %s", key))
}

func (rn *RaftNode) GetLeaderId() int {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	return rn.leaderId
}

