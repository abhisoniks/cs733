package main
import (
    "fmt"
    "net"
    "os"
)
func main(){
     serverMain()
}
func serverMain(){
    service:=":8080"
    tcpAddr,err:= net.ResolveTCPAddr("tcp",service)
    checkError(err)
    listener,err := net.ListenTCP("tcp",tcpAddr)
    checkError(err)
    for {
        conn, err := listener.Accept()
        if err!= nil {
            continue
        }
        sss := "Hello"
	fmt.Println("Hello")
        conn.Write([]byte(sss))
fmt.Println("Hello")
        conn.Close()              
    }
}
func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

