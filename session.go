package mulitstream

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var log = logrus.New()

func init() {
	// 以JSON格式为输出，代替默认的ASCII格式
	log.SetFormatter(&logrus.JSONFormatter{})
	// 以Stdout为输出，代替默认的stderr
	log.SetOutput(os.Stdout)
	// 设置日志等级
	log.SetLevel(logrus.DebugLevel)
}

var (
	packageHandlers = []func(*Session, header) error{
		DataPackage:   (*Session).handleStreamMessage,
		//WindowPackage: (*Session).handleStreamMessage,
		PingPackage: (*Session).handlePing,
		//FinalPackage:  (*Session).handleGoAway,
	}

	WaitTimeout = 2 * time.Second
)

type Session struct {
	conn   io.ReadWriteCloser
	sendCh chan Package

	//stream
	streams    map[uint32]*Stream
	streamLock sync.Mutex
	nextStreamID uint32

	//ping
	pings    map[uint32]chan struct{}
	pingID   uint32
	pingLock sync.Mutex

	//shutdown
	shutdownCh chan struct{}
	shutdown   bool

	//accept
	acceptCh chan *Stream
}

func NewSession(conn io.ReadWriteCloser) *Session {
	s := &Session{
		conn:       conn,
		sendCh:     make(chan Package, 64),
		streams:    make(map[uint32]*Stream),
		pings:      make(map[uint32]chan struct{}),
		shutdownCh: make(chan struct{}),
		acceptCh:   make(chan *Stream, 100),
	}

	go s.receive()
	go s.send()

	return s
}

func (s *Session) Close() error {
	s.conn.Close()
	s.shutdown = true
	close(s.shutdownCh)
	return nil
}

func (s *Session) IsClose() bool {
	return s.shutdown
}

func (s *Session) OpenStream() (*Stream, error) {
	if s.IsClose(){
		return nil,fmt.Errorf("connection close")
	}

	nextID := atomic.AddUint32(&s.nextStreamID,2)
	stream := NewStream(s, nextID,streamInit)
	s.streamLock.Lock()
	s.streams[nextID] =stream
	s.streamLock.Unlock()
	return stream,nil
}

func (s *Session) send() {

	for {
		select {
		case pack := <-s.sendCh:
			//1.write header
			_, err := s.conn.Write(pack.Hdr)
			if err != nil {
				log.Error("write header", err)
				continue
			}
			//2.write body
			if pack.Body != nil {
				_, err = io.Copy(s.conn, pack.Body)
				if err != nil {
					log.Error("write body", err)
					continue
				}
			}

		case <-s.shutdownCh:

		}
	}

}

func (s *Session) receive() {
	hdr := header(make([]byte, headerSize))
	for {
		time.Sleep(time.Millisecond * 300)
		if s.IsClose() {
			return
		}
		_, err := io.ReadFull(s.conn, hdr)
		if err != nil {
			log.Warn("read head ", err)
			continue
		}
		if !hdr.validate() {
			continue
		}
		if err := packageHandlers[hdr.MsgType()](s, hdr); err != nil {
			log.Error(err)
		}
	}
}

// PING
func (s *Session) handlePing(hdr header) error {
	flags := hdr.Flags()
	pingID := hdr.Length()

	//syn ,we should return ack
	if flags&flagSYN == flagSYN {
		log.Debug("get ping syn", pingID)
		hdr := header(make([]byte, headerSize))
		hdr.encode(PingPackage, flagACK, 0, pingID)
		s.sendCh <- Package{Hdr: hdr}
		return nil
	}

	// ack
	log.Debug("get ping ack:", pingID)
	s.pingLock.Lock()
	close(s.pings[pingID])
	delete(s.pings, pingID)
	s.pingLock.Unlock()
	return nil
}

func (s *Session) keepalive(interval time.Duration) {
	for {
		select {
		case <-time.After(interval):
			err := s.ping()
			if err != nil {
				log.Error("keepalive failed ", err)
			}

		case <-s.shutdownCh:
			return
		}
	}
}
func (s *Session) ping() error {
	hdr := header(make([]byte, headerSize))
	hdr.encode(PingPackage, flagSYN, 0, s.pingID)

	log.Debug("send ping syn:", s.pingID)
	s.sendCh <- Package{Hdr: hdr}
	pch := make(chan struct{})

	s.pingLock.Lock()
	s.pings[s.pingID] = pch
	s.pingLock.Unlock()
	s.pingID++

	//wait response
	select {
	case <-pch:
		//log.Debug("get ping response")
		return nil

	case <-time.After(WaitTimeout):
		return fmt.Errorf("wiat ping timeout")

	}
}

//STREAM
func (s *Session) handleStreamMessage(hdr header) error {
	id := hdr.StreamID()
	flags := hdr.Flags()
	//syn
	if flags&flagSYN == flagSYN {
		stream := NewStream(s, id,streamSYNReceived)

		s.streamLock.Lock()
		s.streams[id] = stream
		s.streamLock.Unlock()
		s.acceptCh <- stream
	}
    log.Debug("handleStreamMessage")
	s.streamLock.Lock()
	stream := s.streams[id]
	s.streamLock.Unlock()

	if err := stream.readData(hdr,flags,s.conn);err != nil{
		log.Errorf("readData %v",err)
		return err
	}
	return nil
}


//implemenet Listener interface
func (s *Session) Accept() (net.Conn, error) {
	conn, err := s.AcceptStream()
	if err != nil {
		return nil, err
	}
	return conn, err
}

func (s *Session) AcceptStream() (*Stream, error) {
	select {
	case stream := <-s.acceptCh:
		//if err := stream.sendWindowUpdate(); err != nil {
		//	return nil, err
		//}
		return stream, nil
	case <-s.shutdownCh:
		return nil, fmt.Errorf("listen close")
	}
}