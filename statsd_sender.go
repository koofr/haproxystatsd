package haproxystatsd

import (
	"bytes"
	"fmt"
	"net"
)

type StatsdSender struct {
	conn *net.UDPConn
}

func NewStatsdSender(addr string) (s *StatsdSender, err error) {
	serverAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return
	}
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		return
	}
	s = &StatsdSender{conn}
	return
}

func (s *StatsdSender) Send(msgs []string) {
	var buf *bytes.Buffer

	for _, msg := range msgs {
		buf.WriteString(msg)
		buf.WriteByte(10)
	}

	s.conn.Write(buf.Bytes())
}

type MockStatsdSender struct{}

func NewMockStatsdSender() *MockStatsdSender {
	return &MockStatsdSender{}
}

func (s *MockStatsdSender) Send(msgs []string) {
	for _, msg := range msgs {
		fmt.Println(msg)
	}
}
