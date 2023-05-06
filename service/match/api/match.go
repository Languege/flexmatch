// api
// @author LanguageY++2013 2022/11/23 15:20
// @company soulgame
package match_api

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"context"
	"github.com/Languege/flexmatch/service/match/proto/open"
	"log"
)

var(
	endpoint string
	FlexMatchClient open.FlexMatchClient
)

func init() {
	if viper.IsSet("rpc.endpoints.match") {
		endpoint = viper.GetString("rpc.endpoints.match")
	}else{
		endpoint = viper.GetString("project") + ".match.rpc"
	}
	var err error
	FlexMatchClient, err = newMatchClient()
	if err != nil {
		log.Panicln(err)
	}
}

func newMatchClient() (open.FlexMatchClient,error) {
	conn, err := grpc.DialContext(context.TODO(), "127.0.0.1:10007")
	if err != nil {
		return nil, err
	}

	return open.NewFlexMatchClient(conn), nil
}
