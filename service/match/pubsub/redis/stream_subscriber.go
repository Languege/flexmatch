// redis
// @author LanguageY++2013 2023/5/16 17:06
// @company soulgame
package redis

import (
	"fmt"
	"github.com/Languege/flexmatch/common/logger"
	redis_wrapper "github.com/Languege/flexmatch/common/wrappers/redis"
	"github.com/Languege/flexmatch/service/match/pubsub"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"time"
	"encoding/json"
	"github.com/Languege/flexmatch/service/match/proto/open"
	"os"
)

const (
	defaultReadTimeOut = time.Second * 30
)

type RedisStreamSubscriber struct {
	redis *redis_wrapper.RedisWrapper
	pubsub.EventHandlers
	topic       string
	group       string
	consumer    string
	readTimeout time.Duration
}

type Option func(s *RedisStreamSubscriber)

func WithConsumer(consumer string) Option {
	return func(s *RedisStreamSubscriber) {
		s.consumer = consumer
	}
}

func WithReadTimeout(timeout time.Duration) Option {
	return func(s *RedisStreamSubscriber) {
		s.readTimeout = timeout
	}
}

func NewRedisStreamSubscriber(conf redis_wrapper.Configure, topic string, group string, opts ...Option) *RedisStreamSubscriber {
	s := &RedisStreamSubscriber{
		redis:         redis_wrapper.NewRedisWrapper(conf),
		EventHandlers: pubsub.EventHandlers{},
		topic:         topic,
		group:         group,
		consumer:      uuid.NewString(),
		readTimeout:   defaultReadTimeOut,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s RedisStreamSubscriber) Name() string {
	return "redis_stream"
}

func(s RedisStreamSubscriber) consumerGroupExist(infos []*redis_wrapper.ConsumerGroup) bool {
	for _, info := range infos {
		if info.Name == s.group {
			return true
		}
	}

	return false
}

func (s RedisStreamSubscriber) Start() {
	infos, _ := s.redis.XInfoGroups(s.topic)
	if !s.consumerGroupExist(infos) {
		reply, err := s.redis.XGroupCreateFromBeginning(s.topic, s.group)
		if err != nil {
			logger.Panic(reply, err)
		}
	}

	go s.handleEvents()
}


func(s RedisStreamSubscriber) handleEvents() {
	for {
		ml := []*redis_wrapper.ByteStreamEntry{}
		err := s.redis.XReadGroup(s.group, s.consumer, 10, s.readTimeout, s.topic, &ml)
		if err != nil {
			if os.IsTimeout(err) {
				logger.Infof("read group %s %s", s.group, err)
			}else{
				logger.DPanicf("%s XREAD,  group %s topic %s err %s", s.consumer, s.group, s.topic, err)
			}
			continue
		}

		for _, msg := range ml {
			ev := &open.MatchEvent{}
			err := json.Unmarshal(msg.Data, ev)
			if err != nil {
				logger.Errorf("json unmarshal  '%s' err %s", string(msg.Data), err)
				goto MarkMessage
			}

			if err = s.Receive(ev); err != nil {
				logger.Errorw(fmt.Sprintf("handle match event err %s ", err), zap.Any("ev", ev))
			}

		MarkMessage:
			_, err = s.redis.XAck(s.group, s.topic, msg.ID)
			if err != nil {
				logger.DPanicf("%s XACK ID %s , group %s topic %s err %s", s.consumer, msg.ID, s.group, s.topic, err)
			}
		}
	}
}