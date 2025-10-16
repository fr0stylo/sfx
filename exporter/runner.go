package exporter

import (
	"fmt"
	"os"

	"sfx/internal/rpc"
)

// Request is the data made available to exporter handlers.
type Request struct {
	Values  map[string][]byte
	Options []byte
}

// Handler processes the provided values and returns an error if export fails.
type Handler interface {
	Handle(Request) (Response, error)
}

// Response represents the binary payload produced by the exporter.
type Response struct {
	Payload []byte
}

// HandlerFunc adapts a function to the Handler interface.
type HandlerFunc func(Request) (Response, error)

// Handle calls f(req).
func (f HandlerFunc) Handle(req Request) (Response, error) {
	return f(req)
}

// Run wires stdin/stdout to the protobuf transport and invokes the provided handler.
func Run(h Handler) {
	req := &rpc.ExportRequest{}
	if err := rpc.ReadDelimited(os.Stdin, req); err != nil {
		writeError(fmt.Errorf("decode request: %w", err))
		return
	}

	resp, err := h.Handle(Request{Values: req.GetValues(), Options: req.GetOptions()})
	if err != nil {
		writeError(err)
		return
	}

	if err := rpc.WriteDelimited(os.Stdout, &rpc.ExportResponse{Payload: resp.Payload}); err != nil {
		fmt.Fprintf(os.Stderr, "write response: %v\n", err)
	}
}

func writeError(err error) {
	_ = rpc.WriteDelimited(os.Stdout, &rpc.ExportResponse{Error: err.Error()})
}
