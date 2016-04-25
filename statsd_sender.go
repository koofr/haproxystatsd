package haproxystatsd

import (
	"bytes"
	"fmt"
	"net"
)

type StatsdSender struct {
	conn  *net.UDPConn
	msgCh chan []string
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
	s = &StatsdSender{
		conn:  conn,
		msgCh: make(chan []string, 1024),
	}
	go s.sendloop()
	return
}

func (s *StatsdSender) Send(msgs []string) {
	s.msgCh <- msgs
}

func (s *StatsdSender) sendloop() {

	buf := new(bytes.Buffer)
	for msgs := range s.msgCh {
		for _, msg := range msgs {
			buf.WriteString(msg)
			buf.WriteByte(10)
		}
		buf.WriteTo(s.conn)
		buf.Reset()
	}
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
