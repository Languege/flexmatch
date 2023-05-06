// parser
// @author LanguageY++2013 2023/4/3 20:44
// @company soulgame
package chain

import (
	"context"
)

const(
	CtxRulesKey = "rules"
	CtxTeamsKey = "teams"
	CtxReturnKey = "return"
)

type UserHandler func(ctx context.Context)

type UserHandlerWrapper func(ctx context.Context, handler UserHandler)

//ChainUserHandlerWrapper 构建调用链
func ChainUserHandlerWrapper(stack bool, interceptors ...UserHandlerWrapper) UserHandlerWrapper {
	n := len(interceptors)

	return func(ctx context.Context, handler UserHandler) {
		chainer := func(currentInter UserHandlerWrapper, currentHandler UserHandler) UserHandler {
			return func(currentCtx context.Context) {
				currentInter(currentCtx,  currentHandler)
			}
		}

		chainedHandler := handler
		if stack {
			for i := 0; i < n; i++ {
				chainedHandler = chainer(interceptors[i], chainedHandler)
			}
		}else{
			for i := n - 1; i >= 0; i-- {
				chainedHandler = chainer(interceptors[i], chainedHandler)
			}
		}


		chainedHandler(ctx)
	}
}


type Chain struct {
	wrappers []UserHandlerWrapper
}

func(c *Chain) Add(wrapper UserHandlerWrapper) {
	c.wrappers = append(c.wrappers, wrapper)
}

func(c *Chain) BuildStack() UserHandlerWrapper {
	return ChainUserHandlerWrapper(true, c.wrappers...)
}

func(c *Chain) BuildQueue() UserHandlerWrapper {
	return ChainUserHandlerWrapper(false, c.wrappers...)
}