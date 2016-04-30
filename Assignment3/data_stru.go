package main

import (
	"github.com/cs733-iitb/cluster"
	"github.com/cs733-iitb/log"
	"time"
)

type STATEMACHINE struct {
	ID            int //own id
	LEADERID      int
	STATE         string
	VOTEDFOR      int // 0 if not voted for anyone else Id of candidate to whom voted
	CURRENTTERM   int
	CURRENTCOMMIT int // it tell INDEX upto which log is COMMITed
	LOG           []LOGENTRY
	COMMITINDEX   int
	LASTAPPLIED   int //INDEX of highest log entry applied to state machine
	NEXTINDEX     [5]int
	MATCHINDEX    [5]int
	VOTEGRANTED   [2]int   //First INDEX counts number of positive VOTERESPonse.second INDEX counts number of negative VOTERESPonse
	SMchanels     *Channel //Channels between stateMachins and Node
}

// Structure of Channels .
type Channel struct {
	Action    chan interface{}
	clientCh  chan interface{}
	timeoutCh chan interface{}
	netCh     chan interface{}
}

type LOGENTRY struct {
	INDEX   int
	TERM    int
	COMMAND []byte
}

type APPEND struct {
	A []byte
}

type APPENDENTRIESREQ struct { //APPENDENTRIESREQ(TERM int,LEADERID int, PREVLOGINDEX int,PREVLOGTERM int,ENTRIES[] string,LEADERCOMMIT int)
	TERM         int
	LEADERID     int
	PREVLOGINDEX int
	PREVLOGTERM  int
	ENTRIES      []LOGENTRY
	LEADERCOMMIT int
}
type APPENDENTRIESRESP struct {
	TERM     int
	SENDERID int
	AERESP   bool
}

type TIMEOUT struct {
}

type VOTEREQ struct {
	TERM         int //candidate’s TERM
	CANDIDATEID  int //	int candidate requesting vote
	LASTLOGINDEX int //INDEX of candidate’s last log entry (§5.4)
	LASTLOGTERM  int //of candidate’s last log entry (§5.4)

}
type VOTERESP struct {
	TERM     int  // for which TERM this voteRes is
	VOTERESP bool //tells VOTERESPonse yes or No
}

type COMMIT struct {
	DATA  string
	ERROR int //this is ERROR code if  I am not a leader Its value is 404
	INDEX int
}

type SEND struct {
	ID_TOSEND int // id to whom this EVENT is to SEND
	EVENT     interface{}
}

type ALARM struct {
	TIME int
}

type LOGSTORE struct {
	INDEX int
	DATA  LOGENTRY
}

type Config struct {
	Id               int    // this node's id. One of the cluster's entries should match.
	LogDir           string // Log file directory for this node
	ElectionTimeout  int
	HeartbeatTimeout int
}

type COMMITINFO struct {
	Data  []byte
	Index int64 // or int .. whatever you have in your code
	Err   error // Err can be errred
}

type RaftNode struct {
	config    *Config // for raft node it holds configuration of node
	timerFlag int
	sm        *STATEMACHINE //This is stateMachine of this node
	//Exp_Time time.Time        //This is Expiry time.It maintains time at which raft node will send Timeout event to stateMachine
	CommitChannel chan interface{}
	logDir        *log.Log
	timer         *time.Timer
	//now2 time.Time

}

type Node interface {
	Append(string)                    // Client's message to Raft node
	CommitChannel() <-chan COMMITINFO // A channel for client to listen on. What goes into Append must come out of here at some point.
	CommittedIndex() int              // Last known committed index in the log. This could be -1 until the system stabilizes.
	Get(index int) (error, []byte)    // Returns the data at a log index, or an error.
	Id()                              // Node's id
	leaderId() int                    // Id of leader. -1 if unknown
	Shutdown()                        // Signal to shut down all goroutines, stop sockets, flush log and close it, cancel timers.
	check_Inbox(cluster.Server)
	check_SMAction(cluster.Server)
}
