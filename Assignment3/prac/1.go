package main
import "github.com/cs733-iitb/log"
func main(){
     lg,_ := log.Open("mylog")
     defer lg.Close()

     lg.Append([]byte("foo"))
     lg.Append([]byte("bar"))
     lg.Append([]byte("baz"))
     lg.Append([]byte("foo"))
     var a []byte
     f:=[]string{"foo2","foo3","foo4"}
     for k:=range([]string{"foo2","foo3","foo4"}){
         copy(a[:],f[k])
     }
     
     lg.Append([]byte(a))
}