package main

import (
	"fmt"
	"os"
	"time"

	"github.com/pranav718/hotaru/raft"
)

func main() {
	ports := map[int]string{
		0: "127.0.0.1:8000",
		1: "127.0.0.1:8001",
		2: "127.0.0.1:8002",
	}

	for i := 0; i < 3; i++ {
		os.Remove(fmt.Sprintf("raft_state_%d.json", i))
	}

	node0 := raft.NewRaftNode(0, []int{1, 2}, ports)
	node1 := raft.NewRaftNode(1, []int{0, 2}, ports)
	node2 := raft.NewRaftNode(2, []int{0, 1}, ports)

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

	if leader == nil {
		fmt.Println("[Test] No leader found.")
		return
	}

	fmt.Println("[Test] Proposing commands to set initial database values...")
	leader.Propose("SET x 100")
	leader.Propose("SET y 200")
	time.Sleep(400 * time.Millisecond)

	fmt.Printf("[Test] Node 0: key 'x' = %q, key 'y' = %q\n", node0.QueryKey("x"), node0.QueryKey("y"))
	fmt.Printf("[Test] Node 1: key 'x' = %q, key 'y' = %q\n", node1.QueryKey("x"), node1.QueryKey("y"))
	fmt.Printf("[Test] Node 2: key 'x' = %q, key 'y' = %q\n", node2.QueryKey("x"), node2.QueryKey("y"))

	fmt.Println("[Test] Killing Node 2...")
	node2.StopServer()
	time.Sleep(100 * time.Millisecond)

	fmt.Println("[Test] Proposing SET x 500 (value update) while Node 2 is offline...")
	leader.Propose("SET x 500")
	time.Sleep(500 * time.Millisecond)

	fmt.Println("[Test] Restarting Node 2...")
	newNode2 := raft.NewRaftNode(2, []int{0, 1}, ports)
	if err := newNode2.StartServer(); err != nil {
		fmt.Printf("error restarting server 2: %v\n", err)
		return
	}
	defer newNode2.StopServer()

	time.Sleep(600 * time.Millisecond)

	fmt.Println("[Test] Verifying Node 2 has recovered and successfully caught up...")
	fmt.Printf("[Test] Restarted Node 2: key 'x' = %q, key 'y' = %q\n", newNode2.QueryKey("x"), newNode2.QueryKey("y"))
}