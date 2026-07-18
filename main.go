package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/pranav718/hotaru/raft"
)

func httpDo(method, url string) (string, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	return string(body), nil
}

func httpPortFor(id int) int {
	return 8010 + id
}

func httpSet(port int, key, value string) error {
	url := fmt.Sprintf("http://127.0.0.1:%d/set?key=%s&value=%s", port, key, value)
	res, err := httpDo("POST", url)
	if err != nil {
		return fmt.Errorf("SET %s=%s on port %d: %v", key, value, port, err)
	}
	fmt.Printf("[HTTP] SET %s=%s on :%d → %s\n", key, value, port, res)
	return nil
}

func httpGet(port int, key string) (string, error) {
	url := fmt.Sprintf("http://127.0.0.1:%d/get?key=%s", port, key)
	res, err := httpDo("GET", url)
	if err != nil {
		return "", fmt.Errorf("GET %s on port %d: %v", key, port, err)
	}
	return res, nil
}

func httpDel(port int, key string) error {
	url := fmt.Sprintf("http://127.0.0.1:%d/del?key=%s", port, key)
	res, err := httpDo("DELETE", url)
	if err != nil {
		return fmt.Errorf("DEL %s on port %d: %v", key, port, err)
	}
	fmt.Printf("[HTTP] DEL %s on :%d → %s\n", key, port, res)
	return nil
}

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

	fmt.Println("=== Test 1: Write via leader HTTP API ===")
	if err := httpSet(httpPortFor(leader.GetId()), "x", "100"); err != nil {
		fmt.Printf("FAIL: %v\n", err)
		return
	}
	if err := httpSet(httpPortFor(leader.GetId()), "y", "200"); err != nil {
		fmt.Printf("FAIL: %v\n", err)
		return
	}
	time.Sleep(400 * time.Millisecond)

	fmt.Println("\n=== Test 2: Read via HTTP on all nodes (linearizable) ===")
	for i := 0; i < 3; i++ {
		val, err := httpGet(httpPortFor(i), "x")
		if err != nil {
			fmt.Printf("FAIL: node %d: %v\n", i, err)
		} else {
			fmt.Printf("[Test] Node %d: key 'x' = %q\n", i, val)
		}
		val, err = httpGet(httpPortFor(i), "y")
		if err != nil {
			fmt.Printf("FAIL: node %d: %v\n", i, err)
		} else {
			fmt.Printf("[Test] Node %d: key 'y' = %q\n", i, val)
		}
	}

	fmt.Println("\n=== Test 3: Write to follower via HTTP (expect proxy) ===")
	var followerPort int
	for i := 0; i < 3; i++ {
		if nodes[i] != leader {
			followerPort = httpPortFor(i)
			break
		}
	}
	if err := httpSet(followerPort, "z", "300"); err != nil {
		fmt.Printf("FAIL: write to follower proxied: %v\n", err)
		return
	}
	time.Sleep(200 * time.Millisecond)

	val, err := httpGet(httpPortFor(leader.GetId()), "z")
	if err != nil {
		fmt.Printf("FAIL: read z from leader: %v\n", err)
	} else {
		fmt.Printf("[Test] Key 'z' after follower proxy write = %q\n", val)
	}

	fmt.Println("\n=== Test 4: Delete via HTTP ===")
	if err := httpDel(httpPortFor(leader.GetId()), "y"); err != nil {
		fmt.Printf("FAIL: delete via leader: %v\n", err)
		return
	}
	time.Sleep(200 * time.Millisecond)

	for i := 0; i < 3; i++ {
		val, err := httpGet(httpPortFor(i), "y")
		if err != nil {
			fmt.Printf("[Test] Node %d: key 'y' deleted — %v\n", i, err)
		} else {
			fmt.Printf("[Test] Node %d: key 'y' = %q\n", i, val)
		}
	}

	fmt.Println("\n=== Test 5: Crash-recovery via HTTP ===")
	fmt.Println("[Test] Killing Node 2...")
	node2.StopServer()
	time.Sleep(300 * time.Millisecond)

	var newLeader *raft.RaftNode
	for _, node := range nodes {
		if node == node2 {
			continue
		}
		state, _ := node.GetState()
		if state == raft.Leader {
			newLeader = node
			break
		}
	}
	if newLeader == nil {
		fmt.Println("[Test] No new leader found after killing Node 2.")
		return
	}
	fmt.Printf("[Test] New leader is Node %d\n", newLeader.GetId())

	fmt.Println("[Test] Writing SET x 500 while Node 2 is offline...")
	if err := httpSet(httpPortFor(newLeader.GetId()), "x", "500"); err != nil {
		fmt.Printf("FAIL: %v\n", err)
		return
	}
	time.Sleep(500 * time.Millisecond)

	fmt.Println("[Test] Restarting Node 2...")
	newNode2 := raft.NewRaftNode(2, []int{0, 1}, ports)
	if err := newNode2.StartServer(); err != nil {
		fmt.Printf("error restarting server 2: %v\n", err)
		return
	}
	defer newNode2.StopServer()
	time.Sleep(600 * time.Millisecond)

	fmt.Println("[Test] Verifying Node 2 caught up via HTTP...")
	val, err = httpGet(httpPortFor(2), "x")
	if err != nil {
		fmt.Printf("FAIL: %v\n", err)
	} else {
		fmt.Printf("[Test] Restarted Node 2: key 'x' = %q (expected \"500\")\n", val)
	}
	val, err = httpGet(httpPortFor(2), "y")
	if err != nil {
		fmt.Printf("[Test] Restarted Node 2: key 'y' deleted (expected) — %v\n", err)
	} else {
		fmt.Printf("[Test] Restarted Node 2: key 'y' = %q\n", val)
	}

	fmt.Println("\n=== All HTTP integration tests complete ===")
}