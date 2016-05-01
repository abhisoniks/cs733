# cs733 Advanced Distributed System Engineering a Cloud Project
### Aim :- Distributed File Server
Aim of project is to build a Distributed File Server. A File server based on raft consenus protocol. This distributed file server
is consist of 5 servers. Each have there own raft node and state machine. Project was accomplished in form of assignment where each assignment
is one module of File server.
### -------------Assignment Detail-----------------
#### Assignemnt1:
In assignment I made a single file server. This file server takes command from client and execute them . Command used are write a file, read a file,
compare and swap a file and delete a file.
##### Files Include:
1. basic_test.go 
2. server.go 

#### Assignment2:
Assignment2 is about state Machine implementation. We were testing statemachine behivour through various event like AppendEntriesRequest,
AppendEntriesResponse, VoteRequest, VoteResponse and timeout event were analyzing its output action for each event.
##### Files Include
1. assign2.go
2. assign2_test.go

#### Assignemnt3:
This was a wrapper to statemachine. In this assignment we integrate statemachine with raftmachine. Here we implemnted 5 raftnode exchanging
information, replicating data, electing leader. We also handled cases of partitions, leader's crashing.
#####Files Included
1. data_stru.go
2. peers.go
3. raft.go
4. raft_test.go
5. sm.go

####Assignment4:
This is integration of Raftamchine with file server of assignment1. Here we build a distributed file server.
#####Files Included
1. data_stru.go
2. peers.go
3. raft.go
4. server.go
5. sm.go
6. rpc_test.go




