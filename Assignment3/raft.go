package main
import (
	"time"
	"github.com/cs733-iitb/cluster"
	"github.com/cs733-iitb/log"
	"encoding/gob"
	"strings"
	"fmt"
//	"reflect"

	)

var log0,log1,log2,log3,log4 *log.Log     //These are log variable name for each node


func (rn *RaftNode) Append(data string) {    //This functioin handles every Append request coming to this node
	rn.sm.SMchanels.clientCh <- APPEND{[]byte(data)}      //It passes this Append request to stateMAchine through clientCh channel
}

func (rn *RaftNode) CommittedIndex() int{     //This function returns commited Index 
	return rn.sm.COMMITINDEX;
}

func checkError(e error) {
	if e != nil {
		fmt.Println(e)
	}
}

// makerafts function will make 5 rodeNodes .It will associate a stateMAchine to all nodes.
func makerafts() []*RaftNode {
	gob.Register(VOTERESP{}) 		// register a struct name by giving it a dummy object of that name.
	gob.Register(VOTEREQ{})
	gob.Register(APPENDENTRIESREQ{})
	gob.Register(APPENDENTRIESRESP{})
	gob.Register(LOGENTRY{})
	gob.Register(COMMIT{})
	gob.Register(SEND{})
	
	raft0, err := cluster.New(0, "peer.json")
	checkError(err)
	raft1, err := cluster.New(1, "peer.json")
	checkError(err)
	raft2, err := cluster.New(2, "peer.json")
	checkError(err)
	raft3, err := cluster.New(3, "peer.json")
	checkError(err)
	raft4, err := cluster.New(4, "peer.json")
	checkError(err)
	node0:=initialize(raft0,0)
	node1:=initialize(raft1,1)
	node2:=initialize(raft2,2)
	node3:=initialize(raft3,3)
	node4:=initialize(raft4,4)
	raftSet:= []*RaftNode{node0,node1,node2,node3,node4}
	return raftSet
}

//Initialize function initialize all RaftNode related structure and also Initialize RaftNode.It will initialize STATEMACHINE
//This function starts stateMachine, timer of Node ,a routine of checking Inbox and a routine for checking Action coming from stateMAchine

func initialize(serv cluster.Server, id int) *RaftNode{
	var err error;
	switch id {
		//THis will initialize RaftNode with ID 0
	case 0:
		ch0 := new(Channel)                                    //Make channels between SM and node
		ch0.Action = make(chan interface{},40)
		ch0.clientCh=make(chan interface{},40)
		ch0.timeoutCh=make(chan interface{},40)
		ch0.netCh=make(chan interface{},40)

		raft_config := &Config{0,"log0",2200,500}	     //Initialize config of node
		temp_sm0 := &STATEMACHINE{ID: 0, VOTEDFOR: 6, STATE: "Follower",SMchanels:ch0,NEXTINDEX:[5]int{2,2,2,2,2}}                  ///Initialize stateMAchine
		
		// Enter a sample data in log and write on disk
		temp_sm0.LOG = make([]LOGENTRY,2)
		temp_sm0.LOG[0]=LOGENTRY{0,0,[]byte("foo")}
		temp_sm0.LOG[1]=LOGENTRY{1,0,[]byte("bar")}

		raftN0 := &RaftNode{config:raft_config,timerFlag:1,sm:temp_sm0}    //FInally Make Node
		raftN0.CommitChannel = make(chan interface{},40)

		log0, err = log.Open("log0")	//Open Log file related to this stateMachine
		checkError(err)
		log0.RegisterSampleEntry(LOGENTRY{}) 
		
		log0.Append(temp_sm0.LOG[0])    //Put some sample data in Log of this Node
		log0.Append(temp_sm0.LOG[1])
		
		go raftN0.check_Inbox(serv)    // this routine is checking for other peer Message at Inbox
		go raftN0.check_SMAction(serv) // this routine is getting actionn from State Machine through Action channel
        go temp_sm0.processEvent()     //start stateMachine
	
		//start Timer for this Node	
		now:= time.Now()	
		raftN0.Exp_Time = now.Add(time.Duration(raft_config.ElectionTimeout) * time.Millisecond)
		go raftN0.timer2()
		return raftN0;
	case 1:
		ch1 := new(Channel)
		ch1.Action = make(chan interface{},40)
		ch1.clientCh=make(chan interface{},40)
		ch1.timeoutCh=make(chan interface{},40)
		ch1.netCh=make(chan interface{},40)
	    
		raft_config := &Config{Id: 1, LogDir: "log1", ElectionTimeout: 2700, HeartbeatTimeout: 500}
		temp_sm1 := new(STATEMACHINE)
		temp_sm1 = &STATEMACHINE{ID: 1, VOTEDFOR: 6, STATE: "Follower",SMchanels:ch1,NEXTINDEX:[5]int{2,2,2,2,2}}

		//Put some Entry in Log 
		temp_sm1.LOG = make([]LOGENTRY,2)
		temp_sm1.LOG[0]=LOGENTRY{0,0,[]byte("foo")}
		temp_sm1.LOG[1]=LOGENTRY{1,0,[]byte("bar")}
		
        //Intialize RaftNode
		raftN1 := new(RaftNode)
		raftN1 = &RaftNode{config:raft_config,timerFlag:1,sm:temp_sm1}    //FInally Make Node
		raftN1.CommitChannel = make(chan interface{},40)
		log1, err= log.Open("log1")	//Open Log file
		checkError(err)
		log1.RegisterSampleEntry(LOGENTRY{})
		
		//Append data In file
		log1.Append(temp_sm1.LOG[0])
		log1.Append(temp_sm1.LOG[1])
		go raftN1.check_Inbox(serv) // this routine is checking for other peer Message at Inbox
		go raftN1.check_SMAction(serv) // this routine is getting actionn from State Machine through Action channel
		
		//Following routine handles timer for this raftNode
		now:= time.Now()	
		raftN1.Exp_Time = now.Add(time.Duration(raft_config.ElectionTimeout) * time.Millisecond)
	    go raftN1.timer2()
		go temp_sm1.processEvent() //start stateMachine
		
		return raftN1;
	case 2:
		//This will intialize RaftNode with ID 2
		ch2 := new(Channel)
		ch2.Action = make(chan interface{},40)
		ch2.clientCh=make(chan interface{},40)
		ch2.timeoutCh=make(chan interface{},40)
		ch2.netCh=make(chan interface{},40)
		raft_config := &Config{Id: 2, LogDir: "log2", ElectionTimeout: 3200, HeartbeatTimeout: 500}
		temp_sm2 := new(STATEMACHINE)
		temp_sm2 = &STATEMACHINE{ID: 2, VOTEDFOR: 6, STATE: "Follower",SMchanels:ch2,NEXTINDEX:[5]int{2,2,2,2,2}}

		temp_sm2.LOG = make([]LOGENTRY,2)
		temp_sm2.LOG[0]=LOGENTRY{0,0,[]byte("foo")}
		temp_sm2.LOG[1]=LOGENTRY{1,0,[]byte("bar")}
		//temp_sm2.LOG=append(temp_sm2.LOG,LOGENTRY{0,0,[]byte("foo")}) 	    // Entry in Log

		raftN2 := &RaftNode{config:raft_config,timerFlag:1,sm:temp_sm2}    //FInally Make Node
		raftN2.CommitChannel = make(chan interface{},40)
		log2,err = log.Open("log2")	//Open Log file
		log2.RegisterSampleEntry(LOGENTRY{})
		checkError(err)
		log2.Append(temp_sm2.LOG[0])
		log2.Append(temp_sm2.LOG[1])
		
		go raftN2.check_Inbox(serv) // this routine is checking for other peer Message at Inbox
		go raftN2.check_SMAction(serv) // this routine is getting actionn from State Machine through Action channel
		go temp_sm2.processEvent() //start stateMachine
		now:= time.Now()	
		raftN2.Exp_Time = now.Add(time.Duration(raft_config.ElectionTimeout) * time.Millisecond)
		raftN2.CommitChannel = make(chan interface{},40)
		go raftN2.timer2()
		return raftN2;
		

	case 3:
		// This will Intiliaze RaftNode with id   3
		ch3 := new(Channel)
		ch3.Action = make(chan interface{},40)
		ch3.clientCh=make(chan interface{},40)
		ch3.timeoutCh=make(chan interface{},40)
		ch3.netCh=make(chan interface{},40)
	    //	raft_config := new(Config)
		raft_config := &Config{Id: 3, LogDir: "log3", ElectionTimeout: 3700, HeartbeatTimeout: 500}
		temp_sm3 := new(STATEMACHINE)
		temp_sm3 = &STATEMACHINE{ID: 3, VOTEDFOR: 6, STATE: "Follower",SMchanels:ch3,NEXTINDEX:[5]int{2,2,2,2,2}}

		temp_sm3.LOG = make([]LOGENTRY,2)

		temp_sm3.LOG[0]=LOGENTRY{0,0,[]byte("foo")}
		temp_sm3.LOG[1]=LOGENTRY{1,0,[]byte("bar")}

		//MAke a Node .Intialize it with stateMachine
		raftN3 := &RaftNode{config:raft_config,timerFlag:1,sm:temp_sm3}    //FInally Make Node
		raftN3.CommitChannel = make(chan interface{},40)
		log3,err = log.Open("log3")	//Open Log file
		log3.RegisterSampleEntry(LOGENTRY{})
		checkError(err)
		log3.Append(temp_sm3.LOG[0])
		log3.Append(temp_sm3.LOG[1])
		go raftN3.check_Inbox(serv) // this routine is checking for other peer Message at Inbox
		go raftN3.check_SMAction(serv) // this routine is getting actionn from State Machine through Action channel
		go temp_sm3.processEvent() //start stateMachine

		// FOllowing goroutine will Handle timer for this raftNode
		now:= time.Now()
		raftN3.Exp_Time = now.Add(time.Duration(raft_config.ElectionTimeout) * time.Millisecond)
		go raftN3.timer2()
		return raftN3;
	case 4:
		ch4 := new(Channel)
		ch4.Action = make(chan interface{},40)
		ch4.clientCh=make(chan interface{},40)
		ch4.timeoutCh=make(chan interface{},40)
		ch4.netCh=make(chan interface{},40)
		
		raft_config := &Config{Id: 4, LogDir: "log4", ElectionTimeout: 4200, HeartbeatTimeout: 500}
		temp_sm4 := new(STATEMACHINE)
		temp_sm4 = &STATEMACHINE{ID: 4, VOTEDFOR: 6, STATE: "Follower",SMchanels:ch4,NEXTINDEX:[5]int{2,2,2,2,2}}
		temp_sm4.LOG = make([]LOGENTRY,2)
		temp_sm4.LOG[0]=LOGENTRY{0,0,[]byte("foo")}
		temp_sm4.LOG[1]=LOGENTRY{1,0,[]byte("bar")}
		//temp_sm4.LOG=append(temp_sm4.LOG,LOGENTRY{0,0,[]byte("foo")}) 	    // Entry in Log
		
		raftN4 := &RaftNode{config:raft_config,timerFlag:1,sm:temp_sm4}    //FInally Make Node
		raftN4.CommitChannel = make(chan interface{},40)
		log4, _ = log.Open("log4")	//Open Log file
		log4.RegisterSampleEntry(LOGENTRY{})
		checkError(err)
		log4.Append(temp_sm4.LOG[0])
		log4.Append(temp_sm4.LOG[1])
		
		go raftN4.check_Inbox(serv) // this routine is checking for other peer Message at Inbox
		go raftN4.check_SMAction(serv) // this routine is getting actionn from State Machine through Action channel
		go temp_sm4.processEvent() //start stateMachine
		now:= time.Now()
		//raftN4.timerFlag=1	
		raftN4.Exp_Time = now.Add(time.Duration(raft_config.ElectionTimeout) * time.Millisecond)
		go raftN4.timer2()
		return raftN4;
		default : return new(RaftNode)
		
	}
	
}

// Following function will iterate over all RaftNode and it will return RaftNode who is Leader.
func getLeader(rns []*RaftNode) *RaftNode{
  for{
	for i:=0;i<5;i++{
		if strings.Compare(rns[i].sm.STATE,"LEADER")==1{
			return rns[i];
		}
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
		actionEvent := <-rn.sm.SMchanels.Action
		switch actionEvent.(type) {
			// If Action is SEND Event.It will forward that Message to respective peers
		case SEND:
			tempEvent := actionEvent.(SEND)
			//fmt.Println("-> ",tmp)
			id := tempEvent.ID_TOSEND
			mid := getMsgId(tempEvent.EVENT)
			msgid := int64(mid)
			if id == 500 { // 100 means it  is a broadcast request
				node.Outbox() <- &cluster.Envelope{Pid: cluster.BROADCAST, MsgId: msgid, Msg: tempEvent.EVENT}
			} else {
				node.Outbox() <- &cluster.Envelope{Pid: id, MsgId: msgid, Msg: tempEvent.EVENT}
			}
		case COMMIT:
		
		// This will update StateMachine Timeout period.		
		case ALARM:
			tempEvent := actionEvent.(ALARM)
			t := tempEvent.TIME
			now := time.Now()
			rn.Exp_Time = now.Add(time.Duration(t) * time.Millisecond)
			rn.timerFlag = 1
		case LOGSTORE:
			tempEvent := actionEvent.(LOGSTORE)
			index:=tempEvent.INDEX
			data:=tempEvent.DATA.COMMAND
			CI:=COMMITINFO{Index:int64(index),Data:data}
			rn.CommitChannel <-CI
		    rn.writeLog(tempEvent)
		}
	}
} //End of check_SMAction() function

// This function is holding Timeout logic of stateMachine.

func (rn *RaftNode) timer2() {
	for {	fmt.Print("")

			if rn.timerFlag == 1 {
			for {
				now2:= time.Now()
				if now2.After(rn.Exp_Time) {
					//	fmt.Println("$$$$")

					rn.sm.SMchanels.timeoutCh <- TIMEOUT{}
					rn.timerFlag = 0
					break
				}
			}
		}
	}
}

// This function is writing Log at Disk for RaftNode
func (rn *RaftNode) writeLog(ls LOGSTORE){
	switch rn.sm.ID{
	case 0:	
	fmt.Println("0")
	log0.Append(ls.DATA)
	case 1: 
	fmt.Println("1")
	log1.Append(ls.DATA)
	
	case 2: 
	fmt.Println("2")
	log2.Append(ls.DATA)
	
	case 3: 
	fmt.Println("3")
	log3.Append(ls.DATA)
	
	case 4: 
	fmt.Println("4")
	log4.Append(ls.DATA)
	
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
