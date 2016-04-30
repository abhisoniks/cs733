package main

import (
	"fmt"
	"testing"
	"time"

	//	   "reflect"
	"strconv"
	//	"github.com/cs733-iitb/log"
	"github.com/cs733-iitb/cluster"
	"github.com/cs733-iitb/cluster/mock"
	"os"
)

func TestRaft(t *testing.T) {
	rmlog()

	/////////////////////////////////////////////////////////////////////////
	//                  Mock Cluster Testing		                       //
	/////////////////////////////////////////////////////////////////////////

	cl, err := mkCluster()
	checkError(err)
	raftset := make([]*RaftNode, 5)
	for i := 0; i < 5; i++ {
		node := initialize(cl.Servers[i], i)
		raftset[i] = node
	}

	///////////   Lets append something. /////////////

	Leader := getLeader(raftset)
	Leader[0].Append("file")
	for i := 0; i < 5; i++ {
		select { // to avoid blocking on channel.
		case ci := <-raftset[i].CommitChannel:
			temp := ci.(COMMITINFO)
			if temp.Err != nil {
				t.Fatal(temp.Err)
			} else if string(temp.Data) != "file" {
				t.Fatal("Got different data")
			} else {
				expect(t, string(temp.Data), "file")
			}
		}
	}
	fmt.Println("one append test case pass")
	/////////   Lets  try 100 append at once////////////////

	for j := 0; j < 100; j++ {
		Leader = getLeader(raftset)
		Leader[0].Append("delete")
		for i := 0; i < 5; i++ {
			select { // to avoid blocking on channel.
			case ci := <-raftset[i].CommitChannel:
				temp := ci.(COMMITINFO)
				if temp.Err != nil {
					t.Fatal(temp.Err)
				} else if string(temp.Data) != "delete" {
					t.Fatal("Got different data", string(temp.Data))
				} else {
					expect(t, string(temp.Data), "delete")
				}
			}
		} //for loop end
	} //outer loop

	time.Sleep(1000 * time.Millisecond)
	for i := 0; i < 5; i++ {
		expect(t, strconv.Itoa(raftset[i].sm.COMMITINDEX), strconv.Itoa(102))
	}
	fmt.Println("100 append test case pass")

	////////////////////////////////////////////////////////////////////
	//                      Test Partitioning                         //
	////////////////////////////////////////////////////////////////////

	// Put Leader at Majority side(a side with greater number of server),After Healing all entries should be stored at other node also////////////////

	cl.Partition([]int{0, 1, 2}, []int{3, 4})
	Leader = getLeader(raftset)
	for i := 0; i < 5; i++ {
		Leader[0].Append("cas")
	}
	time.Sleep(2000 * time.Millisecond) //lets keep this partition for 2 seconds
	cl.Heal()
	for j := 0; j < 5; j++ {
		for i := 0; i < 5; i++ {
			select { // to avoid blocking on channel.
			case ci := <-raftset[i].CommitChannel:
				temp := ci.(COMMITINFO)
				if temp.Err != nil {
					t.Fatal(temp.Err)
				} else if string(temp.Data) != "cas" {
					t.Fatal("Got different data>>", string(temp.Data))
				} else {
					expect(t, string(temp.Data), "cas")
				}
			}
		} //for loop end
	}
	time.Sleep(1000 * time.Millisecond)
	for i := 0; i < 5; i++ {
		expect(t, strconv.Itoa(raftset[i].sm.COMMITINDEX), strconv.Itoa(107))
	}
	fmt.Println("TestPartition1")

	oldLeader := Leader[0].sm.ID

	///////////////////////////////////////////////////////////////////////////
	//                         Put Leader at Minority side
	////////////////////////////////////////////////////////////////////////////

	cl.Partition([]int{0, 1}, []int{2, 3, 4})
	time.Sleep(2000 * time.Millisecond)
	Leader = getLeader(raftset)

	///////we expect till this time both of these partition should have their own Leader
	var newLeader *RaftNode
	for i := 0; i < len(Leader); i++ {
		if Leader[i].sm.ID != oldLeader {
			newLeader = Leader[i]
		}
	}

	// Lets append  some entries to new leader.
	for i := 0; i < 5; i++ {
		newLeader.Append("truncate")
	}

	// These entries should come at Leader's commit channel
	for j := 0; j < 5; j++ {
		select { // to avoid blocking on channel.
		case ci := <-newLeader.CommitChannel:
			temp := ci.(COMMITINFO)
			if temp.Err != nil {
				t.Fatal(temp.Err)
			} else if string(temp.Data) != "truncate" {
				t.Fatal("Got different data>>", string(temp.Data))
			} else {
				expect(t, string(temp.Data), "truncate")
			}
		}
	}
	cl.Heal()
	time.Sleep(1000 * time.Millisecond)
	// 1 second is enough time for re election,and getting one leader out of 2 and agreeing on common log
	// After healing I am just expecting that all 5 nodes will agree on 1 leader.
	//and all node will have commitindex equal to 112 now

	for i := 0; i < 5; i++ {
		expect(t, strconv.Itoa(raftset[i].sm.COMMITINDEX), strconv.Itoa(112))
	}

	fmt.Println("Test Partition 2")

	////////////////////////////////////////////////////////////////////
	//                   SHUT DOWN THE LEADER                         //
	////////////////////////////////////////////////////////////////////
	Leader = getLeader(raftset)
	cl.Servers[Leader[0].sm.ID].Close()

	//get new Leader
	newraftset := append(raftset[:Leader[0].sm.ID], raftset[Leader[0].sm.ID+1:]...)

	time.Sleep(1000 * time.Millisecond)
	// 1 second to elect new Leader
	Leader = nil
	for Leader == nil {
		Leader = getLeader(newraftset)
	}
	Leader[0].Append("write")

	/*  for i:=0;i<5;i++{
		fmt.Println(raftset[i].logDir.GetLastIndex());
		fmt.Println(raftset[i].sm.LOG[len(raftset[i].sm.LOG)-2:]);
	}*/

}

func mkCluster() (*mock.MockCluster, error) {
	config := cluster.Config{Peers: []cluster.PeerConfig{
		{Id: 0}, {Id: 1}, {Id: 2}, {Id: 3}, {Id: 4},
	}}
	cl, err := mock.NewCluster(config)
	return cl, err
}

func rmlog() {
	os.RemoveAll("log0")
	os.RemoveAll("log1")
	os.RemoveAll("log2")
	os.RemoveAll("log3")
	os.RemoveAll("log4")
}

func expect(t *testing.T, a string, b string) {
	if a != b {
		t.Error(fmt.Sprintf("Expected %v, found %v", b, a)) // t.Error is visible when running `go test -verbose`
	}
}
