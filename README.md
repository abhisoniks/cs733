# cs733 Advanced Distributed System Engineering a Cloud Project
### Aim :- Distributed File Server
Aim of project is to build a Distributed File Server. A File server based on raft consenus protocol. This distributed file server
is consist of 5 servers. Each have there own raft node and state machine. Project was accomplished in form of assignment where each assignment
is one module of final File server.
### -------------Assignment Detail-----------------
#### Assignemnt1:
In assignment I made a single file server. This file server takes command from client and execute them . File server supports operation of read, write, delete file . Besides this, file server also supports unusual operation like compare and swap and automatic deletion of file on basis of time expiry.
##### Files Include:
1. basic_test.go 
2. server.go 

#### Assignment2:
Assignment2 is about state Machine implementation. We were testing statemachine behivour through various event like AppendEntriesRequest,
AppendEntriesResponse, VoteRequest, VoteResponse and timeout event were analyzing its output action for each event.
##### Files Include
1. assign2.go
2. assign2_test.go
3. 

#### Assignemnt3:
This was a wrapper to statemachine. In this assignment we integrate statemachine with its raftmachine. Here we implemnted 5 raftnode exchanging
information, replicating data, electing leader. We also handled cases of partitions, leader's crashing

These 5 raftNodes are behaving concurrently and whole system will work untill majority of raft machine are in working state.
#####Files Included
1. data_stru.go
2. peers.go
3. raft.go
4. raft_test.go
5. sm.go

####Assignment4:
This is integration of Raftamchine with file server of assignment1. Here we build a highly concurrent,fault tolerant distributed file server consisting of five file server which are behaving as one for client.
#####Files Included
1. data_stru.go
2. peers.go
3. raft.go
4. server.go
5. sm.go
6. rpc_test.go




