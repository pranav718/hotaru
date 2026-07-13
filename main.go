package main

import (
	"fmt"

	"github.com/pranav718/hotaru/raft"
)

func main() {
	//creating 3 node cluster with node 0's peer being [1,2] and so on

	node0 := raft.NewRaftNode(0, []int{1,2})
	node1 := raft.NewRaftNode(1, []int{0,2})
	node2 := raft.NewRaftNode(2, []int{0,1}) 

	for i, node := range []*raft.RaftNode{node0, node1, node2} {
		state, term := node.GetState()
		fmt.Printf("[Test] node %d -> state: %s, term: %d\n", i, state, term)
	}

}