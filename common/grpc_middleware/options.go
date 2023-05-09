// grpc_middleware
// @author LanguageY++2013 2023/5/9 16:54
// @company soulgame
package grpc_middleware

import (
	"time"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func ClientOptions()(opts []grpc.DialOption) {
	opts = []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			PermitWithoutStream: true,
			Time:                time.Minute,
			Timeout:             time.Second * 20,
		}),
	}

	return
}