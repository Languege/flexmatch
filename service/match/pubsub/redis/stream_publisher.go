// redis
// @author LanguageY++2013 2023/5/12 09:27
// @company soulgame
package redis

import (
	"github.com/Languege/flexmatch/service/match/proto/open"
	redis_wrapper "github.com/Languege/flexmatch/common/wrappers/redis"
	"encoding/json"
)

type RedisStreamPublisher struct {
	redis *redis_wrapper.RedisWrapper
}

func NewRedisStreamPublisher(conf redis_wrapper.Configure)  *RedisStreamPublisher {
	p := &RedisStreamPublisher{}
	p.redis = redis_wrapper.NewRedisWrapper(conf)

	return p
}

func(p RedisStreamPublisher) Name() string {
	return "redis_stream"
}

func(p RedisStreamPublisher) Send(topic string, ev *open.MatchEvent) error {
	if ev.MatchEventType == open.MatchEventType_MatchmakingQueued {
		return nil
	}
	data, err := json.Marshal(ev)
	if err != nil {
		return err
	}

	entry := &redis_wrapper.ByteStreamEntry{
		Data: data,
	}

	return p.redis.XAdd(topic, entry)
}