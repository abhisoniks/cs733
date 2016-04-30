package main

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
		sm.MATCHINDEX[sm.ID] = sm.LOG[len(sm.LOG)-1].INDEX
		t := LOGENTRY{pINDEX + 1, pTERM, appendEVENT.A}
		l := LOGSTORE{INDEX: pINDEX + 1, DATA: t}
		sm.SMchanels.Action <- l
		apr := APPENDENTRIESREQ{TERM: sm.CURRENTTERM, LEADERID: sm.ID, PREVLOGINDEX: pINDEX, PREVLOGTERM: pTERM, ENTRIES: sm.LOG[len(sm.LOG)-2:], LEADERCOMMIT: sm.COMMITINDEX}
		sm.SMchanels.Action <- SEND{ID_TOSEND: 300, EVENT: apr}
		sm.SMchanels.Action <- ALARM{TIME: 100}
		//	sm.SMchanels.Action <- [3]string{"LOGSTORE(" + strconv.Itoa(pINDEX+1) + ")", "for each server p SEND(p,APPENDENTRIESREQ(" + strconv.Itoa(sm.CURRENTTERM) + "," + strconv.Itoa(0) + "," + strconv.Itoa(pINDEX) + "," + strconv.Itoa(pTERM) + "," + "log[" + strconv.Itoa(pINDEX) + ":]" + "," + strconv.Itoa(sm.COMMITINDEX) + ")", "ALARM(t)"}

	}
}
func (sm *STATEMACHINE) APPENDENTRIESREQ(AER_EVENT APPENDENTRIESREQ) {
	switch sm.STATE {
	case "Follower":
		prevINDEX, prevTERM, prevCommand := sm.getPrev()                                // previous Index of stateMachine log
		if AER_EVENT.TERM < sm.CURRENTTERM /*|| AER_EVENT.PREVLOGINDEX > prevINDEX */ { // leader's TERM is less than CURRENTTERM
			sm.SMchanels.Action <- SEND{ID_TOSEND: AER_EVENT.LEADERID, EVENT: APPENDENTRIESRESP{TERM: sm.CURRENTTERM, SENDERID: sm.ID, AERESP: false}}
			t := sm.getElectionTime()
			sm.SMchanels.Action <- ALARM{TIME: t}
			return
		} // if end
		if AER_EVENT.TERM >= sm.CURRENTTERM && AER_EVENT.ENTRIES == nil { //check if its a heartbeat check previous TERM and INDEX
			if prevINDEX == AER_EVENT.PREVLOGINDEX { //in both case reset Timeout
				t := sm.getElectionTime()
				sm.SMchanels.Action <- ALARM{TIME: t}
				sm.CURRENTTERM = AER_EVENT.TERM
				sm.LEADERID = AER_EVENT.LEADERID
				sm.COMMITINDEX = AER_EVENT.LEADERCOMMIT //update your COMMITIN
			} else {
				sm.SMchanels.Action <- SEND{ID_TOSEND: AER_EVENT.LEADERID, EVENT: APPENDENTRIESRESP{TERM: sm.CURRENTTERM, SENDERID: sm.ID, AERESP: false}}
				t := sm.getElectionTime()
				sm.SMchanels.Action <- ALARM{TIME: t}
			}
		} else { // this is not Heartbeat message.
			sm.LEADERID = AER_EVENT.LEADERID
			var match int
			var flag int
			match = 0
			flag = 0
			if AER_EVENT.PREVLOGINDEX <= 0 {
				copy(sm.LOG[:], AER_EVENT.ENTRIES)
				//sm.LOG = AER_EVENT.ENTRIES
				sm.COMMITINDEX = AER_EVENT.LEADERCOMMIT
				sm.SMchanels.Action <- SEND{ID_TOSEND: AER_EVENT.LEADERID, EVENT: APPENDENTRIESRESP{TERM: sm.CURRENTTERM, SENDERID: sm.ID, AERESP: true}}
				t := sm.getElectionTime()
				sm.SMchanels.Action <- ALARM{TIME: t}
				return
			}

			for i := len(AER_EVENT.ENTRIES) - 2; i >= 0; i-- {
				if AER_EVENT.ENTRIES[i].TERM != prevTERM || AER_EVENT.ENTRIES[i].INDEX != prevINDEX || prevCommand != string(AER_EVENT.ENTRIES[i].COMMAND) {

				} else {
					flag = 1
					match = i
				}
			} // end of for loop
			if flag == 0 {
				sm.SMchanels.Action <- SEND{ID_TOSEND: AER_EVENT.LEADERID, EVENT: APPENDENTRIESRESP{TERM: sm.CURRENTTERM, SENDERID: sm.ID, AERESP: false}}
				t := sm.getElectionTime()
				sm.SMchanels.Action <- ALARM{TIME: t}
				return
			}
			ti := sm.getElectionTime()
			sm.SMchanels.Action <- ALARM{TIME: ti}
			for k := match + 1; k < len(AER_EVENT.ENTRIES); k++ { // whereever matches copy whole log of leader and set CURRENTTERM equal to leader's TERM
				sm.LOG = append(sm.LOG[:prevINDEX+1], AER_EVENT.ENTRIES[k])
				prevINDEX = prevINDEX + 1
				sm.SMchanels.Action <- LOGSTORE{k, AER_EVENT.ENTRIES[k]}
			}
			ti = sm.getElectionTime()
			sm.SMchanels.Action <- ALARM{TIME: ti}
			sm.COMMITINDEX = AER_EVENT.LEADERCOMMIT
			sm.CURRENTTERM = AER_EVENT.TERM
			sm.VOTEDFOR = 6 // for this TERM votedFor is null
			sm.LEADERID = AER_EVENT.LEADERID
			sm.SMchanels.Action <- SEND{ID_TOSEND: AER_EVENT.LEADERID, EVENT: APPENDENTRIESRESP{TERM: sm.CURRENTTERM, SENDERID: sm.ID, AERESP: true}}

		}
	case "Leader": //

		if AER_EVENT.TERM < sm.CURRENTTERM { // leader's TERM is more than CURRENTTERM
			aer := APPENDENTRIESREQ{TERM: sm.CURRENTTERM, LEADERID: sm.ID, PREVLOGINDEX: sm.LOG[len(sm.LOG)-1].INDEX, PREVLOGTERM: sm.LOG[len(sm.LOG)-1].TERM, ENTRIES: nil, LEADERCOMMIT: sm.COMMITINDEX}
			sm.SMchanels.Action <- SEND{ID_TOSEND: AER_EVENT.LEADERID, EVENT: aer}
			sm.SMchanels.Action <- ALARM{TIME: 100}
		} else if AER_EVENT.TERM == sm.CURRENTTERM && AER_EVENT.LEADERCOMMIT < sm.COMMITINDEX {
			aer := APPENDENTRIESREQ{TERM: sm.CURRENTTERM, LEADERID: sm.ID, PREVLOGINDEX: sm.LOG[len(sm.LOG)-1].INDEX, PREVLOGTERM: sm.LOG[len(sm.LOG)-1].TERM, ENTRIES: nil, LEADERCOMMIT: sm.COMMITINDEX}
			sm.SMchanels.Action <- SEND{ID_TOSEND: AER_EVENT.LEADERID, EVENT: aer}
			sm.SMchanels.Action <- ALARM{TIME: 100}
		} else {
			if AER_EVENT.LEADERID != sm.ID {
				t := sm.getElectionTime()
				sm.SMchanels.Action <- ALARM{TIME: t}
				sm.VOTEDFOR = 6
				sm.STATE = "Follower"
				sm.CURRENTTERM = AER_EVENT.TERM
				sm.LEADERID = AER_EVENT.LEADERID
				sm.VOTEGRANTED = [2]int{0, 0}
				sm.SMchanels.Action <- SEND{ID_TOSEND: sm.LEADERID, EVENT: APPENDENTRIESRESP{TERM: sm.CURRENTTERM, SENDERID: sm.ID, AERESP: false}}
			}

		}
	case "Candidate":
		prevINDEX, prevTERM, prevCommand := sm.getPrev()
		if AER_EVENT.TERM < sm.CURRENTTERM { // leader's TERM is less than CURRENTTERM

			sm.SMchanels.Action <- SEND{ID_TOSEND: AER_EVENT.LEADERID, EVENT: APPENDENTRIESRESP{TERM: sm.CURRENTTERM, SENDERID: sm.ID, AERESP: false}}
			t := sm.getElectionTime()
			sm.SMchanels.Action <- ALARM{TIME: t}
			return
		}
		if AER_EVENT.TERM >= sm.CURRENTTERM && AER_EVENT.ENTRIES == nil { //check if its a heartbeat check previous TERM and INDEX
			if prevINDEX == AER_EVENT.PREVLOGINDEX { //in both case reset Timeout
				t := sm.getElectionTime()
				sm.STATE = "Follower"
				sm.SMchanels.Action <- ALARM{TIME: t}
				sm.CURRENTTERM = AER_EVENT.TERM
				sm.LEADERID = AER_EVENT.LEADERID
				sm.COMMITINDEX = AER_EVENT.LEADERCOMMIT //update your COMMITIN
			} else {

				sm.SMchanels.Action <- SEND{ID_TOSEND: sm.LEADERID, EVENT: APPENDENTRIESRESP{TERM: sm.CURRENTTERM, SENDERID: sm.ID, AERESP: false}}
				t := sm.getElectionTime()
				sm.SMchanels.Action <- ALARM{TIME: t}
			}
		} else { // this is not Heartbeat message.
			sm.LEADERID = AER_EVENT.LEADERID
			var match int
			var flag int
			match = 0
			flag = 0

			if AER_EVENT.PREVLOGINDEX <= 0 {

				sm.LOG = AER_EVENT.ENTRIES
				sm.COMMITINDEX = AER_EVENT.LEADERCOMMIT
				sm.SMchanels.Action <- SEND{ID_TOSEND: AER_EVENT.LEADERID, EVENT: APPENDENTRIESRESP{TERM: sm.CURRENTTERM, SENDERID: sm.ID, AERESP: true}}
				t := sm.getElectionTime()
				sm.SMchanels.Action <- ALARM{TIME: t}
				return
			}

			for i := len(AER_EVENT.ENTRIES) - 2; i >= 0; i-- {
				if AER_EVENT.ENTRIES[i].TERM != prevTERM || AER_EVENT.ENTRIES[i].INDEX != prevINDEX || prevCommand != string(AER_EVENT.ENTRIES[i].COMMAND) {

				} else {
					flag = 1
					match = i
				}
			} // end of for loop
			if flag == 0 {
				sm.SMchanels.Action <- SEND{ID_TOSEND: AER_EVENT.LEADERID, EVENT: APPENDENTRIESRESP{TERM: sm.CURRENTTERM, SENDERID: sm.ID, AERESP: false}}
				t := sm.getElectionTime()
				sm.SMchanels.Action <- ALARM{TIME: t}
				return
			}
			ti := sm.getElectionTime()
			sm.SMchanels.Action <- ALARM{TIME: ti}
			for k := match + 1; k < len(AER_EVENT.ENTRIES); k++ { // whereever matches copy whole log of leader and set CURRENTTERM equal to leader's TERM
				sm.LOG = append(sm.LOG[:prevINDEX+1], AER_EVENT.ENTRIES[k])
				sm.SMchanels.Action <- LOGSTORE{k, AER_EVENT.ENTRIES[k]}
			}
			ti = sm.getElectionTime()
			sm.SMchanels.Action <- ALARM{TIME: ti}
			sm.STATE = "Follower"
			sm.COMMITINDEX = AER_EVENT.LEADERCOMMIT
			sm.CURRENTTERM = AER_EVENT.TERM
			sm.VOTEDFOR = 6 // for this TERM votedFor is null
			sm.LEADERID = AER_EVENT.LEADERID
			sm.SMchanels.Action <- SEND{ID_TOSEND: AER_EVENT.LEADERID, EVENT: APPENDENTRIESRESP{TERM: sm.CURRENTTERM, SENDERID: sm.ID, AERESP: true}}
		}
	} //switch
} //function end

func (sm *STATEMACHINE) Timeout() {

	switch sm.STATE {
	case "Follower": //change state increase TERM by 1 and start election and set election timer
		sm.CURRENTTERM = sm.CURRENTTERM + 1
		pINDEX, pTERM, _ := sm.getPrev()
		sm.STATE = "Candidate"
		sm.VOTEDFOR = sm.ID // vote to itself
		sm.VOTEGRANTED[0] = 1
		VR := VOTEREQ{TERM: sm.CURRENTTERM, CANDIDATEID: sm.ID, LASTLOGINDEX: pINDEX, LASTLOGTERM: pTERM}
		sm.SMchanels.Action <- SEND{ID_TOSEND: 300, EVENT: VR}
		t := sm.getElectionTime()
		sm.SMchanels.Action <- ALARM{TIME: t}

	case "Leader":
		pINDEX, pTERM, _ := sm.getPrev()
		aer := APPENDENTRIESREQ{TERM: sm.CURRENTTERM, LEADERID: sm.ID, PREVLOGINDEX: pINDEX, PREVLOGTERM: pTERM, ENTRIES: nil, LEADERCOMMIT: sm.COMMITINDEX}
		sm.SMchanels.Action <- SEND{ID_TOSEND: 300, EVENT: aer}
		sm.SMchanels.Action <- ALARM{TIME: 100}
	case "Candidate":

		sm.CURRENTTERM = sm.CURRENTTERM + 1
		pINDEX, pTERM, _ := sm.getPrev()
		VR := VOTEREQ{TERM: sm.CURRENTTERM, CANDIDATEID: sm.ID, LASTLOGINDEX: pINDEX, LASTLOGTERM: pTERM}
		sm.SMchanels.Action <- SEND{ID_TOSEND: 300, EVENT: VR}
		t := sm.getElectionTime()
		sm.SMchanels.Action <- ALARM{TIME: t}
		sm.VOTEDFOR = sm.ID // vote to itself first
		sm.VOTEGRANTED[0] = 1

	}
}

func (sm *STATEMACHINE) APPENDENTRIESRESP(AER_EVENT APPENDENTRIESRESP) {

	switch sm.STATE {
	case "Follower":
		break //ignore it
	case "Candidate":
		break //ignore it
	case "Leader":
		if AER_EVENT.AERESP == true {
			lastINDEX := sm.LOG[len(sm.LOG)-1].INDEX
			sm.NEXTINDEX[AER_EVENT.SENDERID] = lastINDEX + 1
			sm.MATCHINDEX[AER_EVENT.SENDERID] = lastINDEX

		} else {
			if AER_EVENT.TERM > sm.CURRENTTERM {
				sm.CURRENTTERM = AER_EVENT.TERM
				newpINDEX := sm.LOG[len(sm.LOG)-1].INDEX
				newpTERM := sm.LOG[len(sm.LOG)-1].TERM
				aer := APPENDENTRIESREQ{TERM: sm.CURRENTTERM, LEADERID: sm.ID, PREVLOGINDEX: newpINDEX, PREVLOGTERM: newpTERM, ENTRIES: sm.LOG[(sm.NEXTINDEX[AER_EVENT.SENDERID]):], LEADERCOMMIT: sm.COMMITINDEX}
				sm.SMchanels.Action <- SEND{ID_TOSEND: AER_EVENT.SENDERID, EVENT: aer}
				return
				//	sm.SMchanels.Action <- ALARM{TIME: 100}
			}
			sm.NEXTINDEX[AER_EVENT.SENDERID] -= 1
			if sm.NEXTINDEX[AER_EVENT.SENDERID] <= 0 {
				aer := APPENDENTRIESREQ{TERM: sm.CURRENTTERM, LEADERID: sm.ID, PREVLOGINDEX: 0, PREVLOGTERM: 0, ENTRIES: sm.LOG, LEADERCOMMIT: sm.COMMITINDEX}
				sm.SMchanels.Action <- SEND{ID_TOSEND: AER_EVENT.SENDERID, EVENT: aer}
				sm.SMchanels.Action <- ALARM{TIME: 100}
				return
			}
			var newpINDEX, newpTERM int
			newpINDEX = sm.NEXTINDEX[AER_EVENT.SENDERID]
			newpTERM = sm.LOG[sm.NEXTINDEX[AER_EVENT.SENDERID]].TERM
			aer := APPENDENTRIESREQ{TERM: sm.CURRENTTERM, LEADERID: sm.ID, PREVLOGINDEX: newpINDEX, PREVLOGTERM: newpTERM, ENTRIES: sm.LOG[(sm.NEXTINDEX[AER_EVENT.SENDERID]):], LEADERCOMMIT: sm.COMMITINDEX}
			sm.SMchanels.Action <- SEND{ID_TOSEND: AER_EVENT.SENDERID, EVENT: aer}
			sm.SMchanels.Action <- ALARM{TIME: 100}

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

	pINDEX, _, _ := sm.getPrev() //will return previouscommand,previousId and previousTERM

	switch sm.STATE {
	case "Follower":
		if (sm.VOTEDFOR == 6) && VRQ_EVENT.TERM >= sm.CURRENTTERM {
			if VRQ_EVENT.LASTLOGINDEX >= pINDEX {
				sm.VOTEDFOR = VRQ_EVENT.CANDIDATEID
				sm.CURRENTTERM = VRQ_EVENT.TERM
				V_res := VOTERESP{TERM: sm.CURRENTTERM, VOTERESP: true}
				sm.SMchanels.Action <- SEND{ID_TOSEND: VRQ_EVENT.CANDIDATEID, EVENT: V_res}
				t := sm.getElectionTime()
				sm.SMchanels.Action <- ALARM{TIME: t}
			} else {
				V_res := VOTERESP{sm.CURRENTTERM, false}
				sm.SMchanels.Action <- SEND{VRQ_EVENT.CANDIDATEID, V_res}
			}
		} else {
			//sm.SMchanels.Action <- "SEND(" + cID + ",VOTERESP(" + cT + ",VOTEGRANTED=No))"
			V_res := VOTERESP{sm.CURRENTTERM, false}
			sm.SMchanels.Action <- SEND{VRQ_EVENT.CANDIDATEID, V_res}

		}

	case "Leader":

		if VRQ_EVENT.TERM > sm.CURRENTTERM {
			sm.VOTEGRANTED = [2]int{0, 0}
			sm.CURRENTTERM = VRQ_EVENT.TERM
			V_res := VOTERESP{sm.CURRENTTERM, true}
			sm.SMchanels.Action <- SEND{VRQ_EVENT.CANDIDATEID, V_res}
			t := sm.getElectionTime()
			sm.SMchanels.Action <- ALARM{TIME: t}
			sm.STATE = "Follower" // some of initialization before going in Follower state
			sm.VOTEDFOR = VRQ_EVENT.CANDIDATEID

		} else {
			V_res := VOTERESP{sm.CURRENTTERM, false}
			sm.SMchanels.Action <- SEND{VRQ_EVENT.CANDIDATEID, V_res}
		}

	case "Candidate":
		if VRQ_EVENT.TERM > sm.CURRENTTERM { //if voteFrom higher TERM
			sm.STATE = "Follower"

			sm.VOTEGRANTED = [2]int{0, 0}
			sm.VOTEDFOR = 6 //VRQ_EVENT.CANDIDATEID
			sm.CURRENTTERM = VRQ_EVENT.TERM
			V_res := VOTERESP{sm.CURRENTTERM, false}
			sm.SMchanels.Action <- SEND{VRQ_EVENT.CANDIDATEID, V_res}
			t := sm.getElectionTime()
			sm.SMchanels.Action <- ALARM{TIME: t}
		} else {
			V_res := VOTERESP{sm.CURRENTTERM, false}
			sm.SMchanels.Action <- SEND{VRQ_EVENT.CANDIDATEID, V_res}
		}
	}
}

func (sm *STATEMACHINE) VOTERESP(VRS_EVENT VOTERESP) {
	switch sm.STATE {
	case "Follower":
		return
	case "Leader":
		return
	case "Candidate":

		if VRS_EVENT.TERM > sm.CURRENTTERM {
			sm.STATE = "Follower"
			sm.VOTEDFOR = 6
			sm.CURRENTTERM = VRS_EVENT.TERM
			sm.VOTEGRANTED = [2]int{0, 0}
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
			aer := APPENDENTRIESREQ{TERM: sm.CURRENTTERM, LEADERID: sm.ID, PREVLOGINDEX: pINDEX, PREVLOGTERM: pTERM, ENTRIES: nil, LEADERCOMMIT: sm.COMMITINDEX}
			sm.SMchanels.Action <- SEND{300, aer}
			sm.SMchanels.Action <- ALARM{TIME: 100}
			sm.VOTEGRANTED = [2]int{0, 0} //reintialize VOTEGRANTED
			sm.VOTEDFOR = 6               //reintialize voteFor
			sm.STATE = "Leader"
			sm.LEADERID = sm.ID
			sm.NEXTINDEX = [5]int{len(sm.LOG), len(sm.LOG), len(sm.LOG), len(sm.LOG), len(sm.LOG)}
			sm.MATCHINDEX[sm.ID] = sm.LOG[len(sm.LOG)-1].INDEX

		}
		if sm.VOTEGRANTED[1] >= 3 { //become Follower here
			sm.VOTEGRANTED = [2]int{0, 0} //reintialize VOTEGRANTED
			sm.VOTEDFOR = 6               //reintialize voteFor
			sm.STATE = "Follower"
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
			sm.SMchanels.Action <- SEND{300, vr}
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
		return 800
	case 1:
		return 1000
	case 2:
		return 1200
	case 3:
		return 1400
	case 4:
		return 1600
	}
	return 0
}
