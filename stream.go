package mulitstream

import (
	"bytes"
	"io"
	"sync"
	"time"
)

type Stream struct {
	id      uint32
	session *Session

	recvBuf  *bytes.Buffer
	recvLock sync.Mutex
	recvNotifyCh chan struct{}

	state streamState
	stateLock sync.Mutex
}

func NewStream(session *Session, id uint32, state streamState) *Stream {
	return &Stream{
		id: id,
		session: session,
		state: state,
		recvNotifyCh: make(chan struct{}),
	}
}



func (s *Stream) Write(b []byte) (n int, err error) {
	return  s.write(b)
}

func (s *Stream) write(b []byte) (n int, err error) {
	body := bytes.NewReader(b)
	hdr := header(make([]byte, headerSize))
	flag := s.sendFlags()
	hdr.encode(DataPackage,flag,s.id,uint32(len(b)))
	s.session.sendCh <- Package{Hdr: hdr,Body: body}
	return len(b),err
}

func (s *Stream) Read(b []byte) (n int, err error) {

START:
	s.recvLock.Lock()
	if s.recvBuf == nil || s.recvBuf.Len() == 0 {
		s.recvLock.Unlock()
		goto WAIT
	}

	n,_= s.recvBuf.Read(b)
	s.recvLock.Unlock()

	return  n,nil

WAIT:
	select {
	case <-s.recvNotifyCh:
		goto START
	}

	return 0,nil
}



func (s *Stream) readData(hdr header, flags uint16, conn io.Reader) error {
	length := hdr.Length()
	if length == 0 {
		return nil
	}

	conn = &io.LimitedReader{R: conn, N: int64(length)}

	s.recvLock.Lock()


	if s.recvBuf == nil {
		s.recvBuf = bytes.NewBuffer(make([]byte, 0, length))
	}
    if _,err :=io.Copy(s.recvBuf,conn);err != nil{
		s.recvLock.Unlock()
	    return err
   }
   s.recvLock.Unlock()
	asyncNotify(s.recvNotifyCh)
   return nil
}


func (s *Stream) Close() error {
	log.Debug("Close  be called")
	return nil
}
func (s *Stream) SetDeadline(t time.Time) error {
	log.Debug("SetDeadline be called",t)
	return nil
}

func (s *Stream) SetReadDeadline(t time.Time) error {
	log.Debug("SetReadDeadline be called",t)
	asyncNotify(s.recvNotifyCh)
	return nil
}
func (s *Stream) SetWriteDeadline(t time.Time) error {
	log.Debug("SetWriteDeadline be called",t)
	asyncNotify(s.recvNotifyCh)
	return nil
}


func asyncNotify(ch chan struct{}) {
	select {
	case ch <- struct{}{}:
	default:
	}
}

func (s *Stream) sendFlags() uint16 {
	s.stateLock.Lock()
	defer s.stateLock.Unlock()
	var flags uint16
	switch s.state {
	case streamInit:
		flags |= flagSYN
		s.state = streamSYNSent
	case streamSYNReceived:
		flags |= flagACK
		s.state = streamEstablished
	}
	return flags
}