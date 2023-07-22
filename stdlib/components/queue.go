package components

import (
	"context"
)

type Callback_fn func([]byte)

type Queue interface {
	Send(ctx context.Context, msg []byte) error
	Recv(callback Callback_fn)
}