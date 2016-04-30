package main

import (
	"bufio"
	"fmt"
	"net"
	"time"
	"strconv"
	"strings"
	"testing"
)

// Simple serial check of getting and setting
func TestTCPSimple(t *testing.T) {
   	go serverMain()
     time.Sleep(1 * time.Second)                // one second is enough time for the server to start
    name := "hi2.txt"
	contents := "bye"
	exptime := 300000
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		t.Error(err.Error()) // report error through testing framework
	}

//---------------------------Simple testcases---------

	scanner := bufio.NewScanner(conn)
	// Write a file
	fmt.Fprintf(conn, "write %v %v %v\r\n%v\r\n", name, len(contents), exptime, contents)
	scanner.Scan() // read first line
	resp := scanner.Text() // extract the text from the buffer
	arr := strings.Split(resp, " ") // split into OK and <version>
	expect(t, arr[0], "OK")
	version, err := strconv.ParseInt(arr[1], 10, 64) // parse version as number
	if err != nil {
		t.Error("Non-numeric version found")
	}
  
     //lets try another write we will delete it later
    fmt.Fprintf(conn, "write %v %v %v\r\n%v\r\n", "file2", len(contents), exptime, contents)           
	scanner.Scan() // read first line
	resp = scanner.Text() // extract the text from the buffer
	arr = strings.Split(resp, " ") // split into OK and <version>
	expect(t, arr[0], "OK")
	version, err = strconv.ParseInt(arr[1], 10, 64) // parse version as number
	if err != nil {
		t.Error("Non-numeric version found")
	}

	//lets give more data than 

      // Let's try cas 
   /* fmt.Fprintf(conn, "cas %v %v %v\r\n%v\r\n", "file2","1",len(contents), exptime, contents)           
	scanner.Scan() // read first line
	resp = scanner.Text() // extract the text from the buffer
	arr = strings.Split(resp, " ") // split into OK and <version>
	expect(t, arr[0], "OK")
	version, err = strconv.ParseInt(arr[1], 10, 64) // parse version as number
	if err != nil {
		t.Error("Non-numeric version found")
	}*/
     
      // lets give more data than mentioned
    fmt.Fprintf(conn, "write %v %v %v\r\n%v\r\n", "file2", "3", exptime, "jadoo")           
	scanner.Scan() // read first line
	resp = scanner.Text() // extract the text from the buffer
	arr = strings.Split(resp, " ") // split into OK and <version>
	expect(t, arr[0], "OK")
	version, err = strconv.ParseInt(arr[1], 10, 64) // parse version as number
	if err != nil {
		t.Error("Non-numeric version found")
	}


    
   fmt.Fprintf(conn, "read %v\r\n", name) // try a read now
	scanner.Scan()

	arr = strings.Split(scanner.Text(), " ")
	expect(t, arr[0], "CONTENTS")
	expect(t, arr[1], fmt.Sprintf("%v", version)) // expect only accepts strings, convert int version to string
	expect(t, arr[2], fmt.Sprintf("%v", len(contents)))	
	scanner.Scan()
	expect(t, contents, scanner.Text())

	fmt.Fprintf(conn, "read %v\r\n", "random")       // try a read with bad file name now
	scanner.Scan()
	arr = strings.Split(scanner.Text(), " ")
	expect(t, arr[0], "ERR_FILE_NOT_FOUND")  


    fmt.Fprintf(conn, "read\r\n")                    // try a bad read format No file name given
	scanner.Scan()
    arr = strings.Split(scanner.Text(), " ")
	expect(t, arr[0], "ERR_INTERNAL")


	fmt.Fprintf(conn, "delete %v\r\n", "file2")              // try a delete now
	scanner.Scan()
	arr = strings.Split(scanner.Text(), " ")
	expect(t, arr[0], "OK")
	
	fmt.Fprintf(conn, "delete %v\r\n", "random")            // try a delete with bad file name now
	scanner.Scan()
    arr = strings.Split(scanner.Text(), " ")
	expect(t, arr[1], "ERR_FILE_NOT_FOUND")

	fmt.Fprintf(conn, "delete\r\n")                               // try a bad delete format No file name given
	scanner.Scan()
    arr = strings.Split(scanner.Text(), " ")
	expect(t, arr[0], "ERR_INTERNAL")

    /// ------------- ----------Concurrency TESTCASES------------------------- //////////////////////////
   
    
    for i:=0;i<1;i++{
    	go checkconcurrency(t)
    	time.Sleep(30*time.Second);
    }
}


func checkconcurrency(t *testing.T){
	name1 := "myfile.txt"
	contents := "bye"
	exptime := 300000
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		t.Error(err.Error()) // report error through testing framework
	}

	scanner := bufio.NewScanner(conn)
	
	// Write a file
	fmt.Fprintf(conn, "write %v %v %v\r\n%v\r\n", name1, len(contents), exptime, contents)
	scanner.Scan() // read first line
	resp := scanner.Text() // extract the text from the buffer
	arr := strings.Split(resp, " ") // split into OK and <version>
	expect(t, arr[0], "OK")
	version, err := strconv.ParseInt(arr[1], 10, 64) // parse version as number
	if err != nil {
		t.Error("Non-numeric version found")
	}
	
	fmt.Fprintf(conn, "read %v\r\n", name1) // try a read now
	scanner.Scan()
    arr = strings.Split(scanner.Text(), " ")
	expect(t, arr[0], "CONTENTS")
	expect(t, arr[1], fmt.Sprintf("%v", version)) // expect only accepts strings, convert int version to string
	expect(t, arr[2], fmt.Sprintf("%v", len(contents)))	
	scanner.Scan()
	expect(t, contents, scanner.Text())
	

	fmt.Fprintf(conn, "read %v\r\n", "random") // try a read with bad file name now
	scanner.Scan()
	arr = strings.Split(scanner.Text(), " ")
	expect(t, arr[0], "ERR_FILE_NOT_FOUND")  


    fmt.Fprintf(conn, "read\r\n")        // try a bad read format No file name given
	scanner.Scan()
    arr = strings.Split(scanner.Text(), " ")
	expect(t, arr[0], "ERR_INTERNAL")


	fmt.Fprintf(conn, "delete %v\r\n", name1) // try a delete now
	scanner.Scan()
	arr = strings.Split(scanner.Text(), " ")
	expect(t, arr[0], "OK")
	
	fmt.Fprintf(conn, "delete %v\r\n", "random") // try a delete with bad file name now
	scanner.Scan()
    arr = strings.Split(scanner.Text(), " ")
	expect(t, arr[1], "ERR_FILE_NOT_FOUND")

	fmt.Fprintf(conn, "delete\r\n")        // try a bad delete format No file name given
	scanner.Scan()
    arr = strings.Split(scanner.Text(), " ")
	expect(t, arr[0], "ERR_INTERNAL")


}

// Useful testing function
func expect(t *testing.T, a string, b string) {
	if a != b {
		t.Error(fmt.Sprintf("Expected %v, found %v", b, a)) // t.Error is visible when running `go test -verbose`
	}
}