package main
import (
    "fmt"
 //   "io/ioutil"
    "net"
    "os"
    "bufio"
)
func main(){
   clientMain();
}
func clientMain(){
     
if len(os.Args) != 2 {
        fmt.Fprintf(os.Stderr, "Usage: %s host:port ", os.Args[0])
        os.Exit(1)
    }
    
    port := os.Args[1]
    tcpAddr, err := net.ResolveTCPAddr("tcp",port)
    checkError("1",err)
    conn, err := net.DialTCP("tcp", nil, tcpAddr)                   // connects to a server func Dial(network, address string) (Conn, error)
    checkError("2",err)
    for{
    fmt.Println("Enter command:")
    reader := bufio.NewReader(os.Stdin)
    text, _ := reader.ReadString('\n')
    _, err = conn.Write([]byte(text))             // con is an object of interface Conn  which has function read,write 
    checkError("3",err)
    fmt.Println("server reply is");
       var buf [512]byte                      //client side buffer
       n, err := conn.Read(buf[0:])
       checkError("4",err)
       fmt.Println(string(buf[0:n]));
    }
       os.Exit(0)
}
func checkError(msg string,err error) {
           
	if err != nil {
                fmt.Println(msg);
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
