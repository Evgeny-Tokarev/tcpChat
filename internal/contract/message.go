package contract

import (
	"errors"
)

const (
	nickBLen         = 32
	bodyBLen         = 256
	bytesMessageBLen = nickBLen + bodyBLen
)

var ErrInvalidBytesMessage = errors.New("invalid bytes message")

type BytesMessage [bytesMessageBLen]byte

type Message struct {
	Nick string
	Body string
}

func (bm *BytesMessage) Bytes() []byte {
	b := make([]byte, bytesMessageBLen)
	for i := 0; i < bytesMessageBLen; i++ {
		b[i] = bm[i]
	}
	return b
}

func (msg *Message) ToBytesMessage() *BytesMessage {
	b := new(BytesMessage)
	for i := 0; i < nickBLen; i++ {
		if i < len(msg.Nick) {
			b[i] = msg.Nick[i]
		} else {
			b[i] = 0
		}
	}

	for i, j := nickBLen, 0; i < bytesMessageBLen; i, j = i+1, j+1 {
		if j < len(msg.Body) {
			b[i] = msg.Body[j]
		} else {
			b[i] = 0
		}
	}
	return b

}

func (bm *BytesMessage) ToMessage() *Message {
	nick := make([]byte, 0, nickBLen)
	body := make([]byte, 0, bodyBLen)

	for i := 0; i < nickBLen; i++ {
		if bm[i] != 0 {
			nick = append(nick, bm[i])
		}
	}
	for i := nickBLen; i < bytesMessageBLen; i++ {
		if bm[i] != 0 {
			body = append(body, bm[i])
		}
	}

	return &Message{
		Nick: string(nick),
		Body: string(body),
	}
}

func NewBytesMessage(msg []byte) (*BytesMessage, error) {
	if len(msg) < bytesMessageBLen {
		return nil, ErrInvalidBytesMessage
	}
	b := new(BytesMessage)
	for i := 0; i < bytesMessageBLen; i++ {
		b[i] = msg[i]
	}
	return b, nil
}
