package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net"
	"net/http"
	"time"
)



var s *MyServer
func init() {

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

	l, err := net.Listen("tcp", ":9999")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = http.Serve(l,s.r)
	if err != nil{
		fmt.Println(err)
	}

}


type MyServer struct {
	r *mux.Router
}
