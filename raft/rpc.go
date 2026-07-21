package raft

import (
	"fmt"
	"net"
	"net/rpc"
)

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

func (rn *RaftNode) sendInstallSnapshot(peerId int, args *InstallSnapshotArgs, reply *InstallSnapshotReply) bool {
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
	err = client.Call("RaftNode.InstallSnapshot", args, reply)
	if err != nil {
		fmt.Printf("[Node %d] RPC Error calling InstallSnapshot on peer %d: %v\n", rn.id, peerId, err)
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
