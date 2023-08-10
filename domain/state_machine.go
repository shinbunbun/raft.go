package domain

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/secamp-y3/raft.go/server"
)

type Log struct {
	Log  string
	Term int
}

type StateMachine struct {
	Node           *server.Node
	Log            []Log
	HeartbeatWatch chan int
	Term           int
	Leader         string
	Role           string
	CommitIndex    int
	LastApplied    int
	NextIndex      map[string]int
	MatchIndex     map[string]int
}

type AppendLogsArgs struct {
	Entries []Log
}

type AppendLogsReply int

type AppendEntriesArgs struct {
	Log          []Log
	Term         int
	PrevLogIndex int
	LeaderCommit int
	PrevLogTerm  int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

type RequestVoteArgs struct {
	Term   int
	Leader string
}

type RequestVoteReply struct {
	VoteGranted bool
}

type GetArgs struct{}

type GetReply struct {
	Node        *server.Node
	Log         []Log
	Term        int
	Leader      string
	Role        string
	CommitIndex int
	LastApplied int
	NextIndex   map[string]int
	MatchIndex  map[string]int
}

func (s *StateMachine) AppendLogs(input AppendLogsArgs, reply *AppendLogsReply) error {
	for _, v := range input.Entries {
		s.Log = append(s.Log, Log{Log: v.Log, Term: s.Term})
	}
	channel := s.Node.Channels()
	for k, c := range channel {
		appendEntriesReply := &AppendEntriesReply{}
		sendLogs := []Log{}
		fmt.Printf("NextIndex: %v\n", s.NextIndex)
		sendLogs = append(sendLogs, s.Log[s.NextIndex[k]-1:]...)
		prevLogIndex := s.NextIndex[k] - 1
		var prevLogTerm int
		if s.NextIndex[k] == 1 {
			prevLogTerm = s.Term
		} else {
			prevLogTerm = s.Log[prevLogIndex-1].Term
		}
		appendEntriesArgs := AppendEntriesArgs{
			Log:          sendLogs,
			Term:         s.Term,
			PrevLogIndex: prevLogIndex,
			LeaderCommit: s.CommitIndex,
			PrevLogTerm:  prevLogTerm,
		}
		err := c.Call("StateMachine.AppendEntries", appendEntriesArgs, appendEntriesReply)
		if err != nil {
			fmt.Printf("Failed to send heartbeat: %v\n", err)
			s.Node.Network().Remove(k)
		}

		if appendEntriesReply.Success {
			s.MatchIndex[k] = len(s.Log)
			s.NextIndex[k] = len(s.Log) + 1
		} else {
			fmt.Printf("NextIndex: %v\n", s.NextIndex)
			s.NextIndex[k] -= 1
			s.AppendLogs(input, reply)
		}
	}
	matchIndexSlice := []int{}
	for _, v := range s.MatchIndex {
		matchIndexSlice = append(matchIndexSlice, v)
	}
	matchIndexSlice = sort.IntSlice(matchIndexSlice)
	for _, v := range matchIndexSlice {
		cnt := 0
		for _, v2 := range matchIndexSlice {
			if v2 >= v {
				cnt++
			}
		}
		if cnt > len(matchIndexSlice)/2 {
			s.CommitIndex = v
			break
		}
	}
	fmt.Printf("Log: %v\n", s.Log)
	return nil
}

func (s *StateMachine) AppendEntries(input AppendEntriesArgs, reply *AppendEntriesReply) error {
	if input.Term < s.Term {
		return nil
	}
	s.Term = input.Term
	s.Role = "follower"
	s.HeartbeatWatch <- 1
	if len(input.Log) == 0 {
		// s.Log = append(s.Log, input.Log...)
		fmt.Printf("Log: %v\n", s.Log)
		return nil
	}
	/* if input.PrevLogIndex >= len(s.Log) {
		reply.Success = false
		return nil
	} */
	// s.Log = append(s.Log[:input.PrevLogIndex-1], input.Log...)
	fmt.Printf("len(s.Log): %d, input.PrevLogIndex: %d\n", len(s.Log), input.PrevLogIndex)
	if len(s.Log) < input.PrevLogIndex {
		reply.Success = false
		return nil
	}

	s.Log = append(s.Log[:input.PrevLogIndex], input.Log...)

	if input.LeaderCommit > s.CommitIndex {
		s.CommitIndex = int(math.Min(float64(input.LeaderCommit), float64(len(s.Log)-1)))
	}
	reply.Success = true
	return nil
}

func (s *StateMachine) RequestVote(input RequestVoteArgs, reply *RequestVoteReply) error {
	fmt.Println("RequestVote Start")
	if input.Term < s.Term {
		fmt.Printf("RequestVote failed. InputTerm: %d, Term: %d\n", input.Term, s.Term)
		return nil
	}
	s.Term = input.Term
	s.Leader = input.Leader
	s.Role = "follower"
	reply.VoteGranted = true
	fmt.Printf("Role began follower, Term: %d, Role: %s, Leader: %s\n", s.Term, s.Role, s.Leader)
	return nil
}

func (s *StateMachine) Get(input GetArgs, reply *GetReply) error {
	reply.Node = s.Node
	reply.Log = s.Log
	reply.Term = s.Term
	reply.Leader = s.Leader
	reply.Role = s.Role
	reply.CommitIndex = s.CommitIndex
	reply.LastApplied = s.LastApplied
	reply.NextIndex = s.NextIndex
	reply.MatchIndex = s.MatchIndex
	return nil
}

func (s *StateMachine) HeartBeat() {
	fmt.Println("HeartBeat")
	channel := s.Node.Channels()
	for k, ch := range channel {
		appendEntriesReply := &AppendEntriesReply{}
		err := ch.Call("StateMachine.AppendEntries", AppendEntriesArgs{Term: s.Term}, appendEntriesReply)
		if err != nil {
			fmt.Printf("Failed to send heartbeat: %v\n", err)
			s.Node.Network().Remove(k)
		}
	}
	time.Sleep(1 * time.Second)
}

func (s *StateMachine) ExecLeader() {
	channels := s.Node.Channels()
	fmt.Printf("Channels: %v\n", channels)
	for k := range channels {
		if _, ok := s.NextIndex[k]; !ok {
			fmt.Printf("NextIndex: %v\n", s.NextIndex)
			fmt.Printf("Log length: %d\n", len(s.Log))
			s.NextIndex[k] = len(s.Log) + 1
			fmt.Printf("NextIndex: %v\n", s.NextIndex)
		}
	}
	s.HeartBeat()
}

func (s *StateMachine) ExecFollower(heartbeatWatch chan int) {
	seed := time.Now().UnixNano()
	r := rand.New(rand.NewSource(seed))
	val := r.Intn(1000) + 2000
	select {
	case v := <-heartbeatWatch:
		if v == 1 {
			log.Println("Heartbeat is working")
		} else {
			log.Println("Heartbeat is not working(heartbeatWatch)")
		}
	case <-time.After(time.Duration(val) * time.Millisecond):
		if s.Role == "leader" {
			return
		}
		log.Println("Timeout")
		log.Println("Heartbeat is not working")
		s.Term++
		s.Role = "candidate"
		fmt.Printf("Role began candidate: Role: %s, Leader: %s, Term: %d\n", s.Role, s.Leader, s.Term)
	}
}

func (s *StateMachine) ExecCandidate() {
	channels := s.Node.Channels()
	voteGrantedCnt := 0
	for k, c := range channels {
		requestVoteReply := RequestVoteReply{}
		ch := make(chan error, 1)
		go func() {
			ch <- c.Call("StateMachine.RequestVote", RequestVoteArgs{Term: s.Term, Leader: s.Node.Name}, &requestVoteReply)
		}()
		select {
		case err := <-ch:
			if err != nil {
				fmt.Printf("RequestVote Error: %v\n", err)
				s.Node.Network().Remove(k)
				fmt.Printf("key: %v, Channel: %v\n", k, s.Node.Channels())
				continue
			}
		case <-time.After(500 * time.Millisecond):
			fmt.Println("RequestVote Timeout")
			s.Node.Network().Remove(k)
			continue
		}
		if requestVoteReply.VoteGranted {
			voteGrantedCnt++
		}
	}
	if voteGrantedCnt > len(channels)/2 {
		s.Role = "leader"
		s.Leader = s.Node.Name
	}
	fmt.Printf("Role began leader: Role: %s, Leader: %s, Term: %d\n", s.Role, s.Leader, s.Term)
}
