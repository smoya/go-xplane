package messages

type Message struct {
	sequence uint64
}

func (m Message) Sequence() uint64 {
	return m.sequence
}
