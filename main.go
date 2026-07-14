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

	nodes := []*raft.RaftNode{
		raft.NewRaftNode(0, []int{1, 2}, ports),
		raft.NewRaftNode(1, []int{0, 2}, ports),
		raft.NewRaftNode(2, []int{0, 1}, ports),
	}

	for i, node := range nodes {
		if err := node.StartServer(); err != nil {
			fmt.Printf("error starting server %d: %v\n", i, err)
			return
		}
		defer node.StopServer()
	}

	time.Sleep(500 * time.Millisecond)

	var leader *raft.RaftNode
	for _, node := range nodes {
		state, _ := node.GetState()
		if state == raft.Leader {
			leader = node
			break
		}
	}

	if leader != nil {
		fmt.Println("[Test] Leader found, proposing command...")
		leader.Propose("SET x = 42")
	} else {
		fmt.Println("[Test] No leader found in time.")
	}

	time.Sleep(500 * time.Millisecond)
}