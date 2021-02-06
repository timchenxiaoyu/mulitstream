package mulitstream

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	flagSYN uint16 = 1 << iota

	flagACK

	flagFIN

	flagRST
)

//package type
const (
	DataPackage uint8 = iota
	WindowPackage
	PingPackage
	FinalPackage
)

type streamState int

const (
	streamInit streamState = iota
	streamSYNSent
	streamSYNReceived
	streamEstablished
	streamLocalClose
	streamRemoteClose
	streamClosed
	streamReset
)


//package header
const (
	protoVersion uint8 = 0
)

type header []byte

var headerSize = 12

func (h header) Version() uint8 {
	return h[0]
}

func (h header) MsgType() uint8 {
	return h[1]
}

func (h header) Flags() uint16 {
	return binary.BigEndian.Uint16(h[2:4])
}

func (h header) StreamID() uint32 {
	return binary.BigEndian.Uint32(h[4:8])
}

func (h header) Length() uint32 {
	return binary.BigEndian.Uint32(h[8:12])
}

func (h header) String() string {
	return fmt.Sprintf("Vsn:%d Type:%d Flags:%d StreamID:%d Length:%d",
		h.Version(), h.MsgType(), h.Flags(), h.StreamID(), h.Length())
}

func (h header) encode(msgType uint8, flags uint16, streamID uint32, length uint32) {
	h[0] = protoVersion
	h[1] = msgType
	binary.BigEndian.PutUint16(h[2:4], flags)
	binary.BigEndian.PutUint32(h[4:8], streamID)
	binary.BigEndian.PutUint32(h[8:12], length)
}
func (h header) validate() bool {
	if h.Version() != protoVersion {
		log.Warn("proto version not support")
		return false
	}
	mt := h.MsgType()
	if mt < DataPackage || mt > FinalPackage {
		return false
	}
	return true

}

type Package struct {
	Hdr  []byte
	Body io.Reader
}
