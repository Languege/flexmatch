package service

import (
	"context"
	"github.com/Languege/flexmatch/service/match/proto/open"
	"github.com/Languege/flexmatch/service/match/sgerrors"
	"github.com/Languege/flexmatch/service/match/repositories/matchmaking_rep"
)

/**
 *@author LanguageY++2013
 *2020/2/17 1:59 PM
 **/
type MatchServerImpl struct {
	open.UnimplementedFlexMatchServer
	//媒介仓储
	MatchmakingRep *matchmaking_rep.MatchmakingRepository
}

func NewMatchServerImpl() *MatchServerImpl {
	s := &MatchServerImpl{}

	s.MatchmakingRep = matchmaking_rep.NewMatchmakingRepository()

	return s
}

func (s *MatchServerImpl) CreateMatchmakingConfiguration(ctx context.Context, req *open.CreateMatchmakingConfigurationRequest) (*open.CreateMatchmakingConfigurationResponse, error) {
	resp := &open.CreateMatchmakingConfigurationResponse{}

	var sgErr error

	//参数校验
	sgErr = s.MatchmakingRep.CheckConfiguration(req.Configuration)
	if sgErr != nil {
		return resp, sgerrors.ConvertGRPCError(sgErr)
	}

	//添加配置
	_, sgErr = s.MatchmakingRep.SaveConfiguration(ctx, req.Configuration)
	if sgErr != nil {
		return resp, sgerrors.ConvertGRPCError(sgErr)
	}

	return resp, nil
}
func (s *MatchServerImpl) DescribeMatchmakingConfiguration(ctx context.Context, req *open.DescribeMatchmakingConfigurationRequest) (*open.DescribeMatchmakingConfigurationResponse, error) {
	resp := &open.DescribeMatchmakingConfigurationResponse{}

	var sgErr error
	resp.Configuartion, sgErr = s.MatchmakingRep.MatchmakingConf(ctx, req.ConfigurationName)
	if sgErr != nil {
		return resp, sgerrors.ConvertGRPCError(sgErr)
	}
	return resp, nil
}
func (s *MatchServerImpl) StartMatchmaking(ctx context.Context, req *open.StartMatchmakingRequest) (*open.StartMatchmakingResponse, error) {
	resp := &open.StartMatchmakingResponse{}
	var sgErr error
	sgErr = s.MatchmakingRep.StartMatchmaking(ctx, req.ConfigurationName, req.TicketId, req.Players)
	if sgErr != nil {
		return resp, sgerrors.ConvertGRPCError(sgErr)
	}

	return resp, nil
}
func (s *MatchServerImpl) DescribeMatchmaking(ctx context.Context, req *open.DescribeMatchmakingRequest) (*open.DescribeMatchmakingResponse, error) {
	resp := &open.DescribeMatchmakingResponse{}

	resp.TicketList = s.MatchmakingRep.DescribeMatchmaking(ctx, req.TicketIds)

	return resp, nil
}
func (s *MatchServerImpl) StopMatchmaking(ctx context.Context, req *open.StopMatchmakingRequest) (*open.StopMatchmakingResponse, error) {
	resp := &open.StopMatchmakingResponse{}
	var sgErr error
	sgErr = s.MatchmakingRep.StopMatchmaking(ctx, req.TicketId)
	if sgErr != nil {
		return resp, sgerrors.ConvertGRPCError(sgErr)
	}

	return resp, nil
}
func (s *MatchServerImpl) AcceptMatch(ctx context.Context, req *open.AcceptMatchRequest) (*open.AcceptMatchResponse, error) {
	resp := &open.AcceptMatchResponse{}
	var sgErr error
	sgErr = s.MatchmakingRep.AcceptMatch(ctx, req.ConfigurationName, req.TicketId, req.AcceptanceType, req.PlayerIds)
	if sgErr != nil {
		return resp, sgerrors.ConvertGRPCError(sgErr)
	}
	return resp, nil
}
