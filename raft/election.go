package raft

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

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
