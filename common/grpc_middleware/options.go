// grpc_middleware
// @author LanguageY++2013 2023/5/9 16:54
// @company soulgame
package grpc_middleware

import (
	"time"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"go.uber.org/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"runtime/debug"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
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

	//日志可选项
	zapOpts := []grpc_zap.Option{
		grpc_zap.WithLevels(grpc_zap.DefaultCodeToLevel),
	}

	grpc_zap.ReplaceGrpcLogger(rpcLogger)

	unaryClientInterceptorList := []grpc.UnaryClientInterceptor{}
	//客户端zap日志拦截器
	unaryClientInterceptorList = append(unaryClientInterceptorList,
		grpc_zap.UnaryClientInterceptor(rpcLogger, zapOpts...),
	)

	opts = append(opts,grpc.WithChainUnaryInterceptor(unaryClientInterceptorList...))

	return
}

func recoveryFunc(p interface{}) error {
	stack := string(debug.Stack())
	err := status.Errorf(codes.Internal, fmt.Sprintf("Unexpected error:%+v \n traceinfo:\n=======\n%s", p, stack))
	rpcLogger.Warn("grpc recovery", zap.String("stack", stack))
	return err
}

func ServerOptions() (opts []grpc.ServerOption) {
	unaryServerInterceptorList := []grpc.UnaryServerInterceptor{}

	// 异常恢复
	recoverOpts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(recoveryFunc),
	}

	unaryServerInterceptorList = append(unaryServerInterceptorList, grpc_recovery.UnaryServerInterceptor(recoverOpts...))

	//zap日志可选项
	zapOpts := []grpc_zap.Option{
		grpc_zap.WithLevels(grpc_zap.DefaultCodeToLevel),
	}

	grpc_zap.ReplaceGrpcLogger(rpcLogger)

	unaryServerInterceptorList = append(unaryServerInterceptorList,
		grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_zap.UnaryServerInterceptor(rpcLogger, zapOpts...),
	)

	opts = []grpc.ServerOption{
		grpc_middleware.WithUnaryServerChain(
			unaryServerInterceptorList...
		),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
			Time:              time.Minute,
			Timeout:           time.Second * 20,
		}),
	}

	return opts
}