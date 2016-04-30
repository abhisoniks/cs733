# cs733 Assignment2 
##**RAFT STATE MACHINE**


### -----Files included----

* All source files (assign2.go,assign2_test.go)
* A script file(a.sh)
* README 


###-----How to run code----
open console and type one of these two commands
* go test
* ./a.sh 
      second way will also give a view of coverage interm of percentage of code

###-----Problem Statement----

In this assignment, We will build a state machine embedded in a raft node.We will test a single instance of it.The job of the state machine is to take input events
from  raft node and  and produce output actions. This action is submitted to raft node which will perform this action. 
Idea is to start state machine in a particular state and then throw
events at it and to see whether it changes to the right state and whether it produces expected actions or not.
This also allows us to create and test situations which are very rarely to occur  in practice.
#### Input Event
1. Append(data:[]byte): This is a request from the layer above to append the data to the replicated log. The response is in the
form of an eventual Commit action (see next section).
2. Timeout : A timeout event is interpreted according to the state. If the state machine is a leader, it is interpreted as a heartbeat
timeout, and if it is a follower or candidate, it is interpreted as an election timeout.
3. AppendEntriesReq: Message from another Raft state machine. For this and the next three event types, the parameters to the
messages can be taken from the Raft paper.
4. AppendEntriesResp: Response from another Raft state machine in response to a previous AppendEntriesReq.
5. VoteReq: Message from another Raft state machine to request votes for its candidature.
6. VoteResp: Response to a Vote request

####  OutPut Action 
1. Send(peerId, event). Send this event to a remote node. The event is one of AppendEntriesReq/Resp or VoteReq/Resp.
Clearly the state machine needs to have some information about its peer node ids and its own id.
2. Commit(index, data, err): Deliver this commit event (index + data) or report an error (data + err) to the layer above. This
is the response to the Append event.
3. Alarm(t): Send a Timeout after t milliseconds.
4. LogStore(index, data []byte): This is an indication to the node to store the data at the given index. Note that data here
can refer to the client’s


### ----Tools & Concept used----
* go language -:All basic Programming feature,Channels,Testing file
* console -: Basic command

### -----Description-----
* One instance of raft state Machine is implemented in this assignment.Various Event are thrown on it by its raft Node.According to current
state It takes Actions and return those Actions to Raft.

* For this Assignement I have assumed that there are 4 more state Machines in network other than my Machine.So there are total 5 state Machines in network.
One of these 5 nodes is leader and rest are either Candidate or Follower.

*  Machine starts in Follower state with Id 0

* Other 4 state Machine has Id respectively 1,2,3,4

* Input Event can come from Three Channels.NetCh,clientCh and timeoutCh for other event from server in network ,event from client and timeout Event respectively
* When My state Machine starts at that time Machine with Leader Id 4 is Leader.

* When stateMachine in Candidate state gets Append Event from raft node.It will give error Message

* A Follower on getting Append Event will send Error Message saying Leader is at id x where x is leader's id at that time

* when an Action is for all other Nodes It specified by send to all peer p where p is serverId.Here for is iterative loop.This is assumed that it is task of Raft Machine to send this message to all other server

###----References----
* https://raft.github.io/
* https://ramcloud.stanford.edu/raft.pdf
* https://www.youtube.com/watch?v=YbZ3zDzDnrw









