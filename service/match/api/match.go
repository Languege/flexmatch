// api
// @author LanguageY++2013 2022/11/23 15:20
// @company soulgame
package match_api

import (
	"google.golang.org/grpc"
	"context"
	"github.com/Languege/flexmatch/service/match/proto/open"
	"log"
	"fmt"
	"github.com/spf13/viper"
)

var(
	FlexMatchClient open.FlexMatchClient
)

func init() {
	var err error
	FlexMatchClient, err = newMatchClient()
	if err != nil {
		log.Panicln(err)
	}
}

func newMatchClient() (open.FlexMatchClient,error) {
	addr := fmt.Sprintf("127.0.0.1:%d", viper.GetInt("rpc.port"))
	conn, err := grpc.DialContext(context.TODO(), addr)
	if err != nil {
		return nil, err
	}

	return open.NewFlexMatchClient(conn), nil
}
