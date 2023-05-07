// entities
// @author LanguageY++2013 2022/11/11 17:39
// @company soulgame
package entities

import (
	"github.com/Languege/flexmatch/service/match/proto/open"
	"github.com/google/uuid"
	"github.com/juju/errors"
	"testing"
	"time"
	"sync"
	"sync/atomic"
	"log"
)

func defaultConf() *open.MatchmakingConfiguration {
	conf := &open.MatchmakingConfiguration{
		Name:                     "5v5ranking",
		AcceptanceRequired:       true,
		AcceptanceTimeoutSeconds: 10,
		Description:              "",
		GameProperties:           []*open.GameProperty{},
		GameSessionData:          "{}",
		RequestTimeoutSeconds:    30,
		MatchEventQueueTopic:     "MatchEventQueueTopic",
		RuleSet: &open.MatchmakingRuleSet{
			PlayerAttributes: []*open.PlayerAttribute{
				{Name: "score", Type: "float64"},
			},
			Teams: []*open.MatchmakingTeamConfiguration{
				{Name: "red", PlayerNumber: 5},
				{Name: "blue", PlayerNumber: 5},
			},
			Rules: []*open.MatchmakingRule{
				{
					Name:           "maxScoreDistance",
					Type:           open.MatchmakingRuleType_MatchmakingRuleType_Distance,
					MaxDistance:    100,
					ReferenceValue: "min(flatten(teams[*].players.attributes[score]))",
					Measurements:   "flatten(teams[*].players.attributes[score])",
				},
			},
			Expansions: []*open.MatchmakingExpansionRule{
				{
					Target: &open.MatchmakingExpansionRuleTarget{
						ComponentType: open.ComponentType_ComponentType_Rules,
						ComponentName: "maxScoreDistance",
						Attribute:     "MaxDistance",
					},
				},
			},
			Algorithm: &open.MatchmakingRuleAlgorithm{
				BatchingPreference: "sorted",
				SortByAttributes:   []string{"score"},
			},
		},
	}

	return conf
}

func newTicket() *open.MatchmakingTicket {
	return &open.MatchmakingTicket{
		TicketId:  uuid.New().String(),
		StartTime: time.Now().Unix(),
		Players: []*open.MatchPlayer{
			{UserId: int64(uuid.New().ID()), Attributes: []*open.PlayerAttribute{{Name: "score", Type: "float64", Value: "1200"}}},
		},
	}
}

func TestMatchmaking_TicketInput(t *testing.T) {
	conf := defaultConf()
	//conf.AcceptanceRequired = false
	matchmaking := NewMatchmaking(conf)

	wg := &sync.WaitGroup{}
	matchmaking.eventSubs.Register(func(topic string, ev *open.MatchEvent) {
		switch ev.MatchEventType {
		case open.MatchEventType_PotentialMatchCreated:
			go func() {
				wg.Done()
				if ev.AcceptanceRequired {
					//循环接收对局
					for _, ticket := range ev.Tickets {
						for _, player := range ticket.Players {
							time.Sleep(time.Millisecond * 100)
							err := matchmaking.AcceptMatch(ticket.TicketId, player.UserId, open.AcceptanceType_ACCEPT)
							if err != nil {
								t.Fatal(errors.ErrorStack(err))
							}
						}
					}
				}
			}()
		}
	})

	st := time.Now()
	N := 1
	wg.Add(N)
	for i := 0; i < N; i++ {
		for j := 0;j < 10; j++ {
			err := matchmaking.TicketInput(newTicket())
			if err != nil {
				t.Fatal(err)
			}
		}
	}

	wg.Wait()
	c := atomic.LoadInt64(&matchmaking.MatchSucceedCounter)
	log.Printf("当前匹配成功对局数%d\n", c)
	cost := time.Now().Sub(st).Nanoseconds() / int64(N)
	t.Log(cost) //N=50时，20695223 ns;N=200时, 5586288 ns;N=2000时，1338885 ns
}

//取消票据然后匹配自动回填测试
func TestMatchmaking_TicketCancel(t *testing.T) {
	conf := defaultConf()
	conf.BackfillMode = open.BackfillMode_AUTOMATIC.String()
	matchmaking := NewMatchmaking(conf)

	matchmaking.eventSubs.Register(func(topic string, ev *open.MatchEvent) {
		switch ev.MatchEventType {
		case open.MatchEventType_PotentialMatchCreated:
			if ev.AcceptanceRequired {
				//循环接收对局
				for _, ticket := range ev.Tickets {
					//最后一个票据拒绝
					acceptType := open.AcceptanceType_ACCEPT

					for _, player := range ticket.Players {
						time.Sleep(time.Millisecond * 100)
						err := matchmaking.AcceptMatch(ticket.TicketId, player.UserId, acceptType)
						if err != nil {
							t.Fatal(errors.ErrorStack(err))
						}
					}
				}
			}
		case open.MatchEventType_AcceptMatch:
			for _, ticket := range ev.Tickets {
				if ticket.CancelRequest {
					go func() {
						time.Sleep(time.Second)
						//重新加入
						ticket.TicketId = uuid.NewString()
						err := matchmaking.TicketInput(ticket)
						if err != nil {
							t.Fatal(err)
						}
					}()

					break
				}
			}
		}
	})

	for i := 0; i < 11; i++ {
		ticket := newTicket()
		err := matchmaking.TicketInput(ticket)
		if err != nil {
			t.Fatal(err)
		}

		go func(t *open.MatchmakingTicket) {
			time.Sleep(time.Second * 5)
			matchmaking.StopMatch(t.TicketId)
		}(ticket)
	}

	time.Sleep(time.Hour)
}

//取消票据然后匹配自动回填测试
func TestMatchmaking_BackfillMode(t *testing.T) {
	conf := defaultConf()
	conf.BackfillMode = open.BackfillMode_AUTOMATIC.String()
	conf.AcceptanceTimeoutSeconds = 1000
	matchmaking := NewMatchmaking(conf)
	var rejectMatchTicketId  string
	matchmaking.eventSubs.Register(func(topic string, ev *open.MatchEvent) {
		switch ev.MatchEventType {
		case open.MatchEventType_PotentialMatchCreated:
			if ev.AcceptanceRequired {
				//循环接收对局
				for _, ticket := range ev.Tickets {
					//最后一个票据拒绝
					acceptType := open.AcceptanceType_ACCEPT
					if ticket.TicketId == rejectMatchTicketId {
						acceptType = open.AcceptanceType_REJECT
					}

					for _, player := range ticket.Players {
						time.Sleep(time.Millisecond * 100)
						err := matchmaking.AcceptMatch(ticket.TicketId, player.UserId, acceptType)
						if err != nil {
							t.Fatal(errors.ErrorStack(err))
						}
						if acceptType == open.AcceptanceType_REJECT {
							break
						}
					}
				}
			}
		case open.MatchEventType_AcceptMatch:
			for _, ticket := range ev.Tickets {
				if ticket.StatusReason == "RejectMatch" && ticket.Status == open.MatchmakingTicketStatus_CANCELLED.String() {
					go func(ticket *open.MatchmakingTicket) {
						time.Sleep(time.Second)
						//重新加入
						//ticket.TicketId = uuid.NewString()
						//err := matchmaking.TicketInput(ticket)
						//if err != nil {
						//	t.Fatal(err)
						//}
					}(ticket)
				}
			}
		}
	})

	for i := 0; i < 10; i++ {
		ticket := newTicket()
		err := matchmaking.TicketInput(ticket)
		if err != nil {
			t.Fatal(err)
		}

		if i == 9 {
			rejectMatchTicketId = ticket.TicketId
		}
	}

	time.Sleep(time.Hour)
}