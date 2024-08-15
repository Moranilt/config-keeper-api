package callback

type CallbackChannel interface {
	// Send sends a CallbackRequest to the channel.
	Send(req *CallbackRequest)

	// Get returns a receive-only channel of CallbackRequest.
	Get() <-chan *CallbackRequest
}

type callbackChannel struct {
	sendCh chan *CallbackRequest
}

// NewChannel creates a new CallbackChannel with the specified buffer capacity.
// The returned CallbackChannel implementation is safe for concurrent use.
func NewChannel(cap int) CallbackChannel {
	return &callbackChannel{
		sendCh: make(chan *CallbackRequest, cap),
	}
}

func (s *callbackChannel) Send(req *CallbackRequest) {
	s.sendCh <- req
}

func (s *callbackChannel) Get() <-chan *CallbackRequest {
	return s.sendCh
}
