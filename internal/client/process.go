package client

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"google.golang.org/protobuf/proto"

	"github.com/fr0stylo/sfx/internal/rpc"
)

// Process owns a spawned plugin binary and the pipes used for RPC communication.
type Process struct {
	cmd *exec.Cmd
	in  io.WriteCloser
	out io.ReadCloser
}

// StartProcess launches the plugin binary at path and returns a Process wrapper.
func StartProcess(ctx context.Context, path string) (*Process, error) {
	cmd := exec.CommandContext(ctx, path)
	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr

	w, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	r, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &Process{
		cmd: cmd,
		in:  w,
		out: r,
	}, nil
}

// Call performs a round-trip protobuf exchange with the running process.
func (p *Process) Call(ctx context.Context, req proto.Message, resp proto.Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if req != nil {
		if err := rpc.WriteDelimited(p.in, req); err != nil {
			_ = p.cmd.Wait()
			return fmt.Errorf("send request: %w", err)
		}
	}

	if err := rpc.ReadDelimited(p.out, resp); err != nil {
		_ = p.cmd.Wait()
		return fmt.Errorf("read response: %w", err)
	}

	return nil
}

// Close closes IO pipes and waits for the process to exit.
func (p *Process) Close() error {
	_ = p.in.Close()
	_ = p.out.Close()

	return p.cmd.Wait()
}
