package callback

type CallbackChannel interface {
	Send(req *CallbackRequest)
	Get() <-chan *CallbackRequest
}

type callbackChannel struct {
	sendCh chan *CallbackRequest
}

func NewChannel() CallbackChannel {
	return &callbackChannel{
		sendCh: make(chan *CallbackRequest),
	}
}

func (s *callbackChannel) Send(req *CallbackRequest) {
	s.sendCh <- req
}

func (s *callbackChannel) Get() <-chan *CallbackRequest {
	return s.sendCh
}
