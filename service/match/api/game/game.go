// api
// @author LanguageY++2013 2022/11/23 15:20
// @company soulgame
package match_game_api

import (
	"google.golang.org/grpc"
	"context"
	"github.com/Languege/flexmatch/service/match/proto/open"
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
	conn, err := grpc.DialContext(context.TODO(), "127.0.0.1:10007")
	if err != nil {
		return nil, err
	}

	return open.NewFlexMatchGameClient(conn), nil
}
