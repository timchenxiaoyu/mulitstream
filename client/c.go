package main

import (
	"fmt"
	"github.com/timchenxiaoyu/mulitstream"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:9999")
	if err != nil {
		fmt.Println(err)
		return
	}

	client := mulitstream.NewClient(conn, mulitstream.Config{Keepalive: true})
	client.SendGet("/hello")
	//time.Sleep(time.Second * 100)

}
