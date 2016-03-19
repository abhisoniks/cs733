package main
import (
	"fmt"
	//"github.com/cs733-iitb/log"
)

func (sm *STATEMACHINE) processEvent() {
	for {
		select {
		case appendMsg := <-sm.SMchanels.clientCh:
			append1, _ := appendMsg.(APPEND) //convert it to Append
			sm.Append(append1)
		case peerMsg := <-sm.SMchanels.netCh:
			switch peerMsg.(type) {
			case APPENDENTRIESREQ: //if coming through channel is struct of type APPENDENTRIESREQ
				appendER, _ := peerMsg.(APPENDENTRIESREQ)
				sm.APPENDENTRIESREQ(appendER)
			case VOTEREQ:
				voteRQ, _ := peerMsg.(VOTEREQ)
				sm.VOTEREQ(voteRQ)
			case VOTERESP:
				vRes, _ := peerMsg.(VOTERESP)
				sm.VOTERESP(vRes)
			case APPENDENTRIESRESP:
				ApRes, _ := peerMsg.(APPENDENTRIESRESP)
				sm.APPENDENTRIESRESP(ApRes)
			}
		//if coming through channel is struct of type APPENDENTRIESREQ	}
		case <-sm.SMchanels.timeoutCh:
				sm.Timeout()
		}
	} // for loop end
}


func (sm *STATEMACHINE) Append(appendEVENT APPEND) {
	switch sm.STATE {
	case "Follower":
		c := COMMIT{DATA: string(appendEVENT.A), ERROR: 404}
		sm.SMchanels.Action <- c
	case "Candidate":
		c := COMMIT{DATA: string(appendEVENT.A), ERROR: 404}
		sm.SMchanels.Action <- c
	case "Leader": //LOGSTORE ,SEND AppendENTRIESRPC ,reset ALARM
		pINDEX, pTERM, _ := sm.getPrev()
		newCommand := appendEVENT.A[:]
		sm.LOG = append(sm.LOG, LOGENTRY{pINDEX + 1, pTERM, newCommand})
		t:=LOGENTRY{pINDEX+1,pTERM,appendEVENT.A}
		l := LOGSTORE{INDEX:pINDEX+1,DATA:t}
		//COMMITINFO{DATA:t,index:pINDEX+1,Err:0}
		sm.SMchanels.Action <- l
		apr := APPENDENTRIESREQ{TERM: sm.CURRENTTERM, LEADERID: sm.ID, PREVLOGINDEX: pINDEX, PREVLOGTERM: pTERM, ENTRIES: sm.LOG, LEADERCOMMIT: sm.COMMITINDEX}
		fmt.Println("apr to send",apr)
		sm.SMchanels.Action <- SEND{ID_TOSEND: 500, EVENT: apr}
		sm.SMchanels.Action <- ALARM{TIME: 500}
		//	sm.SMchanels.Action <- [3]string{"LOGSTORE(" + strconv.Itoa(pINDEX+1) + ")", "for each server p SEND(p,APPENDENTRIESREQ(" + strconv.Itoa(sm.CURRENTTERM) + "," + strconv.Itoa(0) + "," + strconv.Itoa(pINDEX) + "," + strconv.Itoa(pTERM) + "," + "log[" + strconv.Itoa(pINDEX) + ":]" + "," + strconv.Itoa(sm.COMMITINDEX) + ")", "ALARM(t)"}

	}
}

func (sm *STATEMACHINE) APPENDENTRIESREQ(AER_EVENT APPENDENTRIESREQ) {
	switch sm.STATE {
	case "Follower":
		prevINDEX, _, _ := sm.getPrev()
        if AER_EVENT.TERM < sm.CURRENTTERM || AER_EVENT.PREVLOGINDEX > prevINDEX { // leader's TERM is less than CURRENTTERM
			sm.SMchanels.Action <- SEND{ID_TOSEND: sm.LEADERID, EVENT: APPENDENTRIESRESP{TERM: sm.CURRENTTERM, SENDERID: sm.ID, AERESP: false}}
			t := sm.getElectionTime()
			sm.SMchanels.Action <- ALARM{TIME: t}
			break
		}
		if AER_EVENT.TERM >= sm.CURRENTTERM && AER_EVENT.ENTRIES == nil { //check if its a heartbeat check previous TERM and INDEX
			if sm.LOG[AER_EVENT.PREVLOGINDEX].TERM == AER_EVENT.PREVLOGTERM { //in both case reset Timeout
				sm.CURRENTTERM = AER_EVENT.TERM
				//	i := strconv.Itoa(sm.CURRENTTERM)
				//	sm.SMchanels.Action <- [4]string{"true", i, strconv.Itoa(0), "ALARM(t)"}
				sm.LEADERID = AER_EVENT.LEADERID
				sm.SMchanels.Action <- SEND{ID_TOSEND: sm.LEADERID, EVENT: APPENDENTRIESRESP{TERM: sm.CURRENTTERM, SENDERID: sm.ID, AERESP: true}}
				t := sm.getElectionTime()
				sm.SMchanels.Action <- ALARM{TIME: t}
				sm.COMMITINDEX = AER_EVENT.LEADERCOMMIT //update your COMMITINDEX
			} else {
				
				sm.SMchanels.Action <- SEND{ID_TOSEND: sm.LEADERID, EVENT: APPENDENTRIESRESP{TERM: sm.CURRENTTERM, SENDERID: sm.ID, AERESP: false}}
				t := sm.getElectionTime()
				sm.SMchanels.Action <- ALARM{TIME: t}

			}
		} else { // this is not Heartbeat message.
		   
			pINDEX := AER_EVENT.PREVLOGINDEX
			pTERM := sm.LOG[pINDEX].TERM
			pCommand := string(sm.LOG[pINDEX].COMMAND)
			if AER_EVENT.PREVLOGTERM != pTERM {
				sm.SMchanels.Action <- SEND{ID_TOSEND: sm.LEADERID, EVENT: APPENDENTRIESRESP{TERM: sm.CURRENTTERM, SENDERID: sm.ID, AERESP: false}}
				t := sm.getElectionTime()
				sm.SMchanels.Action <- ALARM{TIME: t}
			} else if AER_EVENT.PREVLOGTERM == pTERM && pCommand == string(AER_EVENT.ENTRIES[pINDEX].COMMAND) { //all three matches
				var k int
				for k = pINDEX + 1; k < len(AER_EVENT.ENTRIES); k++ { // whereever matches copy whole log of leader and set CURRENTTERM equal to leader's TERM
					sm.LOG = append(sm.LOG, AER_EVENT.ENTRIES[k])
					sm.SMchanels.Action <- LOGSTORE{k,AER_EVENT.ENTRIES[k]}
				}
				//t:=AER_EVENT.ENTRIES[k]
				//COMMITINFO{DATA:t,index:k,Err:0}
				
				sm.COMMITINDEX = AER_EVENT.LEADERCOMMIT
				sm.CURRENTTERM = AER_EVENT.TERM
				sm.VOTEDFOR = 6 // for this TERM votedFor is null
				sm.LEADERID = AER_EVENT.LEADERID
				sm.SMchanels.Action <- SEND{ID_TOSEND: sm.LEADERID, EVENT: APPENDENTRIESRESP{TERM: sm.CURRENTTERM, SENDERID: sm.ID, AERESP: true}}
				ti := sm.getElectionTime()
				sm.SMchanels.Action <- ALARM{TIME: ti}
			} else if AER_EVENT.PREVLOGTERM == pTERM && pCommand != string(AER_EVENT.ENTRIES[pINDEX].COMMAND) { //command does not matches
	
				sm.SMchanels.Action <- SEND{ID_TOSEND: sm.LEADERID, EVENT: APPENDENTRIESRESP{TERM: sm.CURRENTTERM, SENDERID: sm.ID, AERESP: false}}
				t := sm.getElectionTime()
				sm.SMchanels.Action <- ALARM{TIME: t}
			}
		}
	case "Leader": //
		if AER_EVENT.TERM < sm.CURRENTTERM { // leader's TERM is more than CURRENTTERM
			sm.SMchanels.Action <- SEND{ID_TOSEND: sm.LEADERID, EVENT: APPENDENTRIESRESP{TERM: sm.CURRENTTERM, SENDERID: sm.ID, AERESP: false}}
			sm.SMchanels.Action <- ALARM{TIME: 500}
			break
		} else {
			sm.STATE = "Follower"
			////fmt.Println("New state is ", sm.STATE, " ", sm)
			sm.VOTEGRANTED = [2]int{0, 0}
			sm.VOTEDFOR = 6
			sm.SMchanels.Action <- SEND{ID_TOSEND: sm.LEADERID, EVENT: APPENDENTRIESRESP{TERM: sm.CURRENTTERM, SENDERID: sm.ID, AERESP: false}}
			t := sm.getElectionTime()
			sm.SMchanels.Action <- ALARM{TIME: t}
		}

	case "Candidate":
		prevINDEX, _, _ := sm.getPrev()
		if AER_EVENT.TERM < sm.CURRENTTERM { // leader's TERM is less than CURRENTTERM
			sm.SMchanels.Action <- SEND{ID_TOSEND: sm.LEADERID, EVENT: APPENDENTRIESRESP{TERM: sm.CURRENTTERM, SENDERID: sm.ID, AERESP: false}}
			t := sm.getElectionTime()
			sm.SMchanels.Action <- ALARM{TIME: t}
			return
		}
		pINDEX := AER_EVENT.PREVLOGINDEX
		var pTERM int
		if len(sm.LOG) > 0 { //initial setUp
			pTERM = sm.LOG[pINDEX].TERM
		} else {
			pTERM = 0
		}
		//_ = sm.LOG[pINDEX].command
		if AER_EVENT.TERM >= sm.CURRENTTERM && AER_EVENT.ENTRIES == nil { //check if its a heartbeat check previous TERM and INDEX
			if pTERM >= AER_EVENT.PREVLOGTERM && AER_EVENT.PREVLOGINDEX >= prevINDEX {
				sm.COMMITINDEX = AER_EVENT.LEADERCOMMIT //update your COMMITINDEX
				sm.STATE = "Follower"
				sm.CURRENTTERM = AER_EVENT.TERM
				sm.VOTEGRANTED = [2]int{0, 0}
				sm.VOTEDFOR = 6
				sm.LEADERID = AER_EVENT.LEADERID
				//	i := strconv.Itoa(sm.CURRENTTERM)
				//sm.SMchanels.Action <- [4]string{"true", i, strconv.Itoa(0), "ALARM(t)"}
				sm.SMchanels.Action <- SEND{ID_TOSEND: sm.LEADERID, EVENT: APPENDENTRIESRESP{TERM: sm.CURRENTTERM, SENDERID: sm.ID, AERESP: true}}
				t := sm.getElectionTime()
				sm.SMchanels.Action <- ALARM{TIME: t}
			} else {
				
				sm.SMchanels.Action <- SEND{ID_TOSEND: sm.LEADERID, EVENT: APPENDENTRIESRESP{TERM: sm.CURRENTTERM, SENDERID: sm.ID, AERESP: false}}
				t := sm.getElectionTime()
				sm.SMchanels.Action <- ALARM{TIME: t}
			}
		} else { // this is not Heartbeat message.
			pINDEX := AER_EVENT.PREVLOGINDEX
			pTERM := sm.LOG[pINDEX].TERM
			pCommand := string(sm.LOG[pINDEX].COMMAND)

			if AER_EVENT.PREVLOGTERM != pTERM {
				sm.SMchanels.Action <- SEND{ID_TOSEND: sm.LEADERID, EVENT: APPENDENTRIESRESP{TERM: sm.CURRENTTERM, SENDERID: sm.ID, AERESP: false}}
				t := sm.getElectionTime()
				sm.SMchanels.Action <- ALARM{TIME: t}
			} else if AER_EVENT.PREVLOGTERM == pTERM && pCommand == string(AER_EVENT.ENTRIES[pINDEX].COMMAND) { //all three matches
				//var k int
				for k := pINDEX + 1; k < len(AER_EVENT.ENTRIES); k++ { // whereever matches copy whole log of leader and set CURRENTTERM equal to leader's TERM
					sm.LOG = append(sm.LOG, AER_EVENT.ENTRIES[k])
					fmt.Println("LOGSTORE++")
					sm.SMchanels.Action <- LOGSTORE{k,AER_EVENT.ENTRIES[k]}
				}
				//t:=AER_EVENT.ENTRIES[k]
				//COMMITINFO{DATA:t,index:k,Err:0}
				sm.STATE = "Follower"
				sm.COMMITINDEX = AER_EVENT.LEADERCOMMIT
				sm.CURRENTTERM = AER_EVENT.TERM
				sm.VOTEDFOR = 6 // for this TERM votedFor is null
				sm.LEADERID = AER_EVENT.LEADERID
				sm.SMchanels.Action <- SEND{ID_TOSEND: sm.LEADERID, EVENT: APPENDENTRIESRESP{TERM: sm.CURRENTTERM, SENDERID: sm.ID, AERESP: true}}
				ti := sm.getElectionTime()
				sm.SMchanels.Action <- ALARM{TIME: ti}

			} else if AER_EVENT.PREVLOGTERM == pTERM && pCommand != string(AER_EVENT.ENTRIES[pINDEX].COMMAND) {
				aers := APPENDENTRIESRESP{TERM: sm.CURRENTTERM, SENDERID: sm.ID, AERESP: false}
				sm.SMchanels.Action <- SEND{ID_TOSEND: sm.LEADERID, EVENT: aers}

			}
		}
	}
}

func (sm *STATEMACHINE) Timeout() {
    //	fmt.Println("TIMEOUT")
	switch sm.STATE {
	case "Follower": //change state increase TERM by 1 and start election and set election timer
		sm.CURRENTTERM = sm.CURRENTTERM + 1
		pINDEX, pTERM, _ := sm.getPrev()
		sm.STATE = "Candidate"
		sm.VOTEDFOR = sm.ID // vote to itself
		sm.VOTEGRANTED[0] = 1
		VR := VOTEREQ{TERM: sm.CURRENTTERM, CANDIDATEID: sm.ID, LASTLOGINDEX: pINDEX, LASTLOGTERM: pTERM}
		sm.SMchanels.Action <- SEND{ID_TOSEND: 500, EVENT: VR}
		t := sm.getElectionTime()
		sm.SMchanels.Action <- ALARM{TIME: t}
	//	sm.SMchanels.Action <- [2]string{"for all peer p SEND(p,VOTEREQ(" + strconv.Itoa(sm.CURRENTTERM) + "," + strconv.Itoa(0) + "," + strconv.Itoa(pINDEX) + "," + strconv.Itoa(pTERM) + "))", "ALARM(t)"}
	case "Leader": // SEND heart beat message and reset timer
		pINDEX, pTERM, _ := sm.getPrev()
		//fmt.Println("2")
		aer := APPENDENTRIESREQ{TERM: sm.CURRENTTERM, LEADERID: sm.ID, PREVLOGINDEX: pINDEX, PREVLOGTERM: pTERM, ENTRIES: nil, LEADERCOMMIT: sm.COMMITINDEX}
		sm.SMchanels.Action <- SEND{ID_TOSEND: 500, EVENT: aer}
		sm.SMchanels.Action <- ALARM{TIME: 500}
	//sm.SMchanels.Action <- [2]string{"for all peer p SEND(p,APPEE=E," ",sm)
	case "Candidate":
		sm.STATE = "Candidate"
		sm.CURRENTTERM = sm.CURRENTTERM + 1
		sm.VOTEDFOR = sm.ID // vote to itself first
		sm.VOTEGRANTED[0] = 1
		//fmt.Println("New state is ", sm.STATE, " ", sm)

		pINDEX, pTERM, _ := sm.getPrev()
	    //	fmt.Println("VoteRequest send",sm)
		VR := VOTEREQ{TERM: sm.CURRENTTERM, CANDIDATEID: sm.ID, LASTLOGINDEX: pINDEX, LASTLOGTERM: pTERM}

		sm.SMchanels.Action <- SEND{ID_TOSEND: 500, EVENT: VR}
		t := sm.getElectionTime()

		sm.SMchanels.Action <- ALARM{TIME: t}
		//		sm.SMchanels.Action <- [2]string{"for all peer p SEND(p,VOTEREQ(" + strconv.Itoa(sm.CURRENTTERM) + "," + strconv.Itoa(0) + "," + strconv.Itoa(pINDEX) + "," + strconv.Itoa(pTERM) + "))", "ALARM(t)"}
	}
}

func (sm *STATEMACHINE) APPENDENTRIESRESP(AER_EVENT APPENDENTRIESRESP) {
	//fmt.Println("AppendEntriesResponse RECEIVED ", sm, " ", "response is", AER_EVENT)
	switch sm.STATE {
	case "Follower": break//ignore it
	case "Candidate": break//ignore it
	case "Leader":
		if AER_EVENT.AERESP == true {
			//fmt.Println("true")
			if len(sm.LOG) > 0 { // when leader has log .
				lastINDEX := sm.LOG[len(sm.LOG)-1].INDEX
				sm.NEXTINDEX[AER_EVENT.SENDERID] = lastINDEX + 1
				sm.MATCHINDEX[AER_EVENT.SENDERID] = lastINDEX
			} else { //leader has no log New set up
				lastINDEX := 0
				sm.NEXTINDEX[AER_EVENT.SENDERID] = lastINDEX
				sm.MATCHINDEX[AER_EVENT.SENDERID] = lastINDEX

			}
		} else {
			//fmt.Println("false")
			sm.NEXTINDEX[AER_EVENT.SENDERID] -= 1
			var newpINDEX, newpTERM int
			newpINDEX = 0
			newpTERM = 0

			if len(sm.LOG) > 0 {
				newpINDEX = sm.LOG[sm.NEXTINDEX[AER_EVENT.SENDERID]].INDEX //calculate new previous INDEX to SEND with AppenENTRIESRPC
				newpTERM = sm.LOG[sm.NEXTINDEX[AER_EVENT.SENDERID]].TERM   //calculate new previous TERM to SEND with AppenENTRIESRPC
			//	fmt.Println("AEREQST SEND")
				aer := APPENDENTRIESREQ{TERM: sm.CURRENTTERM, LEADERID: sm.ID, PREVLOGINDEX: newpINDEX, PREVLOGTERM: newpTERM, ENTRIES: sm.LOG[(sm.NEXTINDEX[AER_EVENT.SENDERID]):], LEADERCOMMIT: sm.COMMITINDEX}
				sm.SMchanels.Action <- SEND{ID_TOSEND: AER_EVENT.SENDERID, EVENT: aer}
			} else {
				newpINDEX = 0
				newpTERM = 0
			
				aer := APPENDENTRIESREQ{TERM: sm.CURRENTTERM, LEADERID: sm.ID, PREVLOGINDEX: newpINDEX, PREVLOGTERM: newpTERM, ENTRIES: nil, LEADERCOMMIT: sm.COMMITINDEX}
				sm.SMchanels.Action <- SEND{ID_TOSEND: AER_EVENT.SENDERID, EVENT: aer}
			}
			sm.SMchanels.Action <- ALARM{TIME: 500}

			//sm.SMchanels.Action <- [2]string{"SEND(" + strconv.Itoa(AER_EVENT.SENDERID) + ",APPENDENTRIESREQ(" + strconv.Itoa(sm.CURRENTTERM) + "," + strconv.Itoa(0) + "," + strconv.Itoa(newpINDEX) + "," + strconv.Itoa(newpTERM) + "," + "log[" + strconv.Itoa(sm.NEXTINDEX[AER_EVENT.SENDERID]-1) + ":]" + "," + strconv.Itoa(sm.COMMITINDEX) + ")", "ALARM(t)"}
		}

		count := make(map[int]int)
		for i := 0; i < 5; i++ {
			var sum int = 0
			if sm.MATCHINDEX[i] > sm.COMMITINDEX {
				for j := 0; j < 5; j++ {
					if sm.MATCHINDEX[j] == sm.MATCHINDEX[i] {
						sum = sum + 1
					}
				}
				count[sm.MATCHINDEX[i]] = sum
			} else {
				count[sm.MATCHINDEX[i]] = 0
			}
		}
		temp := sm.COMMITINDEX
		for key, value := range count {
			if key > temp && value >= 3 {
				temp = key
			}
		}
		if temp != sm.COMMITINDEX { //COMMITINDEX is changed
			sm.COMMITINDEX = temp
			sm.SMchanels.Action <- COMMIT{INDEX: sm.COMMITINDEX, DATA: string(sm.LOG[len(sm.LOG)-1].COMMAND)}
		}
	}
}

func (sm *STATEMACHINE) VOTEREQ(VRQ_EVENT VOTEREQ) {
	//fmt.Println("VOTEREQ")
	pINDEX, _, _ := sm.getPrev() //will return previouscommand,previousId and previousTERM
	//fmt.Println("VoteReq came ", sm, " ", "sender", "=", VRQ_EVENT)
	switch sm.STATE {
	case "Follower":
		if (sm.VOTEDFOR == 6) && VRQ_EVENT.TERM >= sm.CURRENTTERM   {
			if VRQ_EVENT.LASTLOGINDEX >= pINDEX {
				//	sm.SMchanels.Action <- "SEND(" + cID + ",VOTERESP(" + cT + ",VOTEGRANTED=yes))"
				V_res := VOTERESP{TERM: sm.CURRENTTERM, VOTERESP: true}
				sm.SMchanels.Action <- SEND{ID_TOSEND: VRQ_EVENT.CANDIDATEID, EVENT: V_res}
				sm.VOTEDFOR = VRQ_EVENT.CANDIDATEID
			//	fmt.Println("response send=", V_res)
			} else {
				//sm.SMchanels.Action <- "SEND(" + cID + ",VOTERESP(" + cT + ",VOTEGRANTED=No))"
				V_res := VOTERESP{sm.CURRENTTERM, false}
				sm.SMchanels.Action <- SEND{VRQ_EVENT.CANDIDATEID, V_res}
			//	fmt.Println("response send=", V_res)
			}
		} else {
			//sm.SMchanels.Action <- "SEND(" + cID + ",VOTERESP(" + cT + ",VOTEGRANTED=No))"
			V_res := VOTERESP{sm.CURRENTTERM, false}
			sm.SMchanels.Action <- SEND{VRQ_EVENT.CANDIDATEID, V_res}

		}

	case "Leader":
		if VRQ_EVENT.TERM >= sm.CURRENTTERM {
			sm.STATE = "Follower" // some of initialization before going in Follower state
						//fmt.Println("New state is ", sm.STATE, " ", sm)

			sm.VOTEDFOR = 6
			sm.VOTEGRANTED = [2]int{0, 0}
			V_res := VOTERESP{sm.CURRENTTERM, false}
			sm.SMchanels.Action <- SEND{VRQ_EVENT.CANDIDATEID, V_res}
		//	fmt.Println("response send=", V_res)
		} else {
			//sm.SMchanels.Action <- "SEND(" + cID + ",VOTERESP(" + cT + ",VOTEGRANTED=No))"
			V_res := VOTERESP{sm.CURRENTTERM, false}
			sm.SMchanels.Action <- SEND{VRQ_EVENT.CANDIDATEID, V_res}
		//	fmt.Println("response send=", V_res)
		}

	case "Candidate":
		if VRQ_EVENT.TERM > sm.CURRENTTERM { //if voteFrom higher TERM
			sm.STATE = "Follower"
			//fmt.Println("New state is ", sm.STATE, " ", sm)
			sm.VOTEGRANTED = [2]int{0, 0}
			sm.VOTEDFOR = 6 //VRQ_EVENT.CANDIDATEID
			sm.CURRENTTERM = VRQ_EVENT.TERM
			V_res := VOTERESP{sm.CURRENTTERM, false}
			sm.SMchanels.Action <- SEND{VRQ_EVENT.CANDIDATEID, V_res}
			t := sm.getElectionTime()
			sm.SMchanels.Action <- ALARM{TIME: t}
		//	fmt.Println("response send=", V_res)
		} else {
			V_res := VOTERESP{sm.CURRENTTERM, false}
			sm.SMchanels.Action <- SEND{VRQ_EVENT.CANDIDATEID, V_res}
		//	fmt.Println("response send=", V_res)
		}
	}
}

func (sm *STATEMACHINE) VOTERESP(VRS_EVENT VOTERESP) {
	//fmt.Println("---> Vote Me")

	switch sm.STATE {
	case "Follower": break
	case "Leader":  // fmt.Println("old voteRes");
	break
	case "Candidate":

		if VRS_EVENT.TERM > sm.CURRENTTERM { // If TERM is greater than sm.TERM go to Follower state & reInitialize its timer
			sm.STATE = "Follower"
			sm.VOTEDFOR = 6 // 6 means for this TERM it has not give vote to anyone
			sm.CURRENTTERM = VRS_EVENT.TERM
			//sm.SMchanels.Action <- "ALARM(t)"
			t := sm.getElectionTime()
			sm.SMchanels.Action <- ALARM{TIME: t}
			return
		}
		if VRS_EVENT.TERM <= sm.CURRENTTERM && VRS_EVENT.VOTERESP == true {
			sm.VOTEGRANTED[0] += 1
		}
		if VRS_EVENT.TERM <= sm.CURRENTTERM && VRS_EVENT.VOTERESP == false {
			sm.VOTEGRANTED[1] += 1
		}
		if sm.VOTEGRANTED[0] >= 3 { // become Leader
			pINDEX, pTERM, _ := sm.getPrev()
			sm.VOTEGRANTED = [2]int{0, 0} //reintialize VOTEGRANTED
			sm.VOTEDFOR = 6               //reintialize voteFor
			sm.STATE = "Leader"
			sm.LEADERID = sm.ID
			sm.LEADERID = 0
			
			aer := APPENDENTRIESREQ{sm.CURRENTTERM, sm.LEADERID, pINDEX, pTERM, nil, sm.COMMITINDEX}
		//	fmt.Println("send AppendEntries Request by",sm.ID," ",aer,)
			sm.SMchanels.Action <- SEND{0, aer}
			sm.SMchanels.Action <- ALARM{TIME: 500}
		}
		if sm.VOTEGRANTED[1] >= 3 { //become Follower here
			sm.VOTEGRANTED = [2]int{0, 0} //reintialize VOTEGRANTED
			sm.VOTEDFOR = 6               //reintialize voteFor
			sm.STATE = "Follower"
			//fmt.Println("New state is ", sm.STATE, " ", sm)
			//sm.SMchanels.Action <- "ALARM(t)"
			t := sm.getElectionTime()
			sm.SMchanels.Action <- ALARM{TIME: t}

		}
		if sm.VOTEGRANTED[1] == 2 && sm.VOTEGRANTED[0] == 2 { //voteSplit  && Re-electionn
			pINDEX, pTERM, _ := sm.getPrev()
			sm.CURRENTTERM = sm.CURRENTTERM + 1
			sm.VOTEDFOR = 0 // vote to itself.sm own id is 0
			sm.VOTEGRANTED = [2]int{0, 0}
			//sm.SMchanels.Action <- [2]string{"for all peer p SEND(p,VOTEREQ(" + strconv.Itoa(sm.CURRENTTERM) + "," + strconv.Itoa(0) + "," + strconv.Itoa(pINDEX) + "," + strconv.Itoa(pTERM) + "))", "ALARM(t)"}
			vr := VOTEREQ{sm.CURRENTTERM, sm.ID, pINDEX, pTERM}
			sm.SMchanels.Action <- SEND{0, vr}
			t := sm.getElectionTime()
			sm.SMchanels.Action <- ALARM{TIME: t}
		}
	}
}

func (sm *STATEMACHINE) getPrev() (int, int, string) { //this function returns previous log INDEX log TERM and previous command
	var prevI int
	var prevT int
	var prevC string
	for i := 0; i < len(sm.LOG); i++ {
		prevI = sm.LOG[i].INDEX
		prevT = sm.LOG[i].TERM
		prevC = string(sm.LOG[i].COMMAND)
	}
	return prevI, prevT, prevC
}

func (sm *STATEMACHINE) getElectionTime() int {
	switch sm.ID {
	case 0:
		return 2200
	case 1:
		return 2700
	case 2:
		return 3200
	case 3:
		return 3700
	case 4:
		return 4200
	}
	return 0
}
