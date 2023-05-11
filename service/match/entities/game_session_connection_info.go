// entities
// @author LanguageY++2013 2022/11/9 18:42
// @company soulgame
package entities

import (
	"github.com/Languege/flexmatch/service/match/proto/open"
	"context"
	"go.uber.org/atomic"
	"google.golang.org/grpc"
)

var gameClient open.FlexMatchGameClient

type MockGameClient struct {
	roomIDGen 		*atomic.Int64
}


func NewMockGameClient() *MockGameClient {
	return &MockGameClient{
		roomIDGen:  &atomic.Int64{},
	}
}

func(c *MockGameClient) CreateGameSession(ctx context.Context, in *open.CreateGameSessionRequest, opts ...grpc.CallOption) (*open.CreateGameSessionResponse, error) {
	resp := &open.CreateGameSessionResponse{}

	resp.GameSession = &open.GameSession{
		GameSessionId: "battle#1@dev/1",
		SvcID:         "battle#1@dev",
		RoomID:        c.roomIDGen.Add(1),
	}

	return resp, nil
}

//设置游戏grpc客户端
func SetGameClient(cli open.FlexMatchGameClient) {
	gameClient = cli
}
