package main

import (
	"flag"
	"fmt"
	"github.com/timchenxiaoyu/mulitstream"
	"net/http"
	"os"
	"runtime"
	"time"
	"github.com/gorilla/mux"
)

var path *string
var s *MyServer
func init() {
	path = flag.String("p", "/dev/virtio-ports/org.qemu.guest_agent.0", "")

	s = &MyServer{
		r: mux.NewRouter(),
	}
	s.r.HandleFunc("/hello", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Println("get request", time.Now())
		_, err := writer.Write([]byte("hello weixian,success!!!"))
		if err != nil {
			fmt.Printf("write err %v", err)
		}
	})
}




func main() {
	flag.Parse()

	if runtime.GOOS == "windows" {
		*path = "\\\\.\\Global\\org.qemu.guest_agent.0"
	}

	f, err := os.OpenFile(*path, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("open chan", err)
		return
	}

	l := mulitstream.NewSession(f)


	err = http.Serve(l, s.r)
	if err != nil {
		fmt.Println("serve", err)
	}

	time.Sleep(time.Second * 1000)
	l.Close()
	//}

}


type MyServer struct {
	r *mux.Router
}
