// timeout
// @author LanguageY++2013 2023/5/19 18:27
// @company soulgame
package timeout

import (
	"time"
	"context"
	"google.golang.org/grpc"
)

func TimeoutInterceptor(timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if timeout <= 0 {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		realTimeout := timeout
		if deadline, ok := ctx.Deadline(); ok {
			leftTime := time.Until(deadline)
			if leftTime < timeout {
				realTimeout = leftTime
			}
		}

		ctx, cancel :=  context.WithDeadline(ctx, time.Now().Add(realTimeout))
		defer cancel()

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

