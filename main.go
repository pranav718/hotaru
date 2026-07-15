package main

import (
	"fmt"
	"time"

	"github.com/pranav718/hotaru/raft"
)

func main() {
	ports := map[int]string{
		0: "127.0.0.1:8000",
		1: "127.0.0.1:8001",
		2: "127.0.0.1:8002",
	}

	node0 := raft.NewRaftNode(0, []int{1, 2}, ports)
	node1 := raft.NewRaftNode(1, []int{0, 2}, ports)
	node2 := raft.NewRaftNode(2, []int{0, 1}, ports)

	//set up divergent logs before starting servers
	validLog := []raft.LogEntry{
		{Term: 1, Index: 1, Command: "SET a = 1"},
		{Term: 1, Index: 2, Command: "SET b = 2"},
	}
	node0.TestSetLogAndTerm(1, validLog)
	node2.TestSetLogAndTerm(1, validLog)

	divergentLog := []raft.LogEntry{
		{Term: 3, Index: 1, Command: "SET a = 99"},
		{Term: 3, Index: 2, Command: "SET b = 99"},
	}
	node1.TestSetLogAndTerm(1, divergentLog)

	nodes := []*raft.RaftNode{node0, node1, node2}

	for i, node := range nodes {
		if err := node.StartServer(); err != nil {
			fmt.Printf("error starting server %d: %v\n", i, err)
			return
		}
		defer node.StopServer()
	}

	time.Sleep(600 * time.Millisecond)

	var leader *raft.RaftNode
	for _, node := range nodes {
		state, _ := node.GetState()
		if state == raft.Leader {
			leader = node
			break
		}
	}

	if leader != nil {
		fmt.Println("[Test] Proposing new command to leader...")
		leader.Propose("SET c = 3")
	} else {
		fmt.Println("[Test] No leader found in time.")
	}

	time.Sleep(600 * time.Millisecond)
}