package callback

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestChannel(t *testing.T) {
	channel := NewChannel()

	go func() {
		channel.Send(&CallbackRequest{
			FileID: "1",
		})
	}()

	select {
	case req := <-channel.Get():
		assert.Equal(t, "1", req.FileID)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected to receive a request, but channel was empty")
	}
}
