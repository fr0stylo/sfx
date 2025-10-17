package client

import (
	"context"

	"google.golang.org/protobuf/proto"
)

var pool = make(map[string]*Process)

func CallContext(ctx context.Context, path string, req proto.Message, resp proto.Message) error {
	mpt, ok := pool[path]
	if !ok {
		p, err := StartProcess(ctx, path)
		if err != nil {
			return err
		}
		mpt = p
		pool[path] = p
	}

	return mpt.Call(ctx, req, resp)
}
