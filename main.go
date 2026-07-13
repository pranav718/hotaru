package main

import (
	"fmt"
	"time"

	"github.com/pranav718/hotaru/raft"
)

func main() {
	//creating 3 node cluster with node 0's peer being [1,2] and so on

	node0 := raft.NewRaftNode(0, []int{1,2}, ports)
	node1 := raft.NewRaftNode(1, []int{0,2}, ports)

	err := node0.StartServer()
	if err != nil {
		fmt.Printf("error starting server 0: %v\n", err)
		return
	}
	defer node0.StopServer()

	err = node1.StartServer()
	if err != nil {
		fmt.Printf("error starting server 1: %v\n", err)
		return
	}
	defer node1.StopServer()

	time.Sleep(100 * time.Millisecond)

	node0.TestTriggerSendRPC(1, false)
	node0.TestTriggerSendRPC(1, true)

	time.Sleep(500 * time.Millisecond)
}