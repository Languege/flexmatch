// chain
// @author LanguageY++2013 2023/4/3 22:12
// @company soulgame
package chain

import (
	"testing"
	"log"
	"context"
)

func TestChain_Build(t *testing.T) {
	chain := &Chain{}

	chain.Add(func(ctx context.Context, handler UserHandler) {
		log.Println("return min")
		handler(ctx)
	})

	chain.Add(func(ctx context.Context, handler UserHandler) {
		log.Println("return avg")
		handler(ctx)
	})

	t.Run("queue", func(t *testing.T) {
		chainHandler := chain.BuildQueue()

		chainHandler(context.TODO(), func(ctx context.Context) {
			log.Println("user handler")
		})
	})

	t.Run("stack", func(t *testing.T) {
		chainHandler := chain.BuildStack()

		chainHandler(context.TODO(), func(ctx context.Context) {
			log.Println("user handler")
		})
	})



}
