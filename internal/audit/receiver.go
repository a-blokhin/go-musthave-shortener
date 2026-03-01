package audit

type Receiver interface {
	Receive(event Event) error
}
