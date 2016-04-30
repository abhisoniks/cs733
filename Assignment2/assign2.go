package main

import (

	"strconv"
)

var Action = make(chan interface{}, 40)
var clientCh = make(chan interface{}, 40)
var timeoutCh = make(chan interface{}, 40)
var netCh = make(chan interface{}, 40)

type logEntry struct {
	index   int
	term    int
	command string
}

type StateMachine struct {
	leaderId      int
	state         string
	votedFor      int // 0 if not voted for anyone else Id of candidate to whom voted
	currentTerm   int
	currentCommit int // it tell index upto which log is commited
	log           []logEntry
	commitIndex   int
	lastApplied   int //index of highest log entry applied to state machine
	nextIndex     [5]int
	matchIndex    [5]int
	voteGranted   [2]int //First index counts number of positive voteResponse.second index counts number of negative voteResponse
}

func defaultSet() { // default value of state machine variable when state Machine is first time starte
	sm.state = "Follower"
	sm.votedFor = 6
}

type Append struct {
	a []byte
}

type AppendEntriesReq struct { //AppendEntriesReq(term int,leaderId int, prevLogIndex int,prevLogTerm int,entries[] string,leaderCommit int)
	term         int
	leaderId     int
	prevLogIndex int
	prevLogTerm  int
	entries      []logEntry
	leaderCommit int
}
type AppendEntriesResp struct {
	term     int
	senderId int
	AEResp   bool
}
type Timeout struct {
}

type VoteReq struct {
	term         int //candidate’s term
	candidateId  int //	int candidate requesting vote
	lastLogIndex int //index of candidate’s last log entry (§5.4)
	lastLogTerm  int //of candidate’s last log entry (§5.4)

}
type VoteResp struct {
	term     int  // for which term this voteRes is
	voteResp bool //tells voteResponse yes or No
}

func (sm *StateMachine) processEvent() {
	select {

	case appendMsg := <-clientCh:
		append1, _ := appendMsg.(Append) //convert it to Append
		sm.Append(append1)
	case peerMsg := <-netCh:
		switch peerMsg.(type) {
		case AppendEntriesReq: //if coming through channel is struct of type AppendEntriesReq
			appendER, _ := peerMsg.(AppendEntriesReq)
			sm.AppendEntriesReq(appendER)
		case VoteReq:
			voteRQ, _ := peerMsg.(VoteReq)
			sm.VoteReq(voteRQ)
		case VoteResp:
			vRes, _ := peerMsg.(VoteResp)
			sm.VoteResp(vRes)
		case AppendEntriesResp:
			ApRes, _ := peerMsg.(AppendEntriesResp)
			sm.AppendEntriesResp(ApRes)
		}
	//if coming through channel is struct of type AppendEntriesReq	}
	case <-timeoutCh:
		sm.Timeout()
	}
}

func (sm *StateMachine) Append(appendEvent Append) {
	switch sm.state {
	case "Follower":
		Action <- "Commit(" + string(appendEvent.a) + ",Leader is at id 4)"
	case "Candidate":
		Action <- "Commit(" + string(appendEvent.a) + ",error)"
	case "Leader": //Logstore ,Send AppendEntriesRPC ,reset Alarm
		pIndex, pTerm, _ := sm.getPrev()
		newCommand := string(appendEvent.a[:])
		sm.log = append(sm.log, logEntry{pIndex + 1, pTerm, newCommand})
		Action <- [3]string{"LogStore(" + strconv.Itoa(pIndex+1) + ")", "for each server p send(p,AppendEntriesReq(" + strconv.Itoa(sm.currentTerm) + "," + strconv.Itoa(0) + "," + strconv.Itoa(pIndex) + "," + strconv.Itoa(pTerm) + "," + "log[" + strconv.Itoa(pIndex) + ":]" + "," + strconv.Itoa(sm.commitIndex) + ")", "Alarm(t)"}

	}
}

func (sm *StateMachine) AppendEntriesReq(AER_Event AppendEntriesReq) {
	switch sm.state {
	case "Follower":
		if len(sm.log) == 0 { //machine has no entry at all ,copy leader's log as it is
			sm.log = make([]logEntry, 0)
			for j := 0; j < len(AER_Event.entries); j++ {
				sm.log = append(sm.log, AER_Event.entries[j])
			}
			sm.currentTerm = AER_Event.entries[len(AER_Event.entries)-1].term //set current term equal to leader's term
			sm.votedFor = 6
			i := strconv.Itoa(sm.currentTerm)
			Action <- [4]string{"true", i, strconv.Itoa(0), "Alarm(t)"} // 0 is sender id
			break
		}
		prevIndex, _, _ := sm.getPrev()
		if AER_Event.term < sm.currentTerm || AER_Event.prevLogIndex > prevIndex { // leader's Term is less than currentTerm
			i := strconv.Itoa(sm.currentTerm)
			Action <- [4]string{"false", i, strconv.Itoa(0), "Alarm(t)"} // send cuurentTerm to leader to update itself
			break
		}
		if AER_Event.term >= sm.currentTerm && AER_Event.entries == nil { //check if its a heartbeat check previous term and index
			if sm.log[AER_Event.prevLogIndex].term == AER_Event.prevLogTerm { //in both case reset Timeout
				sm.currentTerm = AER_Event.term
				i := strconv.Itoa(sm.currentTerm)
				Action <- [4]string{"true", i, strconv.Itoa(0), "Alarm(t)"}
				sm.commitIndex = AER_Event.leaderCommit //update your commitIndex
			} else {
				i := strconv.Itoa(sm.currentTerm)
				Action <- [4]string{"false", i, strconv.Itoa(0), "Alarm(t)"}
			}
		} else { // this is not Heartbeat message.
			pIndex := AER_Event.prevLogIndex
			pTerm := sm.log[pIndex].term
			pCommand := sm.log[pIndex].command
			if AER_Event.prevLogTerm != pTerm {
				i := strconv.Itoa(sm.currentTerm) // reply false
				Action <- [4]string{"false", i, strconv.Itoa(0), "Alarm(t)"}
			} else if AER_Event.prevLogTerm == pTerm && pCommand == AER_Event.entries[pIndex].command { //all three matches
				for k := pIndex + 1; k < len(AER_Event.entries); k++ { // whereever matches copy whole log of leader and set currentTerm equal to leader's term
					sm.log = append(sm.log, AER_Event.entries[k])
				}
				sm.commitIndex = AER_Event.leaderCommit
				sm.currentTerm = AER_Event.term
				sm.votedFor = 6 // for this term votedFor is null
				i := strconv.Itoa(sm.currentTerm)
				Action <- [4]string{"true", i, strconv.Itoa(0), "Alarm(t)"}
			} else if AER_Event.prevLogTerm == pTerm && pCommand != AER_Event.entries[pIndex].command { //command does not matches
				
				i := strconv.Itoa(sm.currentTerm)
				Action <- [4]string{"false", i, strconv.Itoa(0), "Alarm(t)"}
			}
		}
	case "Leader": //
		if AER_Event.term < sm.currentTerm { // leader's Term is less than currentTerm
			i := strconv.Itoa(sm.currentTerm)
			Action <- [4]string{"false", i, strconv.Itoa(0), "Alarm(t)"} // send cuurentTerm to leader to update itself
			break
		} else {
			sm.state = "Follower"
			sm.voteGranted = [2]int{0, 0}
			sm.votedFor = 6
			i := strconv.Itoa(sm.currentTerm)
			Action <- [4]string{"false", i, strconv.Itoa(0), "Alarm(t)"} // send cuurentTerm to leader to update itself
		}

	case "Candidate":
		prevIndex, _, _ := sm.getPrev()
		if AER_Event.term < sm.currentTerm  { // leader's Term is less than currentTerm
			i := strconv.Itoa(sm.currentTerm)
			Action <- [4]string{"false", i, strconv.Itoa(0), "Alarm(t)"} // send cuurentTerm to leader to update itself
			break
		}
		pIndex := AER_Event.prevLogIndex
		pTerm := sm.log[pIndex].term
		_ = sm.log[pIndex].command
		if AER_Event.term >= sm.currentTerm && AER_Event.entries == nil { //check if its a heartbeat check previous term and index
			if  (pTerm >= AER_Event.prevLogTerm &&  AER_Event.prevLogIndex >= prevIndex){
				sm.commitIndex = AER_Event.leaderCommit //update your commitIndex
				sm.state = "Follower"
				sm.currentTerm = AER_Event.term
				sm.voteGranted = [2]int{0, 0}
				sm.votedFor = 6
				i := strconv.Itoa(sm.currentTerm)
				Action <- [4]string{"true", i, strconv.Itoa(0), "Alarm(t)"}
			} else {
				i := strconv.Itoa(sm.currentTerm)
				Action <- [4]string{"false", i, strconv.Itoa(0), "Alarm(t)"}
			}
		} else { // this is not Heartbeat message.
			pIndex := AER_Event.prevLogIndex
			pTerm := sm.log[pIndex].term
			pCommand := sm.log[pIndex].command

			if AER_Event.prevLogTerm != pTerm {
			
				i := strconv.Itoa(sm.currentTerm) // reply false
				Action <- [4]string{"false", i, strconv.Itoa(0), "Alarm(t)"}
			} else if AER_Event.prevLogTerm == pTerm && pCommand == AER_Event.entries[pIndex].command { //all three matches
				for k := pIndex + 1; k < len(AER_Event.entries); k++ { // whereever matches copy whole log of leader and set currentTerm equal to leader's term
					sm.log = append(sm.log, AER_Event.entries[k])
				}
				
				sm.state = "Follower"
				sm.commitIndex = AER_Event.leaderCommit
				sm.currentTerm = AER_Event.term
				sm.votedFor = 6 // for this term votedFor is null
				i := strconv.Itoa(sm.currentTerm)
				Action <- [4]string{"true", i, strconv.Itoa(0), "Alarm(t)"}

			} else if AER_Event.prevLogTerm == pTerm && pCommand != AER_Event.entries[pIndex].command {
				
				i := strconv.Itoa(sm.currentTerm)
				Action <- [4]string{"false", i, strconv.Itoa(0), "Alarm(t)"}
			}
		}
	}
}

func (sm *StateMachine) Timeout() {
	switch sm.state {
	case "Follower": //change state increase term by 1 and start election and set election timer
		sm.currentTerm = sm.currentTerm + 1
		pIndex, pTerm, _ := sm.getPrev()
		sm.state = "Candidate"
		sm.votedFor = 6 // vote to itself
		Action <- [2]string{"for all peer p send(p,VoteReq(" + strconv.Itoa(sm.currentTerm) + "," + strconv.Itoa(0) + "," + strconv.Itoa(pIndex) + "," + strconv.Itoa(pTerm) + "))", "Alarm(t)"}
	case "Leader": // send heart beat message and reset timer
		pIndex, pTerm, _ := sm.getPrev()
		Action <- [2]string{"for all peer p send(p,AppendEntriesReq(" + strconv.Itoa(sm.currentTerm) + "," + strconv.Itoa(sm.leaderId) + "," + strconv.Itoa(pIndex) + "," + strconv.Itoa(pTerm) + "," + "nil," + strconv.Itoa(sm.commitIndex) + "))", "Alarm(t)"}
	case "Candidate": //increase currentTerm ask for vote again and reset Alarm
		sm.currentTerm = sm.currentTerm + 1
		pIndex, pTerm, _ := sm.getPrev()
		sm.state = "Candidate"
		sm.votedFor = 6 // vote to itself first
		Action <- [2]string{"for all peer p send(p,VoteReq(" + strconv.Itoa(sm.currentTerm) + "," + strconv.Itoa(0) + "," + strconv.Itoa(pIndex) + "," + strconv.Itoa(pTerm) + "))", "Alarm(t)"}
	}
}

func (sm *StateMachine) AppendEntriesResp(AER_Event AppendEntriesResp) {
	switch sm.state {
	case "Follower": //ignore it
	case "Candidate": //ignore it
	case "Leader":
		if AER_Event.AEResp == true {
			lastIndex := sm.log[len(sm.log)-1].index
			sm.nextIndex[AER_Event.senderId] = lastIndex + 1
			sm.matchIndex[AER_Event.senderId] = lastIndex
		} else {
			sm.nextIndex[AER_Event.senderId] -= 1
			newpIndex := sm.log[sm.nextIndex[AER_Event.senderId]-1].index //calculate new previous index to send with AppenEntriesRPC
			newpTerm := sm.log[sm.nextIndex[AER_Event.senderId]-1].term   //calculate new previous term to send with AppenEntriesRPC
			Action <- [2]string{"Send(" + strconv.Itoa(AER_Event.senderId) + ",AppendEntriesReq(" + strconv.Itoa(sm.currentTerm) + "," + strconv.Itoa(0) + "," + strconv.Itoa(newpIndex) + "," + strconv.Itoa(newpTerm) + "," + "log[" + strconv.Itoa(sm.nextIndex[AER_Event.senderId]-1) + ":]" + "," + strconv.Itoa(sm.commitIndex) + ")", "Alarm(t)"}
		}

		count := make(map[int]int)
		for i := 0; i < 5; i++ {
			var sum int = 0
			if sm.matchIndex[i] > sm.commitIndex {
				for j := 0; j < 5; j++ {
					if sm.matchIndex[j] == sm.matchIndex[i] {
						sum = sum + 1
					}
				}
				count[sm.matchIndex[i]] = sum
			} else {
				count[sm.matchIndex[i]] = 0
			}
		}
		temp := sm.commitIndex
		for key, value := range count {
			if key > temp && value >= 3 {
				temp = key
			}
		}
		if temp != sm.commitIndex { //commitIndex is changed
			sm.commitIndex = temp
			Action <- "Commit(" + strconv.Itoa(sm.commitIndex) + "," + sm.log[len(sm.log)-1].command + ")"
		}
	}
}

func (sm *StateMachine) VoteReq(VRQ_Event VoteReq) {
	cID := strconv.Itoa(VRQ_Event.candidateId)
	cT := strconv.Itoa(sm.currentTerm)
	pIndex, _, _ := sm.getPrev() //will return previouscommand,previousId and previousTerm
	switch sm.state {
	case "Follower":
		if VRQ_Event.term >= sm.currentTerm && (sm.votedFor == 6 || sm.votedFor == VRQ_Event.candidateId) {
			if VRQ_Event.lastLogIndex >= pIndex {
				Action <- "Send(" + cID + ",VoteResp(" + cT + ",voteGranted=yes))"
				sm.votedFor = VRQ_Event.candidateId
			} else {
				Action <- "Send(" + cID + ",VoteResp(" + cT + ",voteGranted=No))"
			}
		} else {
			Action <- "Send(" + cID + ",VoteResp(" + cT + ",voteGranted=No))"
		}

	case "Leader":
		if VRQ_Event.term >= sm.currentTerm {
			sm.state = "Follower" // some of initialization before going in Follower state
			sm.votedFor = 6
			sm.voteGranted = [2]int{0, 0}
			Action <- "Send(" + cID + ",VoteResp(" + cT + ",voteGranted=No))"
		} else {
			Action <- "Send(" + cID + ",VoteResp(" + cT + ",voteGranted=No))"
		}

	case "Candidate":
		if VRQ_Event.term > sm.currentTerm && VRQ_Event.lastLogIndex >= pIndex { //if voteFrom higher term
			sm.state = "Follower"
			sm.voteGranted = [2]int{0, 0}
			sm.votedFor = VRQ_Event.candidateId
			sm.currentTerm = VRQ_Event.term
			//Action<-;
			Action <- [2]string{"Send(" + cID + ",VoteResp(" + strconv.Itoa(sm.currentTerm) + ",voteGranted=yes))", "Alarm(t)"}
		} else {
			Action <- "Send(" + cID + ",VoteResp(" + cT + ",voteGranted=No))"
		}
	}
}

func (sm *StateMachine) VoteResp(VRS_Event VoteResp) {
	switch sm.state {
	case "Follower": //ignore it
	case "Leader": // ignore it
	case "Candidate":

		if VRS_Event.term > sm.currentTerm { // If term is greater than sm.term go to Follower state & reInitialize its timer

			sm.state = "Follower"
			sm.votedFor = 6 // 6 means for this term it has not give vote to anyone
			sm.currentTerm = VRS_Event.term
			Action <- "Alarm(t)"
		}
		if VRS_Event.term <= sm.currentTerm && VRS_Event.voteResp == true {

			sm.voteGranted[0] += 1
		}
		if VRS_Event.term <= sm.currentTerm && VRS_Event.voteResp == false {

			sm.voteGranted[1] += 1
		}
		if sm.voteGranted[0] >= 3 { // become Leader
			pIndex, pTerm, _ := sm.getPrev()
			sm.voteGranted = [2]int{0, 0} //reintialize voteGranted
			sm.votedFor = 6               //reintialize voteFor
			sm.state = "Leader"
			sm.leaderId = 0
			//pIndex,pTerm,_:=sm.getPrev()
			Action <- [2]string{"for all peer p send(p,AppendEntriesReq(" + strconv.Itoa(sm.currentTerm) + "," + strconv.Itoa(0) + "," + strconv.Itoa(pIndex) + "," + strconv.Itoa(pTerm) + "," + "nil," + strconv.Itoa(sm.commitIndex) + ")", "Alarm(t)"} //Reintialize timer

		}
		if sm.voteGranted[1] >= 3 { //become Follower here
			sm.voteGranted = [2]int{0, 0} //reintialize voteGranted
			sm.votedFor = 6               //reintialize voteFor
			sm.state = "Follower"
			Action <- "Alarm(t)"

		}
		if sm.voteGranted[1] == 2 && sm.voteGranted[0] == 2 { //voteSplit  && Re-electionn
			pIndex, pTerm, _ := sm.getPrev()
			sm.currentTerm = sm.currentTerm + 1
			sm.votedFor = 0 // vote to itself.sm own id is 0
			sm.voteGranted = [2]int{0, 0}
			Action <- [2]string{"for all peer p send(p,VoteReq(" + strconv.Itoa(sm.currentTerm) + "," + strconv.Itoa(0) + "," + strconv.Itoa(pIndex) + "," + strconv.Itoa(pTerm) + "))", "Alarm(t)"}

		}
	}
}

func (sm *StateMachine) getPrev() (int, int, string) { //this function returns previous log index log term and previous command
	var prevI int
	var prevT int
	var prevC string
	for i := 0; i < len(sm.log); i++ {
		prevI = sm.log[i].index
		prevT = sm.log[i].term
		prevC = sm.log[i].command
	}
	return prevI, prevT, prevC
}
