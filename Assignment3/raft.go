package main
import(
//"github.com/cs733-iitb/cluster"
"github.com/cs733-iitb/log"
"fmt"

"encoding/json"

//"github.com/abhisoniks/cs733/Assignment3/Assign2"
)

var server0,server1,server2,server3,server4 RaftNode //Node in network
var sm StateMachine;
var sm0,sm1,sm2,sm3,sm4 StateMachine 			// stateMachine of all nodes
var log0,log1,log2,log3,log4 string//log File of all Machine
var lg *log.Log;

type Node interface {
/*	Append([]byte) // Client's message to Raft node
	CommitChannel() <- chan CommitInfo // A channel for client to listen on. What goes into Append must come out of here at some point.
	CommittedIndex() int  // Last known committed index in the log. This could be -1 until the system stabilizes.
	Get(index int) (err, []byte) // Returns the data at a log index, or an error.
	Id() // Node's id
	leaderId() int // Id of leader. -1 if unknown
	Shutdown() // Signal to shut down all goroutines, stop sockets, flush log and close it, cancel timers.*/
}


type Config struct {
	cluster []NetConfig  // Information about all servers, including this.
	Id int               // this node's id. One of the cluster's entries should match.
	LogDir string        // Log file directory for this node
	ElectionTimeout int
	HeartbeatTimeout int
	sm StateMachine
}

type  NetConfig struct{
	Id int
	Host string
	Port int
}

type  CommitInfo struct{
	Data []byte
	Index int64 // or int .. whatever you have in your code
	Err error // Err can be errred
}

type  RaftNode struct{
	config Config
	sm StateMachine
}

func makeRafts() []RaftNode{
	node0:=NetConfig{0,"LocalHost",9000};
	node1:=NetConfig{1,"LocalHost",9001};
	node2:=NetConfig{2,"LocalHost",9002};
	node3:=NetConfig{3,"LocalHost",9003};
	node4:=NetConfig{4,"LocalHost",9004};
	node:=[]NetConfig{node0,node1,node2,node3,node4}

    // RAftNode0 
	lg,_ = log.Open("log0")
	sm0=StateMachine{state: "Leader", leaderId: 0, votedFor: 6, currentTerm: 0, log: nil, commitIndex: 0, lastApplied: 0, voteGranted: [2]int{0, 0}}
	config:=Config{node,0,log0,20,5,sm0}
	server0= server0.New(config)

    //RaftNode1
    lg,_ = log.Open("log1")
	sm1=StateMachine{state: "Follower", leaderId: 0,log:nil ,votedFor: 6,voteGranted: [2]int{0, 0},currentTerm:0,commitIndex:0,lastApplied:0 }
	config = Config{node,1,log1,20,5,sm1}
	server1= server1.New(config)
	
	//RaftNode2
	lg,_ = log.Open("log2")
	sm2=StateMachine{state: "Follower", leaderId: 0,log:nil ,votedFor: 6,voteGranted: [2]int{0, 0},currentTerm:0,commitIndex:0,lastApplied:0 }
	config = Config{node,2,log2,20,5,sm2}
	server2= server2.New(config)

	//RaftNode3
	lg,_ = log.Open("log3")
	sm3=StateMachine{state: "Follower", leaderId: 0,log:nil ,votedFor: 6,voteGranted: [2]int{0, 0},currentTerm:0,commitIndex:0,lastApplied:0}
	config = Config{node,3,log3,20,5,sm3}
	server3= server3.New(config)

	//RaftNode4
	lg,_ = log.Open("log4")
	sm4=StateMachine{state: "Follower", leaderId: 0,log:nil ,votedFor: 6,voteGranted: [2]int{0, 0},currentTerm:0,commitIndex:0,lastApplied:0 }
	config = Config{node,4,log4,20,5,sm4}
	server4= server4.New(config)

	return []RaftNode{server0,server1,server2,server3,server4}

}
func (rn RaftNode) New(config Config) RaftNode{
		rn.config=config
		rn.sm=config.sm
		temp:=&logEntry{0,0,"foo"}	 // default log added in all raft Machine  just of testing purpose will remove it after checkpoint
		b,err := json.Marshal(temp)
		if err!=nil{
			fmt.Println("error is ",err)
		}
		b2:=[]byte(b)
		lg.Append(b2)
		go rn.NodeEvent();
		return rn
}
func (rn RaftNode) NodeEvent(){
	go rn.sm.processEvent();	//start stateMachine in a goroutine


}

func (rn RaftNode) Append([] byte){

}
