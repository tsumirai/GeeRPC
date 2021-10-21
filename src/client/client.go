package client

import (
	"errors"
	geerpc "gee-rpc/src/codec"
	"gee-rpc/src/codec/codec"
	"io"
	"sync"
)

// Call represents an active RPC.
type Call struct {
	Seq           uint64
	ServiceMethod string      // format "<service>.<method>"
	Args          interface{} // arguments to the function
	Reply         interface{} // reply from the function
	Error         error       // if error occurs, it will be set
	Done          chan *Call  // strobes when call is complete
}

func (call *Call) done() {
	call.Done <- call
}

// Client represents an RPC Client.
// There may be multiple outstanding Calls associated
// with a single Client,  and a Client may be used by
// multiple goroutines simultaneously
type Client struct {
	cc       codec.Codec
	opt      *geerpc.Option
	sending  sync.Mutex // protect following
	header   codec.Header
	mu       sync.Mutex // protect following
	seq      uint64
	pending  map[uint64]*Call
	closing  bool // user has called Close
	shutdown bool // server has told us to stop
}

var _ io.Closer = (*Client)(nil)

var ErrShutdown = errors.New("connection is shut down")

// Close the connection
func (client *Client) Close() error {
	client.mu.Lock()
	defer client.mu.Unlock()

	if client.closing {
		return ErrShutdown
	}
	client.closing = true
	return client.cc.Close()
}

// IsAvailable return true if the client does work
func (client *Client) IsAvailable() bool {
	client.mu.Lock()
	defer client.mu.Unlock()
	return !client.shutdown && !client.closing
}
