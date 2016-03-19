# RAFT NODE
## CS733 Assignment3
### Objective
In this assignment We build a “Raft node”, a wrapper over the state machine. It provides system services for the state machine

### Assignment Statement
We will create a raft Node object that wires up many bits and pieces. It should:<br>
1. Implements the interface shown in the next section.<br>
2. Create Event objects from Append(), incoming network messages, and timeouts. and feed them to the state machine<br>
3. Interpret Actions for Alarm, Network and Log events<br>
### ----Files Include-----
1. **assign2.go**     
   1. This is Assignment2 stateMachine Code.This file holds StateMachine logic
2. **data_stru.go**   
   1. This file contains all structure used in Assignment
3. **raft.go**         
   1.This file holds all Raft Logic 
4. **raft_test.go**     
   1. This is test file for this Assignment

### --------How to run Code------
1. **go test**
  1. It will run test code on all go files 
2. **./a.sh**
  1. This is script file It will run test file and will also show coverage<br>

### ----Previous Work-----
1. **Assignment1**
   1.  This assignment wass to build a simple file server, with a simple read/write interface (there’s
    no open/delete/rename)<br>
2. **Assignment2**
  1. In this assignment, We  build a state machine embedded in a raft node.We test a single instance of it.The job of the state machine is to take input events from raft node and and produce output actions.

  

