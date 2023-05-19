// test
// @author LanguageY++2013 2022/11/23 15:30
// @company soulgame
package test

import (
	_ "github.com/Languege/flexmatch/service/match/conf"
	_ "github.com/Languege/flexmatch/common/bootstraps"
	"context"
	"github.com/Languege/flexmatch/service/match/proto/open"
	match_api "github.com/Languege/flexmatch/service/match/api"
	"github.com/google/uuid"
	"testing"
	"time"
	kafka_pubsub "github.com/Languege/flexmatch/service/match/pubsub/kafka"
	redis_pubsub "github.com/Languege/flexmatch/service/match/pubsub/redis"
	"github.com/spf13/viper"
	"github.com/Languege/flexmatch/common/logger"
	redis_wrapper "github.com/Languege/flexmatch/common/wrappers/redis"
	"sync"
	"sync/atomic"
)

func defaultConf() *open.MatchmakingConfiguration {
	conf := &open.MatchmakingConfiguration{
		Name:                     "5v5ranking",
		AcceptanceRequired:       false,
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
					Name:        "maxScoreDistance",
					Type:        open.MatchmakingRuleType_MatchmakingRuleType_Distance,
					MaxDistance: 100,
					ReferenceValue: "min(flatten(teams[*].players.attributes[score]))",
					Measurements: "flatten(teams[*].players.attributes[score])",
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

func TestCreateMatchmakingConfiguration(t *testing.T) {
	req := &open.CreateMatchmakingConfigurationRequest{
		Configuration: defaultConf(),
	}
	req.Configuration.AcceptanceRequired = false
	_, err := match_api.FlexMatchClient.CreateMatchmakingConfiguration(context.TODO(), req)
	if err != nil {
		t.Fatal(err)
	}
}

func TestStartMatchmaking(t *testing.T) {
	for i := 0; i < 10;i++ {
		ticket := newTicket()
		req := &open.StartMatchmakingRequest{
			ConfigurationName: "5v5ranking",
			TicketId:          ticket.TicketId,
			Players:           ticket.Players,
		}

		_, err := match_api.FlexMatchClient.StartMatchmaking(context.TODO(), req)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestMatchEventConsume(t *testing.T) {
	N := 500
	var successCount, timeoutCount, matchCount int64
	done := make(chan struct{}, 1)
	sub, err := kafka_pubsub.NewKafkaSubscriber(viper.GetStringSlice("publishers.kafka.bootstrapServers"),
		[]string{defaultConf().MatchEventQueueTopic}, "matcheventpushtoclient")
	if err != nil {
		t.Fatal(err)
	}

	sub.RegisterEventHandler(open.MatchEventType_MatchmakingSearching, func(ev *open.MatchEvent) error {
		logger.Info("遍历票据向用户通知匹配开始")
		return nil
	})

	sub.RegisterEventHandler(open.MatchEventType_MatchmakingCancelled, func(ev *open.MatchEvent) error {
		logger.Info("告知用户匹配取消")
		return nil
	})

	sub.RegisterEventHandler(open.MatchEventType_PotentialMatchCreated, func(ev *open.MatchEvent) error {
		logger.Info("广播对局已找到")
		if ev.AcceptanceRequired {
			logger.Info("等待玩家接收对局")
		}
		return nil
	})

	sub.RegisterEventHandler(open.MatchEventType_AcceptMatch, func(ev *open.MatchEvent) error {
		logger.Info("广播对局内所有玩家的接受状态")
		return nil
	})

	sub.RegisterEventHandler(open.MatchEventType_AcceptMatchCompleted, func(ev *open.MatchEvent) error {
		logger.Info("广播对局内所有玩家接受阶段结束，若结束原因为为超时，客户端退回组队页面，若为任意拒绝，拒绝的玩家（组队）退出")
		return nil
	})

	sub.RegisterEventHandler(open.MatchEventType_MatchmakingSucceeded, func(ev *open.MatchEvent) error {
		defer func() {
			newCount :=  atomic.AddInt64(&matchCount, 1)
			if newCount >= int64(N) {
				close(done)
			}
		}()
		atomic.AddInt64(&successCount, 1)
		logger.Info("广播匹配成功通知，包含对局连接信息，客户端进行服务绑定，进入对局服务器房间")
		return nil
	})

	sub.RegisterEventHandler(open.MatchEventType_MatchmakingTimedOut, func(ev *open.MatchEvent) error {
		logger.Info("通知票据内玩家匹配超时")
		return nil
	})

	sub.RegisterEventHandler(open.MatchEventType_MatchmakingFailed, func(ev *open.MatchEvent) error {
		if ev.MatchId != "" {
			defer func() {
				newCount :=  atomic.AddInt64(&matchCount, 1)
				if newCount >= int64(N) {
					close(done)
				}
			}()
		}

		logger.Info("通知票据内玩家匹配失败")
		return nil
	})


	sub.Start()

	st := time.Now()
	for i := 0; i < N;i++ {
		for i := 0; i < 10;i++ {
			ticket := newTicket()
			req := &open.StartMatchmakingRequest{
				ConfigurationName: "5v5ranking",
				TicketId:          ticket.TicketId,
				Players:           ticket.Players,
			}

			_, err := match_api.FlexMatchClient.StartMatchmaking(context.TODO(), req)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	<-done


	t.Logf("success %d timeout %d", successCount, timeoutCount)
	cost := time.Now().Sub(st).Nanoseconds() / int64(N)
	t.Log(cost) //N=50时，28938031 ns;N=200时, 4769162 ns;N=500时，3759681 ns

	time.Sleep(time.Second)
}

func TestMatchEventConsumeUseRedisStream(t *testing.T) {
	ch := make(chan struct{}, 1)
	conf := redis_wrapper.Configure{}
	err := viper.UnmarshalKey("publishers.redis", &conf)
	if err != nil {
		logger.Panicf("redis.publisher unmarshal err %s", err)
	}
	sub := redis_pubsub.NewRedisStreamSubscriber(conf,
		defaultConf().MatchEventQueueTopic, "matcheventpushtoclient", redis_pubsub.WithConsumer("consumer01"))
	if err != nil {
		t.Fatal(err)
	}

	sub.RegisterEventHandler(open.MatchEventType_MatchmakingSearching, func(ev *open.MatchEvent) error {
		logger.Info("遍历票据向用户通知匹配开始")
		return nil
	})

	sub.RegisterEventHandler(open.MatchEventType_MatchmakingCancelled, func(ev *open.MatchEvent) error {
		logger.Info("告知用户匹配取消")
		return nil
	})

	sub.RegisterEventHandler(open.MatchEventType_PotentialMatchCreated, func(ev *open.MatchEvent) error {
		logger.Info("广播对局已找到，等待玩家接受（无需接受自动跳过）")
		return nil
	})

	sub.RegisterEventHandler(open.MatchEventType_AcceptMatch, func(ev *open.MatchEvent) error {
		logger.Info("广播对局内所有玩家的接受状态")
		return nil
	})

	sub.RegisterEventHandler(open.MatchEventType_AcceptMatchCompleted, func(ev *open.MatchEvent) error {
		logger.Info("广播对局内所有玩家接受阶段结束，若结束原因为为超时，客户端退回组队页面，若为任意拒绝，拒绝的玩家（组队）退出")
		return nil
	})

	sub.RegisterEventHandler(open.MatchEventType_MatchmakingSucceeded, func(ev *open.MatchEvent) error {
		logger.Info("广播匹配成功通知，包含对局连接信息，客户端进行服务绑定，进入对局服务器房间")
		ch <- struct{}{}
		return nil
	})

	sub.RegisterEventHandler(open.MatchEventType_MatchmakingTimedOut, func(ev *open.MatchEvent) error {
		logger.Info("通知票据内玩家匹配超时")
		ch <- struct{}{}
		return nil
	})

	sub.RegisterEventHandler(open.MatchEventType_MatchmakingFailed, func(ev *open.MatchEvent) error {
		logger.Info("通知票据内玩家匹配失败")
		ch <- struct{}{}
		return nil
	})

	for i := 0; i < 10;i++ {
		ticket := newTicket()
		req := &open.StartMatchmakingRequest{
			ConfigurationName: "5v5ranking",
			TicketId:          ticket.TicketId,
			Players:           ticket.Players,
		}

		_, err := match_api.FlexMatchClient.StartMatchmaking(context.TODO(), req)
		if err != nil {
			t.Fatal(err)
		}
	}

	sub.Start()

	<- ch
}


func TestMatchEventConsumeUseRedisStreamPerformance(t *testing.T) {
	N := 1000
	var successCount, timeoutCount int64
	wg := &sync.WaitGroup{}
	conf := redis_wrapper.Configure{}
	err := viper.UnmarshalKey("publishers.redis", &conf)
	if err != nil {
		logger.Panicf("redis.publisher unmarshal err %s", err)
	}
	sub := redis_pubsub.NewRedisStreamSubscriber(conf,
		defaultConf().MatchEventQueueTopic, "matcheventpushtoclient",
		redis_pubsub.WithConsumer("consumer01"),
		redis_pubsub.WithGoroutineNum(50))
	if err != nil {
		t.Fatal(err)
	}

	sub.RegisterEventHandler(open.MatchEventType_MatchmakingSearching, func(ev *open.MatchEvent) error {
		logger.Info("遍历票据向用户通知匹配开始")
		return nil
	})

	sub.RegisterEventHandler(open.MatchEventType_MatchmakingCancelled, func(ev *open.MatchEvent) error {
		logger.Info("告知用户匹配取消")
		return nil
	})

	sub.RegisterEventHandler(open.MatchEventType_PotentialMatchCreated, func(ev *open.MatchEvent) error {
		logger.Info("广播对局已找到，等待玩家接受（无需接受自动跳过）")
		return nil
	})

	sub.RegisterEventHandler(open.MatchEventType_AcceptMatch, func(ev *open.MatchEvent) error {
		logger.Info("广播对局内所有玩家的接受状态")
		return nil
	})

	sub.RegisterEventHandler(open.MatchEventType_AcceptMatchCompleted, func(ev *open.MatchEvent) error {
		logger.Info("广播对局内所有玩家接受阶段结束，若结束原因为为超时，客户端退回组队页面，若为任意拒绝，拒绝的玩家（组队）退出")
		return nil
	})

	sub.RegisterEventHandler(open.MatchEventType_MatchmakingSucceeded, func(ev *open.MatchEvent) error {
		defer wg.Done()
		atomic.AddInt64(&successCount, 1)
		logger.Info("广播匹配成功通知，包含对局连接信息，客户端进行服务绑定，进入对局服务器房间")
		return nil
	})

	sub.RegisterEventHandler(open.MatchEventType_MatchmakingTimedOut, func(ev *open.MatchEvent) error {
		defer wg.Done()
		atomic.AddInt64(&timeoutCount, 1)
		logger.Info("通知票据内玩家匹配超时")
		return nil
	})

	sub.RegisterEventHandler(open.MatchEventType_MatchmakingFailed, func(ev *open.MatchEvent) error {
		logger.Info("通知票据内玩家匹配失败")
		return nil
	})


	sub.Start()

	st := time.Now()
	for i := 0; i < N;i++ {
		wg.Add(1)
		for j := 0; j < 10; j++ {
			ticket := newTicket()
			req := &open.StartMatchmakingRequest{
				ConfigurationName: "5v5ranking",
				TicketId:          ticket.TicketId,
				Players:           ticket.Players,
			}

			_, err := match_api.FlexMatchClient.StartMatchmaking(context.TODO(), req)
			if err != nil {
				t.Fatal(err)
			}
		}

	}


	wg.Wait()

	t.Logf("success %d timeout %d", successCount, timeoutCount)
	cost := time.Now().Sub(st).Nanoseconds() / int64(N)
	t.Log(cost) //N=50时，32965389 ns;N=200时, 34534875 ns;N=500时，32311304 ns
}
