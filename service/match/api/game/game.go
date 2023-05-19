// api
// @author LanguageY++2013 2022/11/23 15:20
// @company soulgame
package match_game_api

import (
	"google.golang.org/grpc"
	"context"
	"github.com/Languege/flexmatch/service/match/proto/open"
	resolver_etcd "github.com/Languege/flexmatch/common/grpc_middleware/resolver/etcd"
	"github.com/spf13/viper"
	common_constants "github.com/Languege/flexmatch/common/constants"
	"github.com/Languege/flexmatch/common/grpc_middleware"
)


var(
	FlexMatchGameClient open.FlexMatchGameClient
)

func init() {
	var err error
	FlexMatchGameClient, err = newMatchGameClient()
	if err != nil {
//		log.Panicln(err)
	}
}

func newMatchGameClient() (open.FlexMatchGameClient,error) {
	battleEndpoint := common_constants.ServiceEndpoint_Battle
	if v := viper.GetString("rpc.endpoints.matchgame"); v != "" {
		battleEndpoint =  v
	}
	target := resolver_etcd.BuildTarget(viper.GetStringSlice("etcd.addrs"), battleEndpoint)
	conn, err := grpc.DialContext(context.TODO(), target, grpc_middleware.ClientOptions()...)
	if err != nil {
		return nil, err
	}

	return open.NewFlexMatchGameClient(conn), nil
}
