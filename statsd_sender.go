package haproxystatsd

import "fmt"

type StatsdSender struct{}

func NewStatsdSender() *StatsdSender {
	return &StatsdSender{}
}

func (s *StatsdSender) Send(msgs []string) {
	for _, msg := range msgs {
		fmt.Println(msg)
	}
}
