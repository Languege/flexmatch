// api
// @author LanguageY++2013 2022/11/23 15:20
// @company soulgame
package match_api

import (
	resolver_etcd "github.com/Languege/flexmatch/common/grpc_middleware/resolver/etcd"
	common_constants "github.com/Languege/flexmatch/common/constants"
	"google.golang.org/grpc"
	"context"
	"github.com/Languege/flexmatch/service/match/proto/open"
	"log"
	"github.com/spf13/viper"
	"github.com/Languege/flexmatch/common/grpc_middleware"
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
	target := resolver_etcd.BuildTarget(viper.GetStringSlice("etcd.addrs"), common_constants.ServiceEndpoint_Match)
	conn, err := grpc.DialContext(context.TODO(), target, grpc_middleware.ClientOptions()...)
	if err != nil {
		return nil, err
	}

	return open.NewFlexMatchClient(conn), nil
}
