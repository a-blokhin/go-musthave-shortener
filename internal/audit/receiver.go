// Package audit provides audit logging functionality for the URL shortener service.
// It supports multiple receivers for audit events (file, HTTP, etc.).
package audit

// Receiver defines the interface for receiving audit events.
// Implementations can write audit events to files, send them to HTTP endpoints, etc.
type Receiver interface {
	// Receive processes an audit event.
	//
	// Parameters:
	//   - event: the audit event to process
	//
	// Returns an error if the event cannot be processed.
	Receive(event Event) error
}
