package main

import (
	"fmt"
	"strings"
)

type Message struct {
	Text string
	Age  int
}

func (m *Message) String() string {
	return strings.ToUpper(m.Text)
}
func (m *Message) Useless() string {
	return strings.ToUpper(m.Text)
}

func main() {

	msg := &Message{Text: "Hello, World!"}
	fmt.Println(msg.String())

}
