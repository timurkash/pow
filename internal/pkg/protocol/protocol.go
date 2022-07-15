package protocol

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	Quit              = iota // on quit each side (server or client) should close connection
	RequestChallenge         // from client to server - request new challenge from server
	ResponseChallenge        // from server to client - message with challenge for client
	RequestResource          // from client to server - message with solved challenge
	ResponseResource         // from server to client - message with useful info is solution is correct, or with error if not
)

type Message struct {
	Header  int    //type of message
	Payload string //payload, could be json, quote or be empty
}

func (m *Message) Stringify() string {
	return fmt.Sprintf("%d|%s", m.Header, m.Payload)
}

func ParseMessage(str string) (*Message, error) {
	str = strings.TrimSpace(str)
	var msgType int
	parts := strings.Split(str, "|")
	if len(parts) < 1 || len(parts) > 2 { //only 1 or 2 parts allowed
		return nil, fmt.Errorf("message doesn't match protocol")
	}
	msgType, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("cannot parse header")
	}
	msg := Message{
		Header: msgType,
	}
	if len(parts) == 2 {
		msg.Payload = parts[1]
	}
	return &msg, nil
}
