package main

import (
	"bufio"
	"fmt"
	"github.com/cs733-iitb/cs733/assignment1/fs"
	"net"
	"os"
	"strconv"
	"encoding/json"
)

type struct File_Server{
	raft  *RaftNode
	channels map[int](chan *fs.Msg)

}

type struct msg{
    Id int
	C_msg fs.Msg
}

var crlf = []byte{'\r', '\n'}

func check(obj interface{}) {
	if obj != nil {
		fmt.Println(obj)
		os.Exit(1)
	}
}
func (fs *File_Server) getChannel(int) (*(chan *fs.Msg)) {
	ret := make(chan *fs.Msg,1)
	ch.client_chans[uid] = ret
	return &ret 
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

func serve(conn *net.TCPConn,fs *File_Server,count) {
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
        modified_msg = msg{count,msg}   // We are adding message id with it.
        newmsg,_ := json.Marshal(modified_msg)	
        switch msg.Kind {
		case 'w':
			node.Append(newmsg)
		case 'c':
			node.Append(newmsg)
		case 'd':
			node.Append(newmsg)
		}
		
		response := fs.ProcessMsg(msg)
		if !reply(conn, response) {
			conn.Clos    FS := startFile_server();k
		}
	}
}

func serverMain(address string) {
	var count int;
	count=0;
    cl, err := mkCluster()
    checkError(err)
	Raft_Id,_ := strconv.Atoi(os.Arg[2])
	node:=initialize(cl.Servers[Raft_Id],Raft_Id)
    FS := startFile_server();
    FS.raft = node
    tcpaddr, err := net.ResolveTCPAddr("tcp", address)
	check(err)
	tcp_acceptor, err := net.ListenTCP("tcp", tcpaddr)
	check(err)
	// Node has been intialized here
    for {
		tcp_conn, err := tcp_acceptor.AcceptTCP()
		check(err)
		count++;
		newchannel := make(chan *fs.Msg,1)
	  	FS.channels[count] = newchannel
		go serve(tcp_conn,FS)
	}
}

func main() {
	serverMain()
}
func startFile_server() (*ClientHandeler) (*ClientHandeler) {
	temp := ClientHandeler{}
	temp.channels = make(map[int](chan *fs.Msg))
	return &temp
}	
}