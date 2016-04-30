package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/cs733-iitb/assignment1/fs"
	"github.com/cs733-iitb/cluster"
	"net"
	"os"
	"strconv"
	"sync"
)

type File_Server struct {
	Msgcount int
	Raft     *RaftNode
	Channels map[int](chan *fs.Msg)
}

type Message struct {
	Uid   int
	Id    int
	C_msg fs.Msg
}

var mutex = &sync.Mutex{}
var crlf = []byte{'\r', '\n'}

func check(obj interface{}) {
	if obj != nil {
		fmt.Println(obj)
		os.Exit(1)
	}
}

func reply(conn *net.TCPConn, msg *fs.Msg) bool {
	var err error
	write := func(data []byte) {
		if err != nil {
			return
		}
		_, err = conn.Write(data)
	}
	var resp string
	switch msg.Kind {
	case 'C': // read response
		resp = fmt.Sprintf("CONTENTS %d %d %d", msg.Version, msg.Numbytes, msg.Exptime)
	case 'O':
		resp = "OK "
		if msg.Version > 0 {
			resp += strconv.Itoa(msg.Version)
		}
	case 'F':
		resp = "ERR_FILE_NOT_FOUND"
	case 'V':
		resp = "ERR_VERSION " + strconv.Itoa(msg.Version)
	case 'M':
		resp = "ERR_CMD_ERR"
	case 'I':
		resp = "ERR_INTERNAL"
	case 'L':
		resp = "ERROR_REDIRECTING " + strconv.Itoa(msg.Numbytes)
	default:
		fmt.Printf("Unknown response kind '%c'", msg.Kind)
		return false
	}
	resp += "\r\n"
	write([]byte(resp))
	if msg.Kind == 'C' {
		write(msg.Contents)
		write(crlf)
	}
	return err == nil
}

func serve(conn *net.TCPConn, fileS *File_Server, uid int) {
	reader := bufio.NewReader(conn)
	for {
		msg, msgerr, fatalerr := fs.GetMsg(reader)
		if fatalerr != nil || msgerr != nil {
			reply(conn, &fs.Msg{Kind: 'M'})
			conn.Close()
			break
		}
		if msgerr != nil {
			if (!reply(conn, &fs.Msg{Kind: 'M'})) {
				conn.Close()
				break
			}
		}
		newchannel := make(chan *fs.Msg, 1)
		mutex.Lock()
		fileS.Msgcount = fileS.Msgcount + 1
		count := fileS.Msgcount
		fileS.Channels[count] = newchannel
		fmt.Println("Listening at", count)
		mutex.Unlock()
		modified_msg := Message{uid, count, *msg} // We are adding message id with it.
		newmsg, _ := json.Marshal(modified_msg)
		fileS.Raft.Append(newmsg)

		for {
			select {
			case ci := <-fileS.Channels[count]:
				reply(conn, ci)
			}
			break
		}
	}
}

func serverMain(address string, id string) {
	var uid int
	uid = 0
	cl, err := mkCluster(id)
	checkError(err)
	Raft_Id, _ := strconv.Atoi(id)
	node := initialize(cl, Raft_Id)
	FS := startFile_server()
	FS.Raft = node
	tcpaddr, err := net.ResolveTCPAddr("tcp", "localhost:"+address)
	check(err)
	tcp_acceptor, err := net.ListenTCP("tcp", tcpaddr)
	check(err)

	//  start looking over commitchannel
	go FS.Commit()

	for {
		tcp_conn, err := tcp_acceptor.AcceptTCP()
		check(err)
		go serve(tcp_conn, FS, uid)
		uid++
		fmt.Println("uid is", uid)
	}
}

func (FS *File_Server) Commit() {
	for {
		select {
		case ci := <-FS.Raft.CommitChannel:
			temp := ci.(COMMITINFO)
			var msg_received Message
			var error_code int
			var LeaderId int
			error_code = temp.Err
			var result *fs.Msg
			json.Unmarshal(temp.Data, &msg_received)
			if error_code == 200 {
				result = fs.ProcessMsg(&msg_received.C_msg)
			} else {
				LeaderId = int(temp.Index)
				result = &fs.Msg{Kind: 'L', Numbytes: LeaderId}
			}
			mutex.Lock()
			fmt.Println("sender Id", msg_received.Id)
			FS.Channels[msg_received.Id] <- result
			mutex.Unlock()
		}
	}
}

func main() {
	serverMain(os.Args[1], os.Args[2])
}

// Its intializing file server
func startFile_server() *File_Server {
	temp := File_Server{}
	temp.Channels = make(map[int](chan *fs.Msg))
	return &temp
}

func mkCluster(id string) (cluster.Server, error) {
	int_id, _ := strconv.Atoi(id)
	cl, err := cluster.New(int_id, "peer.json")
	return cl, err
}
