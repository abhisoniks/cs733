package main

import (
	"bufio"
	"net"
	"os"
	"strconv"
	// "fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

func main() {
	serverMain()
}
func serverMain() {
	port := ":8080"
	tcpAddr, err := net.ResolveTCPAddr("tcp", port)
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)
	for {
		conn, err := listener.Accept() // Accept waits for and returns the next connection to the listener.
		if err != nil {
			log.Fatal(err)
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	var previous string
	previous = ""
	var msg2 string
	for {
		// var command string
		msg, _, err := bufio.NewReader(conn).ReadLine()
		if strings.Compare(previous, "") != 0 {
			msg2 = previous + string(msg)
		} else {
			msg2 = string(msg)
		}

		previous = process_Msg(msg2, conn)
		if err != nil {
			return
		}

	}
}

//write(filename,noof_byte,newline,exptime,content,conn)
func write(filename string, sizeof_file string, newline int, exptime string, content string, conn net.Conn) { // write <filename> <numbytes> [<exptime>]\r\n
	var tempsize int = 0
	var size int
	var expiry_time string
	size, _ = strconv.Atoi(sizeof_file)
	tempsize = len(content)
	var ver string
	if (tempsize - newline) == size {
		if File_Exists(filename) == true {
			_, vers, _, expiry_time := get_metadata(filename)
			versi, _ := strconv.Atoi(vers)
			if (is_FileExpired(expiry_time) == true) && (strings.Compare(expiry_time, strconv.Itoa(0)) != 0) {
				conn.Write([]byte("ERR_INTERNAL\n"))
				return
			}
			versi = versi + 1
			ver = strconv.Itoa(versi)
		} else {
			ver = "1"
		}
		if exptime == "" {
			expiry_time = strconv.Itoa(0)
		} else if strings.Compare(exptime, strconv.Itoa(0)) == 0 {
			expiry_time = strconv.Itoa(0)
		} else {
			expiry_time = calculateExpiryTime(exptime)
		}
		content = filename + " " + ver + " " + sizeof_file + " " + expiry_time + "\n" + content
		err3 := ioutil.WriteFile(filename, []byte(content), 0644) //func WriteFile(filename string, data []byte, perm os.FileMode) error
		if err3 != nil {
			log.Fatal(err3)
		}
		conn.Write([]byte("OK " + ver + "\n"))
	} else {
		conn.Write([]byte("$$$ERR_INTERNAL" + "\n"))
	}
}

func read(filename string, conn net.Conn) { //read <filename>\r\n
	if File_Exists(filename) == false {
		_, _ = conn.Write([]byte("ERR_FILE_NOT_FOUND" + "\n"))
		return
	}

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
		return
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	i := 0
	var content string
	for scanner.Scan() {
		i = i + 1
		if i == 1 {
			_ = scanner.Text()
		} else {
			content = content + scanner.Text() + "\n"
		}
	}

	_, version, size, expirydate := get_metadata(filename)
	if (is_FileExpired(expirydate) == true) && (strings.Compare(expirydate, strconv.Itoa(0)) != 0) {
		_, _ = conn.Write([]byte("ERR_INTERNAL" + "\n"))
		err = os.Remove(filename) // Delete the file
		if err != nil {
			log.Fatal(err)
		}
		return
	} else {
		_, _ = conn.Write([]byte("CONTENT" + " " + version + " " + size + "  " + expirydate + "\n"))
		_, _ = conn.Write([]byte(content))
	}

}

func delete(filename string, conn net.Conn) { //delete a file  <filename>\r\n
	if File_Exists(filename) == false {
		_, _ = conn.Write([]byte("ERR_FILE_NOT_FOUND" + "\n"))
		return
	}
	_, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	err = os.Remove(filename)
	if err != nil {
		log.Fatal(err)
	}
	// err = db.Delete([]byte("filename"), nil)
	_, _ = conn.Write([]byte("OK\n "))
}

func casValid(filename string, version string, conn net.Conn) bool { //cas <filename> <version> <numbytes> [<exptime>]\r\n
	if File_Exists(filename) == false {
		_, _ = conn.Write([]byte("ERR_FILE_NOT_FOUND" + "\n"))
		return false
	}
	_, ver, _, _ := get_metadata(filename)
	if strings.Compare(ver, version) == 0 {
		return true
	} else {
		_, _ = conn.Write([]byte("ERR_VERSION"))
		return false
	}
	return true
}

func bytetostring(c []byte) string { // to convert byte array to string
	size := len(c)
	var result string
	for i := 0; i < size; i++ {
		result = result + string(c[i])
	}
	return result
}

func checkError(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func get_metadata(filename string) (string, string, string, string) {
	var first_line string
	var meta_data []string
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
		return "", "", "", ""
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	i := 0
	for scanner.Scan() {
		i = i + 1
		if i == 1 {
			first_line = scanner.Text()
			meta_data = strings.Split(first_line, " ")
		}
	}
	return meta_data[0], meta_data[1], meta_data[2], meta_data[3]
}

func File_Exists(filename string) bool {
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func is_FileExpired(expirydate string) bool {
	t := time.Now()
	current_time, _ := strconv.ParseInt(t.Format("20060102150405"), 10, 64)
	expiry_time, _ := strconv.ParseInt(expirydate, 10, 64)
	if expiry_time < current_time {
		return true
	} else {
		return false
	}
}
func calculateExpiryTime(exptime string) string {
	t := time.Now()
	mentioned_time, _ := strconv.ParseInt(exptime, 10, 64) //string to int64
	temptime := t.Add(time.Duration(mentioned_time) * time.Second)
	tt := temptime.Format("20060102150405")
	tt2, _ := strconv.ParseInt(tt, 10, 64)
	expiry_time := strconv.FormatInt(tt2, 10)
	return expiry_time
}

func validate_command(token []string) bool {
	var filename string
	var err error
	if len(token) > 1 {
		filename = token[1]
		if len(filename) > 250 {
			return false
		}
	}
	if (strings.Compare(token[0], "read") == 0) && (len(token) != 2) {
		return false
	}
	if (strings.Compare(token[0], "delete") == 0) && (len(token) != 2) {
		return false
	}
	if strings.Compare(token[0], "write") == 0 {
		if len(token) == 4 { // exptime is mentioned
			noof_byte := token[2]
			_, err = strconv.Atoi(noof_byte)
			if err != nil {
				return false
			}
		} else if len(token) == 3 { // No exptime is mentoned
			noof_byte := token[2]
			_, err = strconv.Atoi(noof_byte)
			if err != nil {
				return false
			}
		} else {
			return false
		}
	} // write ok

	if strings.Compare(token[0], "cas") == 0 {
		var noof_byte string
		if len(token) == 5 { // exptime is mentioned
			noof_byte := token[3]
			version := token[2]
			_, err = strconv.Atoi(version)
			if err != nil {
				return false
			}
			if len(version) > 8 {
				return false
			}
			_, err = strconv.Atoi(noof_byte)
			if err != nil {
				return false
			}
		} else if len(token) == 4 { // No exptime is mentoned
			noof_byte = token[3]
			version := token[2]
			_, err = strconv.Atoi(version)
			if err != nil {
				return false
			}
			if len(version) > 8 {
				return false
			}
			_, err = strconv.Atoi(noof_byte)
			if err != nil {
				return false
			}
		} else {
			return false
		}
	} //cas ok
	return true
}

func process_Msg(msg string, conn net.Conn) string {
	var pre string
	var flag int = 0
	var flagC int = 0
	var newline int = 0
	var token []string
	var further_token []string
	var content string
	var filename string
	var noof_byte string
	var exptime string
	token = strings.Split(msg, "\\r\\n")
	for i := 0; i < len(token); i++ {
		further_token = strings.Split(token[i], " ")
		if strings.Compare(further_token[0], "delete") == 0 {
			flag = 0
			if validate_command(further_token) == true {
				delete(further_token[1], conn)
			} else {
				_, _ = conn.Write([]byte("ERR_INTERNAL\n"))
			}
		} else if strings.Compare(further_token[0], "read") == 0 {
			flag = 0
			if validate_command(further_token) == true {
				read(further_token[1], conn)
			} else {
				_, _ = conn.Write([]byte("ERR_INTERNAL\n"))
			}
		} else if strings.Compare(further_token[0], "write") == 0 {
			flag = 1
			pre = token[i] + "\\r\\n"
			var further_token_content []string
			filename = further_token[1]
			noof_byte = further_token[2]
			if len(further_token) > 3 {
				exptime = further_token[3]
			} else {
				exptime = ""
			}
			var j int
			for j = i + 1; j < len(token); j++ {
				further_token_content = strings.Split(token[j], " ")
				if (strings.Compare(further_token_content[0], "read") != 0) && (strings.Compare(further_token_content[0], "write") != 0) && (strings.Compare(further_token_content[0], "delete") != 0) && (strings.Compare(further_token_content[0], "write") != 0) && (strings.Compare(further_token_content[0], "cas") != 0) {
					content = content + token[j] + "\n"
					pre = pre + token[j] + "\\r\\n"
					newline = newline + 1
					size, _ := strconv.Atoi(noof_byte)
					tempsize := len(content)
					if (tempsize - newline) == size {
						write(filename, noof_byte, newline, exptime, content, conn)
						pre = ""
					}
				} else {
					flag = 0
					i = j - 1
					break
				}
			} //for loop break
			if flag == 0 {
				write(filename, noof_byte, newline, exptime, content, conn)
				pre = ""
			} else {
				return pre
			}
		} else if strings.Compare(further_token[0], "cas") == 0 {
			flagC = 1
			pre = token[i] + "\\r\\n"
			var further_token_content []string
			filename = further_token[1]
			noof_byte = further_token[3]
			ver := further_token[2]
			if len(further_token) > 4 {
				exptime = further_token[4]
			} else {
				exptime = ""
			}
			if casValid(filename, ver, conn) == false {
				continue
			}
			var j int
			for j = i + 1; j < len(token); j++ {
				further_token_content = strings.Split(token[j], " ")
				if (strings.Compare(further_token_content[0], "read") != 0) && (strings.Compare(further_token_content[0], "write") != 0) && (strings.Compare(further_token_content[0], "delete") != 0) && (strings.Compare(further_token_content[0], "write") != 0) && (strings.Compare(further_token_content[0], "cas") != 0) {
					content = content + token[j] + "\n"
					pre = pre + token[j] + "\\r\\n"
					newline = newline + 1
					size, _ := strconv.Atoi(noof_byte)
					tempsize := len(content)
					if (tempsize - newline) == size {
						write(filename, noof_byte, newline, exptime, content, conn)
						pre = ""
					}
				} else {
					flagC = 0
					i = j - 1
					break
				}
			} //for loop break
			if flagC == 0 {
				write(filename, noof_byte, newline, exptime, content, conn)
				pre = ""
			} else {
				return pre
			}
		} else {
			conn.Write([]byte("ERR_INTERNAL\n"))
		}
	}
	return ""
}
