package client

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"google.golang.org/protobuf/proto"

	"github.com/fr0stylo/sfx/internal/rpc"
)

// Call executes the binary at path, sending the protobuf request and decoding the response.
// The response parameter must be a pointer to the expected message type.
func Call(ctx context.Context, path string, req proto.Message, resp proto.Message) error {
	if resp == nil {
		return errors.New("client: response message must not be nil")
	}

	cmd := exec.CommandContext(ctx, path)
	cmd.Env = os.Environ()
	
	ip, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	defer ip.Close()

	op, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start command: %w", err)
	}

	if req != nil {
		if err := rpc.WriteDelimited(ip, req); err != nil {
			_ = cmd.Wait()
			return fmt.Errorf("send request: %w", err)
		}
	}
	if err := ip.Close(); err != nil {
		return err
	}

	if err := rpc.ReadDelimited(op, resp); err != nil {
		_ = cmd.Wait()
		return fmt.Errorf("read response: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}
