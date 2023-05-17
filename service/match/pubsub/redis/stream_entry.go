// redis
// @author LanguageY++2013 2023/5/16 17:00
// @company soulgame
package redis

import (
	redis_wrapper "github.com/Languege/flexmatch/common/wrappers/redis"
	"github.com/Languege/flexmatch/service/match/proto/open"
)

type StreamEntry struct {
	redis_wrapper.StreamBaseMsg
	open.MatchEvent
	MatchEventType int `redis:"MatchEventType"`
}
