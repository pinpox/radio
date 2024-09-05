package main

import  "sync"

type messageBuffer struct {
	mu                sync.Mutex
	messages          [10]message
	firstMessageIndex int
}

type message struct {
	User string
	Text string
}

func (m *messageBuffer) Add(userID int, message string) {

	m.mu.Lock()
	defer m.mu.Unlock()

	m.messages[m.firstMessageIndex].User = usernames[userID]
	m.messages[m.firstMessageIndex].Text = message
	m.firstMessageIndex = (m.firstMessageIndex + 1) % 10

}

func (m *messageBuffer) Get() [10]message {
	ret := [10]message{}

	for i := 0; i < 10; i++ {

		msg := m.messages[(m.firstMessageIndex+i)%10]
		ret[9-i] = msg
	}
	return ret
}
