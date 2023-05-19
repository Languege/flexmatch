// service
// @author LanguageY++2013 2023/5/19 15:03
// @company soulgame
package service

import (
	"github.com/Languege/flexmatch/service/match/proto/open"
	"context"
	"go.uber.org/atomic"
)

type FlexMatchGameImpl struct {
	open.UnimplementedFlexMatchGameServer
	roomIDGen 		atomic.Int64
}

func NewFlexMatchGameImpl() *FlexMatchGameImpl {
	return &FlexMatchGameImpl{}
}

func (s *FlexMatchGameImpl) CreateGameSession(ctx context.Context, req *open.CreateGameSessionRequest) (*open.CreateGameSessionResponse, error) {
	resp := &open.CreateGameSessionResponse{}

	resp.GameSession = &open.GameSession{
		GameSessionId: "battle#1@dev/1",
		SvcID:         "battle#1@dev",
		RoomID:        s.roomIDGen.Add(1),
	}

	return resp, nil
}


