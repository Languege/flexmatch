// sgerrors
// @author LanguageY++2013 2023/5/6 10:13
// @company soulgame
package sgerrors

import (
	"testing"
	"github.com/Languege/flexmatch/service/match/proto/open"
	"google.golang.org/grpc/status"
)

func TestNewGRpcErrorWithResultCode(t *testing.T) {
	sgErr := NewSGError(open.ResultCode_AccessBindFailure, "ttt")
	t.Log(sgErr.Error())
	//转换成grpc error
	grpcErr := ConvertGRPCError(sgErr)
	t.Log(grpcErr)

	t.Log(Convert(grpcErr))

	err := NewGRpcError(sgErr)
	t.Log(err)
	code := status.Code(err)
	t.Log(int32(code))
}