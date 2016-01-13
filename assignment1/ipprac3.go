package main
import (
    "fmt"
    "io/ioutil"
    "net"
    "os"
)
func main(){
   clientMain();
}
func clientMain(){
if len(os.Args) != 2 {
        fmt.Fprintf(os.Stderr, "Usage: %s host:port ", os.Args[0])
        os.Exit(1)
    }
    service := os.Args[1]
    tcpAddr, err := net.ResolveTCPAddr("tcp", service)
    checkError("1",err)
    conn, err := net.DialTCP("tcp", nil, tcpAddr)
    checkError("2",err)
    _, err = conn.Write([]byte("HEAD / HTTP/1.0\r\n\r\n"))
    checkError("3",err)
    result, err := ioutil.ReadAll(conn)
    fmt.Println(string(result))
    os.Exit(0)
}
func checkError(msg string,err error) {
        fmt.Println(msg);
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
