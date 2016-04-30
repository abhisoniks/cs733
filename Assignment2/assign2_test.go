package main

import (
	"fmt"
	"testing"
	//   "time"
	//   "reflect"
)

var sm StateMachine

func TestTCPSimple(t *testing.T) {
	//There are total of 5 machines with ids 0,1,2,3,4.
	//Our stateMachine has id of 0 and Machine starts in Follower state.Its own id is .
	// At start its votedFor is 6 it means it has not voted to anyone yet.because 6 is not associated with any machine's id
	//and at start leader_id is 4

	defaultSet() // This will put our Machine in Follower state

	// Machine is in Follower state Lets send it some Append request.
	app1 := Append{[]byte("delete file myfile")}
	clientCh <- app1
	sm.processEvent()
	response := <-Action
	res := response.(string)
	expect(t, res, "Commit(delete file myfile,Leader is at id 4)")

	//   Machine is in Follower state .Lets send it AppendEntriesRequest..
	var log1 = make([]logEntry, 3)
	log1[0] = logEntry{0, 1, "append"}
	log1[1] = logEntry{1, 1, "delete"}
	log1[2] = logEntry{2, 1, "append"}
	appER1 := AppendEntriesReq{term: 1, leaderId: 4, prevLogIndex: 2, prevLogTerm: 1, entries: log1, leaderCommit: 2}
	netCh <- appER1
	sm.processEvent()
	response1 := <-Action
	res1 := response1.([4]string)
	expect(t, res1[0], "true") // true or false
	expect(t, res1[1], "1")    //term for which this response is sent
	expect(t, res1[2], "0")    //sender's id
	expect(t, res1[3], "Alarm(t)")

	// Machine is still in Follower state.Its Previndex is 2,term 1,Entry at Previndex is "append" ,leader id is 4.
	//lets send an Event of AppendRequest
	// with empty log   AppendEntriesReq(term int,leader_id int, prevLogIndex int,prevLogTerm int,entries[] string,leaderCommit int)
	//since prevIndex and PrevTerm will match in this case expected Action is true from && it will reset its Timeout Period
	//expected action in this test case is an Array of string {true,term,sender's id,Alarm(t)}

	appER := AppendEntriesReq{term: 1, leaderId: 4, prevLogIndex: 2, prevLogTerm: 1, entries: nil, leaderCommit: 2}
	netCh <- appER
	sm.processEvent()
	response1 = <-Action
	res1 = response1.([4]string)
	expect(t, res1[0], "true")
	expect(t, res1[1], "1")
	expect(t, res1[2], "0")
	expect(t, res1[3], "Alarm(t)")

	// Machine is still in Follower state.Its Previndex is 2,term 1,Entry at Previndex is "cas" ,leader id is 50.
	//lets send an Event of AppendRequest but this time term number of Leader would be less than statemachine(lets 0)
	//since leader's term is less than own term stateMachine will send false ,currentTerm  and reset Timer
	//expected action in this test case is an Array of string {true,1,Alarm(t)}

	appER = AppendEntriesReq{term: 0, leaderId: 4, prevLogIndex: 2, prevLogTerm: 0, entries: nil, leaderCommit: 2}
	netCh <- appER
	sm.processEvent()
	response = <-Action
	res2 := response.([4]string)
	expect(t, res2[0], "false")
	expect(t, res2[1], "1")
	expect(t, res2[2], "0")
	expect(t, res2[3], "Alarm(t)")

	//Test case of heartbeat message where Term is same but Index is lower

	appER = AppendEntriesReq{term: 1, leaderId: 4, prevLogIndex: 1, prevLogTerm: 0, entries: nil, leaderCommit: 1}
	netCh <- appER
	sm.processEvent()
	response41 := <-Action
	res41 := response41.([4]string)
	expect(t, res41[0], "false")
	expect(t, res41[1], "1")
	expect(t, res41[2], "0")
	expect(t, res41[3], "Alarm(t)")

	//fmt.Println("log1 is 1", sm.log[0],sm.log[1],sm.log[2])
	// current log of Follower= [{0,1,"append"} {1,1,"delete"} {2,1,"append"}].Lets send it some AppendEntriesReq
	//since everything is ok it should match prev index and prev term and should reply in true and term number
	//log1[3]=logEntry{3,2,"CSE"}

	log1 = append(log1, logEntry{3, 2, "CSE"})
	appER = AppendEntriesReq{term: 2, leaderId: 4, prevLogIndex: 2, prevLogTerm: 1, entries: log1, leaderCommit: 2}
	netCh <- appER
	sm.processEvent()
	response3 := <-Action
	res3 := response3.([4]string)
	expect(t, res3[0], "true")
	expect(t, res3[1], "2")
	expect(t, res3[2], "0")
	expect(t, res3[3], "Alarm(t)")
	//   lets leader gets 3 new entries at once .And it send these three entries at once to state Mahcine.
	// 	 leader new log would be [{0,1,"append"} {1,1,"delete"} {2,1,"append"},{3,2,"CSE"} {3,2,"cloud733"} {4,2,"cse43"} {5,2,"cse678"}]
	//Follower log should be at [{0,1,"append"} {1,1,"delete"} {2,1,"append"},{3,2,"CSE"}] after previous test case

	log1 = append(log1, logEntry{4, 2, "cloud733"})
	log1 = append(log1, logEntry{5, 2, "cse43"})
	log1 = append(log1, logEntry{6, 2, "cse678"})
	appER = AppendEntriesReq{term: 2, leaderId: 4, prevLogIndex: 3, prevLogTerm: 2, entries: log1, leaderCommit: 3}
	netCh <- appER
	sm.processEvent()
	response4 := <-Action
	res4 := response4.([4]string)

	expect(t, res4[0], "true")
	expect(t, res4[1], "2")
	expect(t, res4[2], "0")
	expect(t, res4[3], "Alarm(t)")

	//After previous test case State Machine's Log would be	[{0,1,"append"} {1,1,"delete"} {2,1,"append"},{3,2,"CSE"} {3,2,"cloud733"} {4,2,"cse43"} {5,2,"cse678"}]
	//Now lets send it Heartbeat message and  tell it to commit previously sent item
	appER = AppendEntriesReq{term: 2, leaderId: 4, prevLogIndex: 6, prevLogTerm: 2, entries: nil, leaderCommit: 6}
	netCh <- appER
	sm.processEvent()
	response5 := <-Action
	res5 := response5.([4]string)
	expect(t, res5[0], "true")
	expect(t, res5[1], "2")
	expect(t, res5[2], "0")
	expect(t, res5[3], "Alarm(t)")

	//Now lets discuss a case where Leader's previous entry does not match with stateMachine's previous entry.but term and index is higher.
	//state Machine Log is [{0,1,"append"} {1,1,"delete"} {2,1,"append"},{3,2,"CSE"} {3,2,"cloud733"} {4,2,"cse43"} {5,2,"cse678"}]
	//Let Leader log has all these entry and some new entry say {6,3,"cloudAssign1"},{7,3,"cloudAssign2"},{8,3,"Asign3"} and leader's
	// commit index is 8.In this case expected output is false as prevIndex will not match

	log1 = append(log1, logEntry{7, 3, "cloudAssign1"})
	log1 = append(log1, logEntry{8, 3, "cloudAssign2"})
	log1 = append(log1, logEntry{9, 3, "cloudAssign3"})
	appER = AppendEntriesReq{term: 3, leaderId: 4, prevLogIndex: 8, prevLogTerm: 3, entries: log1, leaderCommit: 8}
	netCh <- appER
	sm.processEvent()
	response6 := <-Action
	res6 := response6.([4]string)
	expect(t, res6[0], "false")
	expect(t, res6[1], "2")
	expect(t, res6[2], "0")
	expect(t, res6[3], "Alarm(t)")

	//Now leader will decrease its prevIndex.and try again.This time also it should fail

	appER = AppendEntriesReq{term: 3, leaderId: 4, prevLogIndex: 7, prevLogTerm: 3, entries: log1, leaderCommit: 8}
	netCh <- appER
	sm.processEvent()
	response7 := <-Action
	res7 := response7.([4]string)
	expect(t, res7[0], "false")
	expect(t, res7[1], "2")
	expect(t, res7[2], "0")
	expect(t, res7[3], "Alarm(t)")

	//Now leader will decrease its prevIndex.and try again.This time also it's log will match.state Machine will copy all data
	//and will reply with true 	and will append all entries that follow it
	appER = AppendEntriesReq{term: 3, leaderId: 4, prevLogIndex: 6, prevLogTerm: 2, entries: log1, leaderCommit: 8}
	netCh <- appER
	sm.processEvent()
	response8 := <-Action
	res8 := response8.([4]string)
	expect(t, res8[0], "true")
	expect(t, res8[1], "3")
	expect(t, res8[2], "0")
	expect(t, res8[3], "Alarm(t)")

	//fmt.Println("new Commit Index is",sm.commitIndex);

	// Lets send voteReq to this machine .Since this request's term is less than stateMachine's term.So reply should be false
	VR := VoteReq{term: 3, candidateId: 3, lastLogIndex: 8, lastLogTerm: 2} //candidate’s term
	netCh <- VR
	sm.processEvent()
	response9 := <-Action
	res9 := response9.(string)
	expect(t, res9, "Send(3,VoteResp(3,voteGranted=No))")

	//	 	Lets again send this machine voteReq with same term number but index number which is less than it.
	VR = VoteReq{term: 3, candidateId: 2, lastLogIndex: 6, lastLogTerm: 3}
	netCh <- VR
	sm.processEvent()
	response = <-Action
	res = response.(string)
	expect(t, res, "Send(2,VoteResp(3,voteGranted=No))")

	//      Lets again send this machine voteReq with same term number and with updated log
	VR = VoteReq{term: 3, candidateId: 1, lastLogIndex: 9, lastLogTerm: 3}
	netCh <- VR
	sm.processEvent()
	response = <-Action
	res = response.(string)
	expect(t, res, "Send(1,VoteResp(3,voteGranted=yes))")

	//since state_machine has voted for this term it should not give vote to anyother candidate in this term
	VR = VoteReq{term: 3, candidateId: 3, lastLogIndex: 11, lastLogTerm: 3}
	netCh <- VR
	sm.processEvent()
	response = <-Action
	res = response.(string)
	expect(t, res, "Send(3,VoteResp(3,voteGranted=No))")

	//Lets send AppendEntriesRequest to it and increse its term by one and then in next term it should give vote

	log1 = append(log1, logEntry{10, 4, "cloudProject1"})
	log1 = append(log1, logEntry{11, 4, "cloudProject2"})

	appER = AppendEntriesReq{term: 4, leaderId: 4, prevLogIndex: 9, prevLogTerm: 3, entries: log1, leaderCommit: 9}
	netCh <- appER
	sm.processEvent()
	response12 := <-Action
	res8 = response12.([4]string)
	expect(t, res8[0], "true")
	expect(t, res8[1], "4")
	expect(t, res8[2], "0")
	expect(t, res8[3], "Alarm(t)")
	//lets send it voteReq again .Since term of Machine has changed it should give vote to candidate if candidate fullfils requirement
	VR = VoteReq{term: 4, candidateId: 3, lastLogIndex: 11, lastLogTerm: 4} //candidate’s term
	netCh <- VR
	sm.processEvent()
	response = <-Action
	res = response.(string)
	expect(t, res, "Send(3,VoteResp(4,voteGranted=yes))") //Send("+cID+",VoteResp("+cT+",voteGranted=No)

	// Machine is in Follower state.Lets throw Timeout event on it.In action it should change its state to candidate,
	// increase its currentTerm(from 4 to 5) and should send voterReq to all
	// expected Action in this test case is Array of string("for all peer p","send(p,VoteReq(2,CandId,2,1))")

	TimeO := Timeout{}
	timeoutCh <- TimeO
	sm.processEvent()
	response13 := <-Action
	res13 := response13.([2]string)
	expect(t, res13[0], "for all peer p send(p,VoteReq(5,0,11,4))")
	// expect(t,res13[1],"")
	expect(t, res13[1], "Alarm(t)") // reset election timeout

	//Machine is in Candidate state .Lets send Timeout again.It would send same action but term would increase by 1
	TimeO = Timeout{}
	timeoutCh <- TimeO
	sm.processEvent()
	response13 = <-Action
	res13 = response13.([2]string)
	expect(t, res13[0], "for all peer p send(p,VoteReq(6,0,11,4))")
	//expect(t,res13[1],"")
	expect(t, res13[1], "Alarm(t)") // reset election timeout

	//Lets try Append on Candidate
	app1 = Append{[]byte("create a file")}
	clientCh <- app1
	sm.processEvent()
	response27 := <-Action
	res27 := response27.(string)
	expect(t, res27, "Commit(create a file,error)")

	//Machine is in candidate state.Lets send it a voteReq with higher term .
	//It will copy candidate's term and will vote it

	VR = VoteReq{term: 8, candidateId: 2, lastLogIndex: 11, lastLogTerm: 4} //candidate’s term
	netCh <- VR
	sm.processEvent()
	response20 := <-Action
	res20 := response20.([2]string)
	expect(t, res20[0], "Send(2,VoteResp(8,voteGranted=yes))") //Send("+cID+",VoteResp("+cT+",voteGranted=No)
	expect(t, res20[1], "Alarm(t)")

	//Lets again bring it in candidate state
	TimeO = Timeout{}
	timeoutCh <- TimeO
	sm.processEvent()
	response21 := <-Action
	res21 := response21.([2]string)
	expect(t, res21[0], "for all peer p send(p,VoteReq(9,0,11,4))")
	//expect(t,res13[1],"")
	expect(t, res21[1], "Alarm(t)") // reset election timeout

	//Its in candidate state.send it AppendEntriesRPC whose term is less than it

	appER = AppendEntriesReq{term: 3, leaderId: 2, prevLogIndex: 9, prevLogTerm: 3, entries: nil, leaderCommit: 9}
	netCh <- appER
	sm.processEvent()
	response23 := <-Action
	res23 := response23.([4]string)
	expect(t, res23[0], "false")
	expect(t, res23[1], "9")
	expect(t, res23[2], "0")
	expect(t, res23[3], "Alarm(t)")

	//send it AppendEntriesRPC whose term is more than it.Copy term number and go in follower state and

	appER = AppendEntriesReq{term: 10, leaderId: 2, prevLogIndex: 11, prevLogTerm: 4, entries: nil, leaderCommit: 9}
	netCh <- appER
	sm.processEvent()
	response24 := <-Action
	res24 := response24.([4]string)
	expect(t, res24[0], "true")
	expect(t, res24[1], "10")
	expect(t, res23[2], "0")
	expect(t, res24[3], "Alarm(t)")

	//Again bring it in candidate state

	TimeO = Timeout{}
	timeoutCh <- TimeO
	sm.processEvent()
	response25 := <-Action
	res25 := response25.([2]string)
	expect(t, res25[0], "for all peer p send(p,VoteReq(11,0,11,4))")
	//expect(t,res13[1],"")
	expect(t, res25[1], "Alarm(t)") // reset election timeout

	//In previous test case it has sent voteReq to server.Lets check case where it gets voteResponse from higher term.
	//It will get shift to Follower state reintialize its variable and
	// will send an Action="Alarm(t)"

	voteRS := VoteResp{term: 12, voteResp: false}
	netCh <- voteRS
	sm.processEvent()
	response14 := <-Action
	res14 := response14.(string)
	expect(t, res14, "Alarm(t)")
	// Due to last test case system is in Follower state.Lets again get it in candidate state.

	TimeO = Timeout{}
	timeoutCh <- TimeO
	sm.processEvent()
	response13 = <-Action
	res13 = response13.([2]string)
	expect(t, res13[0], "for all peer p send(p,VoteReq(13,0,11,4))")
	//expect(t,res13[1],"")
	expect(t, res13[1], "Alarm(t)") // reset election timeout

	// Now state Machine is getting voteRespose for previous voteRequest from various server
	//In following test case we are checking case where it will result in voteSplit and reElection will take place
	voteRS = VoteResp{term: 5, voteResp: true}
	netCh <- voteRS
	sm.processEvent()

	voteRS = VoteResp{term: 6, voteResp: false}
	netCh <- voteRS
	sm.processEvent()

	voteRS = VoteResp{term: 5, voteResp: true}
	netCh <- voteRS
	sm.processEvent()
	voteRS = VoteResp{term: 9, voteResp: false}
	netCh <- voteRS
	sm.processEvent()

	response15 := <-Action
	res15 := response15.([2]string)
	expect(t, res15[0], "for all peer p send(p,VoteReq(14,0,11,4))")
	//expect(t,res13[1],"")
	expect(t, res15[1], "Alarm(t)") // reset election timeout

	//Again its getting voteResp from different different server.Its term is 9 here.

	//In following test case since its getting false from Majority .it should become follower and reset its timer
	voteRS = VoteResp{term: 8, voteResp: false}
	netCh <- voteRS
	sm.processEvent()

	voteRS = VoteResp{term: 9, voteResp: false}
	netCh <- voteRS
	sm.processEvent()

	voteRS = VoteResp{term: 4, voteResp: false}
	netCh <- voteRS
	sm.processEvent()

	response17 := <-Action
	res17 := response17.(string)
	expect(t, res17, "Alarm(t)") // reset election timeout
	// Here since with first 3 voteResponse it has become Follower.In follower state even if it get VoteResponse of 4th server it will not take anyAction
	//It will simply ignore it .
	voteRS = VoteResp{term: 8, voteResp: false}
	netCh <- voteRS
	sm.processEvent()

	//lets now check a test case where it will become leader after winning election.
	//Right now our machine is in Follower state First take it in candidate state again  by following event

	TimeO = Timeout{}
	timeoutCh <- TimeO
	sm.processEvent()
	response18 := <-Action
	res18 := response18.([2]string)
	expect(t, res18[0], "for all peer p send(p,VoteReq(15,0,11,4))")
	expect(t, res18[1], "Alarm(t)") // reset election timeout

	//	 	 In following test case it will get true from Majority of server and will go in Leader state
	netCh <- voteRS
	sm.processEvent()

	voteRS = VoteResp{term: 8, voteResp: true}
	netCh <- voteRS
	sm.processEvent()

	voteRS = VoteResp{term: 9, voteResp: true}
	netCh <- voteRS
	sm.processEvent()

	voteRS = VoteResp{term: 4, voteResp: true}
	netCh <- voteRS
	sm.processEvent()

	response19 := <-Action
	res19 := response19.([2]string)
	expect(t, res19[0], "for all peer p send(p,AppendEntriesReq(15,0,11,4,nil,9)")
	expect(t, res19[1], "Alarm(t)") // reset election timeout

	//     sm is in leader state parameter => term:15,leader_id:0,previousIndex:11,previousTerm:4 and commitIndex is 9

	var log2 = make([]logEntry, 0)
	//log2[0]=logEntry{0,1,"append"}
	log2 = append(log2, logEntry{0, 1, "append"})
	log2 = append(log2, logEntry{1, 1, "delete"})
	log2 = append(log2, logEntry{2, 1, "append"})
	log2 = append(log2, logEntry{3, 2, "cas"})
	log2 = append(log2, logEntry{4, 2, "read"})
	log2 = append(log2, logEntry{5, 3, "write"})
	log2 = append(log2, logEntry{6, 3, "concat"})
	log2 = append(log2, logEntry{7, 4, "move"})
	log2 = append(log2, logEntry{8, 5, "copy"})

	sm = StateMachine{state: "Leader", leaderId: 0, votedFor: 6, currentTerm: 5, log: log2, commitIndex: 6, lastApplied: 6, voteGranted: [2]int{0, 0}}
	sm.nextIndex = [5]int{9, 9, 9, 9, 9}
	sm.matchIndex = [5]int{6, 4, 7, 6, 3}

	app2 := Append{[]byte("cas myfile")}
	clientCh <- app2
	sm.processEvent()
	response31 := <-Action
	res26 := response31.([3]string)
	expect(t, res26[0], "LogStore(9)")
	expect(t, res26[1], "for each server p send(p,AppendEntriesReq(5,0,8,5,log[8:],6)")
	expect(t, res26[2], "Alarm(t)")

	//AppendEntriesResponse from one_server with Id 2
	AERs := AppendEntriesResp{3, 2, false}
	netCh <- AERs
	sm.processEvent()
	//   fmt.Println(sm.state);

	response = <-Action
	res28 := response.([2]string)
	expect(t, res28[0], "Send(2,AppendEntriesReq(5,0,7,4,log[7:],6)")
	expect(t, res28[1], "Alarm(t)")
	//same server again send it false
	AERs = AppendEntriesResp{3, 2, false}
	netCh <- AERs
	sm.processEvent()
	response = <-Action
	res29 := response.([2]string)
	expect(t, res29[0], "Send(2,AppendEntriesReq(5,0,6,3,log[6:],6)")
	expect(t, res29[1], "Alarm(t)")
	//This time it send true.Leader will update its matchIndex and nextIndex for it.But it will not get any action
	AERs = AppendEntriesResp{6, 2, true}
	netCh <- AERs
	sm.processEvent()
	//fmt.Println("new next index is expected to be 10 and it is",sm2.nextIndex[2]);
	//fmt.Println("new next index is expected to be 9 and it is",sm2.matchIndex[2]);
	//Lets it gets true from other server
	AERs = AppendEntriesResp{6, 1, true}
	netCh <- AERs
	sm.processEvent()

	AERs = AppendEntriesResp{6, 3, true}
	netCh <- AERs
	sm.processEvent()
	response38 := <-Action
	res38 := response38.(string)
	expect(t, res38, "Commit(9,cas myfile)")

	//From Majority of server it got true.So In next heartbeat message its leaderCommit should be updated
	TimeO = Timeout{}
	timeoutCh <- TimeO
	sm.processEvent()
	response18 = <-Action
	res18 = response18.([2]string)
	expect(t, res18[0], "for all peer p send(p,AppendEntriesReq(5,0,9,5,nil,9))")
	expect(t, res18[1], "Alarm(t)") // reset election timeout

	//Again Append Request to Leader
	app2 = Append{[]byte("rename myfile")}
	clientCh <- app2
	sm.processEvent()
	response31 = <-Action
	res26 = response31.([3]string)
	expect(t, res26[0], "LogStore(10)")
	expect(t, res26[1], "for each server p send(p,AppendEntriesReq(5,0,9,5,log[9:],9)")
	expect(t, res26[2], "Alarm(t)")
	//Lets it get negative response from majority of server
	AERs = AppendEntriesResp{6, 2, true}
	netCh <- AERs
	sm.processEvent()

	AERs = AppendEntriesResp{6, 3, false}
	netCh <- AERs
	sm.processEvent()
	response33 := <-Action
	res33 := response33.([2]string)
	expect(t, res33[0], "Send(3,AppendEntriesReq(5,0,8,5,log[8:],9)")
	expect(t, res33[1], "Alarm(t)")

	AERs = AppendEntriesResp{7, 4, false}
	netCh <- AERs
	sm.processEvent()
	response34 := <-Action
	res34 := response34.([2]string)
	expect(t, res34[0], "Send(4,AppendEntriesReq(5,0,7,4,log[7:],9)") //nextIndex of 4th server
	expect(t, res34[1], "Alarm(t)")

	AERs = AppendEntriesResp{6, 1, false}
	netCh <- AERs
	sm.processEvent()
	response37 := <-Action
	res37 := response37.([2]string)
	expect(t, res37[0], "Send(1,AppendEntriesReq(5,0,8,5,log[8:],9)")
	expect(t, res37[1], "Alarm(t)")

	//In next heartbeat Message Leader's commit Index shold be same
	TimeO = Timeout{}
	timeoutCh <- TimeO
	sm.processEvent()
	response39 := <-Action
	res39 := response39.([2]string)
	expect(t, res39[0], "for all peer p send(p,AppendEntriesReq(5,0,10,5,nil,9))")
	expect(t, res39[1], "Alarm(t)") // reset election timeout
	// VoteReq from server with lower term
	VR = VoteReq{term: 3, candidateId: 2, lastLogIndex: 6, lastLogTerm: 3} //candidate’s term
	netCh <- VR
	sm.processEvent()
	response9 = <-Action
	res9 = response9.(string)
	expect(t, res9, "Send(2,VoteResp(5,voteGranted=No))")
	//VoteReq from server with higher term
	VR = VoteReq{term: 6, candidateId: 3, lastLogIndex: 9, lastLogTerm: 5} //candidate’s term
	netCh <- VR
	sm.processEvent()
	response9 = <-Action
	res9 = response9.(string)
	expect(t, res9, "Send(3,VoteResp(5,voteGranted=No))")

	// Lets again intialize statemachine to Leader state
	var log3 = make([]logEntry, 0)
	for i := 0; i < len(sm.log); i++ {
		log3 = append(log3, sm.log[i])
	}
	sm = StateMachine{state: "Leader", leaderId: 0, votedFor: 6, currentTerm: 5, log: log3, commitIndex: 9, lastApplied: 6, voteGranted: [2]int{0, 0}}
	//Leader gets AppendEntries Request from a lower term
	appER = AppendEntriesReq{term: 2, leaderId: 4, prevLogIndex: 6, prevLogTerm: 3, entries: nil, leaderCommit: 4}
	netCh <- appER
	sm.processEvent()
	response = <-Action
	res2 = response.([4]string)
	expect(t, res2[0], "false")
	expect(t, res2[1], "5")
	expect(t, res2[2], "0")
	expect(t, res2[3], "Alarm(t)")

	// Leader receives VoteResponse  from a server.Leader will not respond to it
	voteRS = VoteResp{term: 8, voteResp: true}
	netCh <- voteRS
	sm.processEvent()

	//Leader gets AppendEntries Request from a higher term.It will shift in Follower state
	appER = AppendEntriesReq{term: 6, leaderId: 4, prevLogIndex: 6, prevLogTerm: 3, entries: nil, leaderCommit: 4}
	netCh <- appER
	sm.processEvent()
	response = <-Action
	res2 = response.([4]string)
	expect(t, res2[0], "false")
	expect(t, res2[1], "5")
	expect(t, res2[2], "0")
	expect(t, res2[3], "Alarm(t)")

	//machine is in follower state now.Lets send it AppendEntriesResponse.It will ignore it
	AERs = AppendEntriesResp{6, 1, false}
	netCh <- AERs
	sm.processEvent()

	//A test case in Follower state.where term matches but Index dont matches
	appER = AppendEntriesReq{term: 5, leaderId: 4, prevLogIndex: 5, prevLogTerm: 5, entries: log1, leaderCommit: 8}
	netCh <- appER
	sm.processEvent()
	response6 = <-Action
	res6 = response6.([4]string)
	expect(t, res6[0], "false")
	expect(t, res6[1], "5")
	expect(t, res6[2], "0")
	expect(t, res6[3], "Alarm(t)")

	// A test case previous Index matches & term matches but command does not Maches .

	copy(log1, sm.log)
	log1[10] = logEntry{10, 5, "changeName"}
	appER = AppendEntriesReq{term: 5, leaderId: 4, prevLogIndex: 10, prevLogTerm: 5, entries: log1, leaderCommit: 8}

	netCh <- appER
	sm.processEvent()
	response6 = <-Action
	res6 = response6.([4]string)
	expect(t, res6[0], "false")
	expect(t, res6[1], "5")
	expect(t, res6[2], "0")
	expect(t, res6[3], "Alarm(t)")

	//Lets change its state to candidate state by Timeout event .In candidate state It will ask for vote firstof ALL
	TimeO = Timeout{}
	timeoutCh <- TimeO
	sm.processEvent()
	response39 = <-Action
	res39 = response39.([2]string)
	expect(t, res39[0], "for all peer p send(p,VoteReq(6,0,10,5))")
	expect(t, res39[1], "Alarm(t)")
	//In candidate state Lets cover a case where its get voteReques from a lower term
	VR = VoteReq{term: 5, candidateId: 3, lastLogIndex: 7, lastLogTerm: 5} //candidate’s term
	netCh <- VR
	sm.processEvent()
	response9 = <-Action
	res9 = response9.(string)
	expect(t, res9, "Send(3,VoteResp(6,voteGranted=No))")

	//Test case where a candidate gets AppendEntries Response.It will ignore it.

	AERs = AppendEntriesResp{7, 4, false}
	netCh <- AERs
	sm.processEvent()

	//Now lets condier various AppendEntriesReq for a candidate state
	//Lets consider a case where it gets heartbeat message from equal term But its log is not as updated as its own
	appER = AppendEntriesReq{term: 6, leaderId: 4, prevLogIndex: 5, prevLogTerm: 5, entries: nil, leaderCommit: 4}
	netCh <- appER
	sm.processEvent()
	response = <-Action
	res2 = response.([4]string)
	expect(t, res2[0], "false")
	expect(t, res2[1], "6")
	expect(t, res2[2], "0")
	expect(t, res2[3], "Alarm(t)")

	//Lets consider a AppendEntriesReq whose previous term dont matches with its own previous term
	var log4 = make([]logEntry, 0)
	copy(log4, log1)
	log1[5].term = 2

	appER = AppendEntriesReq{term: 6, leaderId: 4, prevLogIndex: 5, prevLogTerm: 2, entries: log1, leaderCommit: 4}

	netCh <- appER
	sm.processEvent()
	response = <-Action
	res2 = response.([4]string)
	expect(t, res2[0], "false")
	expect(t, res2[1], "6")
	expect(t, res2[2], "0")
	expect(t, res2[3], "Alarm(t)")

	//Lets consider a case where both term and index matches but its command at that index dont matches .in this case it will return false
	log1[11].command = "create"
	appER = AppendEntriesReq{term: 6, leaderId: 4, prevLogIndex: 10, prevLogTerm: 5, entries: log1, leaderCommit: 4}
	netCh <- appER
	sm.processEvent()
	response = <-Action
	res2 = response.([4]string)
	expect(t, res2[0], "false")
	expect(t, res2[1], "6")
	expect(t, res2[2], "0")
	expect(t, res2[3], "Alarm(t)")

	//A test case where incoming AppendEntriesRequest is valid.index,command and term Matches.
	//For this it will append all the log of Leader and will shift to follower state

	log1[10].command = "rename myfile"
	log1 = append(log1, logEntry{11, 6, "newentry1"})
	log1 = append(log1, logEntry{12, 6, "newentry1"})
	log1 = append(log1, logEntry{13, 7, "newentry1"})
	appER = AppendEntriesReq{term: 6, leaderId: 4, prevLogIndex: 10, prevLogTerm: 5, entries: log1, leaderCommit: 4}
	netCh <- appER
	sm.processEvent()
	response = <-Action
	res2 = response.([4]string)
	expect(t, res2[0], "true")
	expect(t, res2[1], "6")
	expect(t, res2[2], "0")
	expect(t, res2[3], "Alarm(t)")

	//Lets consider a case where term,index and command matches .In this case it will copy all log starting from that index
	/* fmt.Println(log1)
	log1=append(log1,logEntry{12,5,"newCommand"})
	appER = AppendEntriesReq{term: 6, leaderId: 4, prevLogIndex: 10, prevLogTerm: 5, entries: log1, leaderCommit: 4}
	netCh <- appER
	sm.processEvent()
	response = <-Action
	res2 = response.([4]string)
	expect(t, res2[0], "true")
	expect(t, res2[1], "6")
	expect(t, res2[2], "0")
	expect(t, res2[3], "Alarm(t)") */

}

func expect(t *testing.T, a string, b string) {
	if a != b {
		t.Error(fmt.Sprintf("Expected %v, found %v", b, a)) // t.Error is visible when running `go test -verbose`
	}
}
