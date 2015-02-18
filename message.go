package gostick

// Message type enables a simple way of sending data to Tellstick devices.
type Message struct {
	data []byte
}

// NewMessage returns a new Message from string s
func NewMessage(s string) Message {
	m := make([]byte, len(s))
	copy(m, s)
	return Message{data: m}
}

// Read implements Reader interface
func (m Message) Read(p []byte) (int, error) {
	c := copy(p, m.data)
	return c, nil
}
