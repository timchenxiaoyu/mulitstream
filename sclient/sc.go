package main

import (
	"flag"
	"fmt"
	"github.com/timchenxiaoyu/mulitstream"
	"net"
	"time"
)

var path *string

func init() {

	path = flag.String("p", "/tmp/linux", "")
}

func main() {
	flag.Parse()
	for{
		conn, err := net.Dial("unix", *path)
		if err != nil {
			fmt.Println(err)
			return
		}

		client := mulitstream.NewClient(conn, mulitstream.Config{Keepalive: true})
		err = client.SendGet("/hello")
		if err != nil{
			fmt.Println(err)
		}

		time.Sleep(time.Second*3)
		err = client.SendGet("/hello")

		if err != nil{
			fmt.Println(err)
		}

		client.Stop()
		time.Sleep(time.Second *2)

	}


	//time.Sleep(time.Second * 100)

}
