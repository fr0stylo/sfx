package plugin

import (
	"fmt"
	"os"

	"sfx/internal/rpc"
)

// Request is the simplified input provided to plugin handlers.
type Request struct {
	Ref     string
	Options []byte
}

// Response is the simplified output expected from plugin handlers.
type Response struct {
	Value []byte
}

// Handler processes a single Request and returns the corresponding Response.
type Handler interface {
	Handle(Request) (Response, error)
}

// HandlerFunc adapts a function to the Handler interface.
type HandlerFunc func(Request) (Response, error)

// Handle calls f(req).
func (f HandlerFunc) Handle(req Request) (Response, error) {
	return f(req)
}

// Run wires stdin/stdout to the protobuf transport and invokes the provided handler.
func Run(h Handler) {
	req := &rpc.SecretRequest{}
	if err := rpc.ReadDelimited(os.Stdin, req); err != nil {
		writeError(fmt.Errorf("decode request: %w", err))
		return
	}

	resp, err := h.Handle(Request{Ref: req.GetRef(), Options: req.GetOptions()})
	if err != nil {
		writeError(err)
		return
	}

	if err := rpc.WriteDelimited(os.Stdout, &rpc.SecretResponse{Value: resp.Value}); err != nil {
		fmt.Fprintf(os.Stderr, "write response: %v\n", err)
	}
}

func writeError(err error) {
	_ = rpc.WriteDelimited(os.Stdout, &rpc.SecretResponse{Error: err.Error()})
}
