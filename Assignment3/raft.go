package main

import (
	"encoding/gob"
	"fmt"
	"github.com/cs733-iitb/cluster"
	"github.com/cs733-iitb/log"
	"strings"
	"time"
)

//var logDir *log.Log
func (rn *RaftNode) Append(data string) { //This functioin handles every Append request coming to this node
	rn.sm.SMchanels.clientCh <- APPEND{[]byte(data)} //It passes this Append request to stateMAchine through clientCh channel
}

func (rn *RaftNode) CommittedIndex() int { //This function returns commited Index
	return rn.sm.COMMITINDEX
}

func checkError(e error) {
	if e != nil {
		fmt.Println(e)
	}
}

// makerafts function will make 5 rodeNodes .It will associate a stateMAchine to all nodes.
func makerafts() []*RaftNode {
	raftset := make([]*RaftNode, 5)
	for i := 0; i < 5; i++ {
		raft, err := cluster.New(i, "peer.json")
		checkError(err)
		node := initialize(raft, i)
		raftset[i] = node
	}
	return raftset
}

//Initialize function initialize all RaftNode related structure and also Initialize RaftNode.It will initialize STATEMACHINE
//This function starts stateMachine, timer of Node ,a routine of checking Inbox and a routine for checking Action coming from stateMAchine

func initialize(serv cluster.Server, id int) *RaftNode {

	gob.Register(VOTERESP{}) // register a struct name by giving it a dummy object of that name.
	gob.Register(VOTEREQ{})
	gob.Register(APPENDENTRIESREQ{})
	gob.Register(APPENDENTRIESRESP{})
	gob.Register(LOGENTRY{})
	gob.Register(COMMIT{})
	gob.Register(SEND{})
	ch0 := new(Channel) //Make channels between SM and node
	ch0.Action = make(chan interface{}, 40)
	ch0.clientCh = make(chan interface{}, 40)
	ch0.timeoutCh = make(chan interface{}, 40)
	ch0.netCh = make(chan interface{}, 40)
	var raftN *RaftNode
	var temp_sm *STATEMACHINE
	var raft_config *Config
	switch id {
	case 0:
		raft_config = &Config{0, "log0", 800, 100}                                                                                                       //Initialize config of node
		temp_sm = &STATEMACHINE{ID: 0, VOTEDFOR: 6, STATE: "Follower", SMchanels: ch0, NEXTINDEX: [5]int{2, 2, 2, 2, 2}, COMMITINDEX: 1, CURRENTTERM: 0} ///Initialize stateMAchine
		logDir0, err := log.Open("log0")                                                                                                                 //Open Log file related to this stateMachine
		checkError(err)
		raftN = &RaftNode{config: raft_config, timerFlag: 1, sm: temp_sm, logDir: logDir0} //FInally Make Node
		//raftN.enterSamplentry();
	case 1:
		raft_config = &Config{Id: 1, LogDir: "log1", ElectionTimeout: 1000, HeartbeatTimeout: 100}
		temp_sm = &STATEMACHINE{ID: 1, VOTEDFOR: 6, STATE: "Follower", SMchanels: ch0, NEXTINDEX: [5]int{2, 2, 2, 2, 2}, COMMITINDEX: 1, CURRENTTERM: 0}
		//Intialize RaftNode
		logDir1, err := log.Open("log1") //Open Log file
		checkError(err)
		raftN = &RaftNode{config: raft_config, timerFlag: 1, sm: temp_sm, logDir: logDir1} //FInally Make Node
		//raftN.enterSamplentry();
	case 2:
		//This will intialize RaftNode with ID 2
		raft_config = &Config{Id: 2, LogDir: "log2", ElectionTimeout: 1200, HeartbeatTimeout: 100}
		temp_sm = &STATEMACHINE{ID: 2, VOTEDFOR: 6, STATE: "Follower", SMchanels: ch0, NEXTINDEX: [5]int{2, 2, 2, 2, 2}, COMMITINDEX: 1, CURRENTTERM: 0}
		logDir2, err := log.Open("log2") //Open Log file
		checkError(err)
		raftN = &RaftNode{config: raft_config, timerFlag: 1, sm: temp_sm, logDir: logDir2} //FInally Make Node
		//raftN.enterSamplentry();
	case 3:
		raft_config = &Config{Id: 3, LogDir: "log3", ElectionTimeout: 1400, HeartbeatTimeout: 100}
		temp_sm = &STATEMACHINE{ID: 3, VOTEDFOR: 6, STATE: "Follower", SMchanels: ch0, NEXTINDEX: [5]int{2, 2, 2, 2, 2}, COMMITINDEX: 1, CURRENTTERM: 0}
		logDir3, err := log.Open("log3") //Open Log file
		checkError(err)
		raftN = &RaftNode{config: raft_config, timerFlag: 1, sm: temp_sm, logDir: logDir3} //FInally Make Node
		//raftN.enterSamplentry();
	case 4:
		raft_config = &Config{Id: 4, LogDir: "log4", ElectionTimeout: 1600, HeartbeatTimeout: 100}
		temp_sm = &STATEMACHINE{ID: 4, VOTEDFOR: 6, STATE: "Follower", SMchanels: ch0, NEXTINDEX: [5]int{2, 2, 2, 2, 2}, COMMITINDEX: 1, CURRENTTERM: 0}
		logDir4, _ := log.Open("log4")                                                     //Open Log file
		raftN = &RaftNode{config: raft_config, timerFlag: 1, sm: temp_sm, logDir: logDir4} //FInally Make Node
		// raftN.enterSamplentry();
	}
	// Enter a sample data in log and write on disk
	//   temp_sm = &STATEMACHINE{ID: 0, VOTEDFOR: 6, STATE: "Follower",SMchanels:ch0,NEXTINDEX:[5]int{2,2,2,2,2},COMMITINDEX:1}                  ///Initialize stateMAchine
	raftN.CommitChannel = make(chan interface{}, 40)
	raftN.sm.LOG = make([]LOGENTRY, 2)
	raftN.sm.LOG[0] = LOGENTRY{0, 0, []byte("foo")}
	raftN.sm.LOG[1] = LOGENTRY{1, 0, []byte("bar")}

	raftN.logDir.RegisterSampleEntry(LOGENTRY{})
	raftN.logDir.Append(raftN.sm.LOG[0])
	raftN.logDir.Append(raftN.sm.LOG[1])
	go raftN.check_Inbox(serv)    // this routine is checking for other peer Message at Inbox
	go raftN.check_SMAction(serv) // this routine is getting actionn from State Machine through Action channel
	go temp_sm.processEvent()     //start stateMachine
	//now:= time.Now()
	//raftN.Exp_Time = now.Add(time.Duration(raft_config.ElectionTimeout) * time.Millisecond)
	raftN.timer = time.AfterFunc(time.Duration(raft_config.ElectionTimeout)*time.Millisecond, func() {
		raftN.sm.SMchanels.timeoutCh <- TIMEOUT{}
	})
	raftN.timerFlag = 1
	return raftN
}

// Following function will iterate over all RaftNode and it will return Array of RaftNodes who are Leader.
// In general there will be only one Leader but in case of partitioning there can be more than one Leader
func getLeader(rns []*RaftNode) []*RaftNode {
	leaderset := make([]*RaftNode, 1)
	for {
		for i := 0; i < len(rns); i++ {
			if strings.Compare(rns[i].sm.STATE, "LEADER") == 1 {
				if leaderset[0] == nil {
					leaderset[0] = rns[i]
				} else {
					leaderset = append(leaderset, rns[i])
				}
			}
		}
		if leaderset[0] != nil {
			return leaderset
		}
	}
	return nil
}

// Following function handles Incoming at Inbox from other peers.It reads those Event and forwards to StateMachine
func (rn *RaftNode) check_Inbox(node cluster.Server) {
	for {
		env := <-node.Inbox()
		switch env.Msg.(type) {
		// this packet is APPENDENTRIESRESP:
		case APPENDENTRIESRESP:
			temp := env.Msg.(APPENDENTRIESRESP)
			rn.sm.SMchanels.netCh <- temp

		//this packet is APPENDENTRIESREQ:
		case APPENDENTRIESREQ:
			temp := env.Msg.(APPENDENTRIESREQ)
			rn.sm.SMchanels.netCh <- temp

		// this packet is VOTERESP
		case VOTERESP:
			//	fmt.Println("VOTERESp");
			temp := env.Msg.(VOTERESP)
			rn.sm.SMchanels.netCh <- temp

		case VOTEREQ:
			temp := env.Msg.(VOTEREQ)
			rn.sm.SMchanels.netCh <- temp
		}
	}
} //End of check_Inbox function

//This goroutine receives Action coming from StateMAchine and take Actions Accordingly.

func (rn *RaftNode) check_SMAction(node cluster.Server) {
	for {
		//	fmt.Println("--->>",rn.sm.ID,"---",rn.sm.SMchanels.Action.(type))
		actionEvent := <-rn.sm.SMchanels.Action
		switch actionEvent.(type) {
		// If Action is SEND Event.It will forward that Message to respective peers
		case SEND:
			tempEvent := actionEvent.(SEND)
			//fmt.Println("-> ",tmp)
			id := tempEvent.ID_TOSEND
			mid := getMsgId(tempEvent.EVENT)
			msgid := int64(mid)
			if id == 300 { // 100 means it  is a broadcast request
				node.Outbox() <- &cluster.Envelope{Pid: cluster.BROADCAST, MsgId: msgid, Msg: tempEvent.EVENT}
			} else {
				node.Outbox() <- &cluster.Envelope{Pid: id, MsgId: msgid, Msg: tempEvent.EVENT}
			}
		case COMMIT:
		// This will update StateMachine Timeout period.
		case ALARM:
			tempEvent := actionEvent.(ALARM)
			t := tempEvent.TIME
			if rn.timerFlag == 1 {
				rn.timer.Stop()
				rn.timer = time.AfterFunc(time.Duration(t)*time.Millisecond, func() {
					rn.sm.SMchanels.timeoutCh <- TIMEOUT{}
				})
			} else {
				rn.timer = time.AfterFunc(time.Duration(t)*time.Millisecond, func() {
					rn.sm.SMchanels.timeoutCh <- TIMEOUT{}
				})
			}
		case LOGSTORE:
			tempEvent := actionEvent.(LOGSTORE)
			index := tempEvent.INDEX
			data := tempEvent.DATA.COMMAND
			CI := COMMITINFO{Index: int64(index), Data: data}
			rn.CommitChannel <- CI
			rn.writeLog(tempEvent)
		}
	}
} //End of check_SMAction() function

// This function is writing Log at Disk for RaftNode
func (rn *RaftNode) writeLog(ls LOGSTORE) {

	switch rn.sm.ID {
	case 0:
		rn.logDir.Append(ls.DATA)

	case 1:
		rn.logDir.Append(ls.DATA)
	case 2:
		rn.logDir.Append(ls.DATA)
	case 3:
		rn.logDir.Append(ls.DATA)
	case 4:
		rn.logDir.Append(ls.DATA)

	}
}

func getMsgId(temp interface{}) int {
	switch temp.(type) {
	case APPENDENTRIESRESP:
		return 1
	case APPENDENTRIESREQ:
		return 2
	case VOTERESP:
		return 3
	case VOTEREQ:
		return 4

	}
	return 81
}
