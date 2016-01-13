                                  // references ->>>  https://golang.org 
package main
import  (
 "net"
 "os" 
 "fmt"
 "strings"
 "io/ioutil"
 )
func main(){
	port := ":9300"
	tcpAddr, err := net.ResolveTCPAddr("tcp", port)
	checkError(err)
	listener,err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)
	for {
		conn,err := listener.Accept() //// Accept waits for and returns the next connection to the listener.
		if err != nil {
			continue        //handle error
		}
		 handleClient(conn)
                 conn.Close()
        }
}
func handleClient(conn net.Conn) {
        var buf[512]byte           // server side buffer
         var token[]string         // tokenize the  command 
         var input_command string
	for {
		n, err := conn.Read(buf[0:])
		if err != nil {
			return
		}
                fmt.Println("comand entered is",string(buf[0:n]));     
                input_command=string(buf[0:n]);          
                token=strings.Split(input_command," ");         
                fmt.Println("first token is",token[0]);
                if(strings.Compare(token[0],"write")==0){        // if first word is write query is for write or update file
                 write(token[1]);                             
                }
                _, err2 := conn.Write(buf[0:n])
		if err2 != nil {
			return
		}
	}
}
func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
func write(filename string){
        temp := []byte("hi first assignment")
        err:=ioutil.WriteFile(filename,temp,0644)          //func WriteFile(filename string, data []byte, perm os.FileMode) error
        checkError(err)
}

