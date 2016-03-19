package main

import (
	"fmt"
	"testing"
    "time"
	//   "reflect"
	"strconv"
)

func TestTCPSimple(t *testing.T) {
		// makerafts function return 5 raftNode objects
		raft:=makerafts();
		time.Sleep(15*time.Second)
		Leader:=getLeader(raft)
		// Lets try to append something.Once appended it should come at commitChannel at some time 
		Leader.Append("file")
		for i:=0;i<5;i++ {
			select { // to avoid blocking on channel.
				case ci := <- raft[i].CommitChannel:
					temp:=ci.(COMMITINFO)
					if temp.Err != nil { t.Fatal(temp.Err) 
					}else if string(temp.Data) != "file" {
						t.Fatal("Got different data")
					}else{
						expect(t,string(temp.Data), "file")
					} 
			}
		}
		time.Sleep(10*time.Second)
	//	Lets append delete 
		Leader.Append("delete")
		for i:=0;i<5;i++ {
			select { // to avoid blocking on channel.
				case ci := <- raft[i].CommitChannel:
					temp:=ci.(COMMITINFO)
					if temp.Err != nil {t.Fatal(temp.Err)
					}else if string(temp.Data) != "delete" {
						t.Fatal("Got different data")
					}else{
						expect(t,string(temp.Data), "delete")
					} 

					
			}
		}	
		k:=Leader.CommittedIndex();
		expect(t,"2",strconv.Itoa(k));



		
}

func expect(t *testing.T, a string, b string) {
	if a != b {
		t.Error(fmt.Sprintf("Expected %v, found %v", b, a)) // t.Error is visible when running `go test -verbose`
	}
}
