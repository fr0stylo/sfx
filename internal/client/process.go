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

type Process struct {
	cmd *exec.Cmd
	in  io.WriteCloser
	out io.ReadCloser
}

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

func (p *Process) Call(ctx context.Context, req proto.Message, resp proto.Message) error {
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

func (p *Process) Close() error {
	_ = p.in.Close()
	_ = p.out.Close()

	return p.cmd.Wait()
}
