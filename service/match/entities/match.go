// entities
// @author LanguageY++2013 2022/11/9 18:31
// @company soulgame
package entities

import (
	"context"
	"github.com/Languege/flexmatch/service/match/proto/open"
	"github.com/google/uuid"
	"time"
)

type Match struct {
	//匹配唯一ID
	MatchId string

	//当前的对局团队
	Teams []*Team

	//创建时间
	CreationTime int64

	//对局会话连接信息
	GameSessionConnectionInfo *open.GameSessionConnectionInfo

	//已经接受/拒绝的票据
	acceptTickets map[string]*open.MatchmakingTicket

	//票据接受通道
	acceptMatchChan chan string

	//票据拒绝通道
	rejectMatchChan chan string

	//媒介
	matchmaking *Matchmaking
}

func NewMatch(matchmaking *Matchmaking, matchTeams []*Team) (match *Match) {
	//形成匹配
	match = &Match{
		MatchId:         uuid.New().String(),
		Teams:           matchTeams,
		CreationTime:    time.Now().Unix(),
		acceptMatchChan: make(chan string, 1),
		rejectMatchChan: make(chan string, 1),
		acceptTickets:   map[string]*open.MatchmakingTicket{},
		matchmaking:     matchmaking,
	}

	return
}

// CheckAccepted 检测对局团队内的所有成员是否均接受了对局
func (m *Match) CheckAccepted() bool {
	for _, team := range m.Teams {
		for _, ticket := range team.Tickets {
			for _, player := range ticket.Players {
				if player.Accepted == false {
					return false
				}
			}
		}
	}

	return true
}

// StartAccept 开始接受用户接受状态检测协程
func (m *Match) StartAccept(acceptanceTimeoutSeconds int64, placingMatchChan chan<- *Match) {
	//开启独立协程定时检测对局用户是否接受
	go func() {
		deadline := time.Tick(time.Duration(acceptanceTimeoutSeconds) * time.Second)
		for {
			select {
			case <-deadline:
				//匹配信息删除
				m.matchmaking.MatchRemove(m.MatchId)
				//接受超时,退出协程
				completedEvent := &open.MatchEvent{
					MatchEventType:            open.MatchEventType_AcceptMatchCompleted,
					Tickets:                   m.AcceptTickets(),
					MatchId:                   m.MatchId,
					AcceptanceCompletedReason: "TimeOut",
				}

				m.matchmaking.eventSubs.MatchEventInput(completedEvent)
				return
			case ticketId := <-m.acceptMatchChan:
				//进行过票据接受/拒绝的票据
				var acceptTicket *open.MatchmakingTicket
				allTickets := m.AllTickets()
				for _, ticket := range allTickets {
					if ticket.TicketId == ticketId {
						acceptTicket = ticket

						goto checkAccepted
					}
				}

			checkAccepted:
				if acceptTicket == nil {
					//票据不存在（可能因为票据ID不唯一导致）
					break
				}

				//已接受或拒绝票据保存
				m.acceptTickets[acceptTicket.TicketId] = acceptTicket
				//AcceptMatch事件上报
				acceptMatchEvent := &open.MatchEvent{
					MatchEventType: open.MatchEventType_AcceptMatch,
					Tickets:        m.AcceptTickets(),
					MatchId:        m.MatchId,
				}

				m.matchmaking.eventSubs.MatchEventInput(acceptMatchEvent)

				if m.CheckAccepted() {
					//匹配信息删除 TODO:是否等待游戏结束再做删除？
					m.matchmaking.MatchRemove(m.MatchId)

					//对局票据状态修改为PLACING
					for _, ticket := range allTickets {
						ticket.Status = open.MatchmakingTicketStatus_PLACING.String()
					}

					//玩家均已结束，上报匹配完成事件（因为接受）
					completedEvent := &open.MatchEvent{
						MatchEventType:            open.MatchEventType_AcceptMatchCompleted,
						Tickets:                   allTickets,
						MatchId:                   m.MatchId,
						AcceptanceCompletedReason: "Acceptance",
					}

					m.matchmaking.eventSubs.MatchEventInput(completedEvent)

					//为对局准备游戏会话
					placingMatchChan <- m

					//接受阶段结束，退出协程
					return
				}
			case ticketId := <-m.rejectMatchChan:
				//进行过票据接受/拒绝的票据
				var rejectTicket *open.MatchmakingTicket
				allTickets := m.AllTickets()
				for _, ticket := range allTickets {
					if ticket.TicketId == ticketId {
						rejectTicket = ticket

						goto checkReject
					}
				}
			checkReject:
				if rejectTicket == nil {
					//票据不存在（可能因为票据ID不唯一导致）
					break
				}
				//匹配信息删除
				m.matchmaking.MatchRemove(m.MatchId)

				//已接受或拒绝票据保存
				rejectTicket.Status = open.MatchmakingTicketStatus_CANCELLED.String()
				rejectTicket.StatusReason = "RejectMatch"
				m.acceptTickets[rejectTicket.TicketId] = rejectTicket
				//AcceptMatch事件上报
				acceptMatchEvent := &open.MatchEvent{
					MatchEventType: open.MatchEventType_AcceptMatch,
					Tickets:        m.AcceptTickets(),
					MatchId:        m.MatchId,
				}

				m.matchmaking.eventSubs.MatchEventInput(acceptMatchEvent)

				//上报接收匹配完成事件
				acceptMatchCompleted := &open.MatchEvent{
					MatchEventType:            open.MatchEventType_AcceptMatchCompleted,
					Tickets:                   allTickets,
					MatchId:                   m.MatchId,
					AcceptanceCompletedReason: "Rejection",
				}

				m.matchmaking.eventSubs.MatchEventInput(acceptMatchCompleted)


				if m.matchmaking.Conf.BackfillMode == open.BackfillMode_AUTOMATIC.String() {
					//匹配自动回填 （拒绝的票据除外）
					for _, ticket := range allTickets {
						if ticket.TicketId != rejectTicket.TicketId {
							m.matchmaking.TicketQueued(ticket)
						}
					}
				}

				//协成退出
				return
			}
		}
	}()
}

func (m *Match) AcceptTickets() (ret []*open.MatchmakingTicket) {
	ret = make([]*open.MatchmakingTicket, 0, len(m.acceptTickets))
	for _, ticket := range m.acceptTickets {
		ret = append(ret, ticket)
	}

	return
}

//AllTickets 所有票据
func (m *Match) AllTickets() (tickets []*open.MatchmakingTicket) {
	tickets = make([]*open.MatchmakingTicket, 0, len(m.Teams))
	for _, team := range m.Teams {
		tickets = append(tickets, team.Tickets...)
	}

	return
}

// StartGameSession 开启游戏会话
func (m *Match) StartGameSession() {
	resp, err := gameClient.CreateGameSession(context.TODO(), &open.CreateGameSessionRequest{
		MatchId:                   m.MatchId,
		GameProperties:            m.matchmaking.Conf.GameProperties,
		GameSessionData:           m.matchmaking.Conf.GameSessionData,
		Name:                      m.matchmaking.Conf.Name,
		MaximumPlayerSessionCount: m.matchmaking.rsw.MatchPlayerNum,
	})
	if err != nil {

		allTickets := m.AllTickets()
		//票据状态变更
		for _, ticket := range allTickets {
			ticket.Status = open.MatchmakingTicketStatus_FAILED.String()
			ticket.StatusReason = "CreateGameSessionFailed"
			ticket.StatusMessage = err.Error()
		}
		//匹配失败事件上报
		failedEvent := &open.MatchEvent{
			MatchEventType: open.MatchEventType_MatchmakingFailed,
			Tickets:        allTickets,
			MatchId:        m.MatchId,
			Reason:         "CreateGameSessionFailed",
			Message:        err.Error(),
		}

		m.matchmaking.eventSubs.MatchEventInput(failedEvent)

		return
	}

	m.GameSessionConnectionInfo = &open.GameSessionConnectionInfo{
		GameSessionId: resp.GameSession.GameSessionId,
		SvcID:         resp.GameSession.SvcID,
		RoomID:        resp.GameSession.RoomID,
	}

	for _, team := range m.Teams {
		for _, ticket := range team.Tickets {
			ticket.GameSessionInfo = m.GameSessionConnectionInfo
			for _, player := range ticket.Players {
				m.GameSessionConnectionInfo.Players = append(m.GameSessionConnectionInfo.Players,
					&open.MatchedPlayerSession{
						UserId: player.UserId,
						Team:   team.Conf.Name,
					})
			}
			//更新状态为COMPLETED
			ticket.Status = open.MatchmakingTicketStatus_COMPLETED.String()
		}
	}

	//更新对局票据状态成功
	succeedEvent := &open.MatchEvent{
		MatchEventType:  open.MatchEventType_MatchmakingSucceeded,
		Tickets:         m.AcceptTickets(),
		MatchId:         m.MatchId,
		GameSessionInfo: m.GameSessionConnectionInfo,
	}

	m.matchmaking.eventSubs.MatchEventInput(succeedEvent)
}

func (m *Match) BuildGameSessionInfo() *open.GameSessionConnectionInfo {
	return m.GameSessionConnectionInfo
}

// AcceptMatch 玩家接受对局
func (m *Match) AcceptMatch(ticketId string) {
	m.acceptMatchChan <- ticketId
}

// RejectMatch 玩家拒绝对局
func (m *Match) RejectMatch(ticketId string) {
	m.rejectMatchChan <- ticketId
}
