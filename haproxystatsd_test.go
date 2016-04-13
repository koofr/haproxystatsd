package haproxystatsd

import (
	"fmt"
	"net"
	"testing"
	"time"
)

const (
	SyslogAddr    = "127.0.0.1:10514"
	ExampleSyslog = `<150>Apr 12 03:19:03 haproxy[31393]: 31.15.130.214:47670 [12/Apr/2016:03:19:03.225] http-frontend cekarcek/neon 1/0/0/1/2 301 407 - - ---- 0/0/0/0/0 0/0 "GET / HTTP/1.1"`
)

var (
	defaultConfig = &Config{
		SyslogBindAddr: "127.0.0.1:10514",
	}
)

type MockSender struct {
	Ch chan []string
}

func (s *MockSender) Send(msgs []string) {
	s.Ch <- msgs
}

func NewMockSender() *MockSender {
	return &MockSender{make(chan []string)}
}

func sendSyslogMsg(msg string) (err error) {
	serverAddr, err := net.ResolveUDPAddr("udp", SyslogAddr)
	if err != nil {
		return
	}
	con, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		return
	}
	_, err = con.Write([]byte(ExampleSyslog))
	return
}

func TestOneLogMessage(t *testing.T) {
	hs, err := New(defaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	sender := NewMockSender()
	hs.sender = sender
	if err := hs.Boot(); err != nil {
		t.Fatal(err)
	}

	sendSyslogMsg(ExampleSyslog)

	select {
	case msgs := <-sender.Ch:
		fmt.Println(msgs)

		return
	case <-time.After(time.Second):
		t.Fatal("No message received")
	}

}

func TestParsePrefixKeys(t *testing.T) {

	tplStr := "{{.frontend_name}}.{{.backend_name}}"

	keys := parsePrefixKeys(tplStr)

	fmt.Println(keys)
}
