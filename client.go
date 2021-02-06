package mulitstream

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

type Client struct {
	s *Session
}

type Config struct {
	Keepalive bool
}

func NewClient(conn io.ReadWriteCloser, c Config) *Client {

	log.Debug("new client")
	s := NewSession(conn)
	client := &Client{s: s}

	if c.Keepalive {
		go client.s.keepalive(time.Second)
	}
	return client

}


func (c *Client) Stop() {
	c.s.Close()
}


func (c *Client)SendGet(url string)error{
	stream,err := c.s.OpenStream()
	if err != nil{
		return err
	}

	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error){

				return stream,nil
			},
		},
	}
    req,_ := http.NewRequest("GET",url,nil)
    req.URL.Scheme = "http"
    req.URL.Host="localhost"
	resp,err := client.Do(req)
	if err != nil{
		fmt.Println(err)
		return err
	}
	defer resp.Body.Close()
	b,err := ioutil.ReadAll(resp.Body)
	if err != nil{
		fmt.Println(err)
		return err
	}
	fmt.Println(string(b))
	return nil
}