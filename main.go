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

	if err := node0.StartServer(); err != nil {
		fmt.Printf("error starting server 0: %v\n", err)
		return
	}
	defer node0.StopServer()

	if err := node1.StartServer(); err != nil {
		fmt.Printf("error starting server 1: %v\n", err)
		return
	}
	defer node1.StopServer()

	if err := node2.StartServer(); err != nil {
		fmt.Printf("error starting server 2: %v\n", err)
		return
	}
	defer node2.StopServer()

	time.Sleep(1 * time.Second)
}