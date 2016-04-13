package haproxystatsd

import (
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
		expected := []string{"ares.http-frontend.cekarcek.neon.Tc:0|g",
			"ares.http-frontend.cekarcek.neon.3xx:301|c",
			"ares.http-frontend.cekarcek.neon.feconn:0|g",
			"ares.http-frontend.cekarcek.neon.srv_queue:0|g",
			"ares.http-frontend.cekarcek.neon.Tq:1|g",
			"ares.http-frontend.cekarcek.neon.Tr:1|g",
			"ares.http-frontend.cekarcek.neon.Tt:2|g",
			"ares.http-frontend.cekarcek.neon.actconn:0|g",
			"ares.http-frontend.cekarcek.neon.beconn:0|g",
			"ares.http-frontend.cekarcek.neon.srv_conn:0|g",
			"ares.http-frontend.cekarcek.neon.retries:0|g",
			"ares.http-frontend.cekarcek.neon.backend_queue:0|g",
			"ares.http-frontend.cekarcek.neon.Tw:0|g",
		}

		for _, e := range expected {
			if sliceContains(msgs, e) == false {
				t.Errorf("Expected message %s not found", e)
				t.Fail()
			}
		}

		return
	case <-time.After(time.Second):
		t.Error("No message received")
		t.Fail()
	}

}

func TestParsePrefixKeys(t *testing.T) {
	tplStr := "{{.frontend_name}}.{{.backend_name}}"
	keys := parsePrefixKeys(tplStr)
	if (len(keys) == 2 && sliceContains(keys, "frontend_name") && sliceContains(keys, "backend_name")) == false {
		t.Fail()
	}
}

func TestStatusCode2Class(t *testing.T) {
	for i := 100; i < 200; i++ {
		if statusCode2Class(i) != "1xx" {
			t.Fail()
		}
	}
	for i := 200; i < 300; i++ {
		if statusCode2Class(i) != "2xx" {
			t.Fail()
		}
	}
	for i := 300; i < 400; i++ {
		if statusCode2Class(i) != "3xx" {
			t.Fail()
		}
	}
	for i := 400; i < 500; i++ {
		if statusCode2Class(i) != "4xx" {
			t.Fail()
		}
	}
	for i := 500; i < 500; i++ {
		if statusCode2Class(i) != "5xx" {
			t.Fail()
		}
	}
	if statusCode2Class(666) != "xxx" {
		t.Fail()
	}
}

func TestParseInt(t *testing.T) {
	if parseInt("-1") != -1 {
		t.Fail()
	}
	if parseInt("42") != 42 {
		t.Fail()
	}
	if parseInt(" 200 ") != 200 {
		t.Fail()
	}
	if parseInt("abc") != 0 {
		t.Fail()
	}
}
