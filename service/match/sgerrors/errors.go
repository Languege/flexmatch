package sgerrors

import (
	"fmt"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
	"github.com/Languege/flexmatch/service/match/proto/open"
	"runtime"
)

/**
 *@author LanguageY++2013
 *2020/2/16 11:33 PM
 **/
type SGError interface {
	error
	Code() open.ResultCode
	Message() string
}

type PBSGError struct {
	code 	open.ResultCode
	msg 	string
}

func(e PBSGError) Error() string {
	return fmt.Sprintf("code=%s msg=%s", e.code.String(), e.msg)
}

func(e PBSGError) Code() open.ResultCode {
	return e.code
}

func(e PBSGError) Message() string {
	return e.msg
}




func NewSGError(code open.ResultCode, params... interface{}) PBSGError {
	var msg string
	if len(params) > 0  {
		if v, ok := params[0].(string);ok {
			msg = v
		}
	}else{
		msg = code.String()
	}

	_, file, line, _ := runtime.Caller(1)
	msg = fmt.Sprintf("%s\n%s:%d", msg, file, line)

	return PBSGError{code:code, msg:msg}
}

// SGError转换为grpc error
func NewGRpcError(sgErr SGError) error {
	msg := sgErr.Message()
	_, file, line, _ := runtime.Caller(1)
	msg = fmt.Sprintf("%s\n%s:%d", msg, file, line)

	return status.Error(codes.Code(sgErr.Code()), msg)
}

// SGError转换为grpc error
func NewGRpcErrorWithResultCode(code open.ResultCode, params... interface{}) error {
	var msg string
	if len(params) > 0  {
		if v, ok := params[0].(string);ok {
			msg = v
		}
	}else{
		msg = code.String()
	}

	_, file, line, _ := runtime.Caller(1)
	msg = fmt.Sprintf("%s\n%s:%d", msg, file, line)

	return status.Error(codes.Code(code), msg)
}

func NewErrorWithResultCode(code open.ResultCode) error {
	msg := code.String()
	_, file, line, _ := runtime.Caller(1)
	msg = fmt.Sprintf("%s\n%s:%d", msg, file, line)

	return status.Error(codes.Code(code), msg)
}

//err转code
func Convert(err error)  open.ResultCode {
	if v, ok := err.(interface{
		Code() open.ResultCode
	});ok {
		return v.Code()
	}

	return open.ResultCode_ErrUnknown
}

//err尝试转换成grpc error
func ConvertGRPCError(err error) error {
	if v, ok := err.(interface{
		Code() open.ResultCode
		Message() string
	});ok {
		return status.Error(codes.Code(v.Code()), v.Message())
	}

	return err
}