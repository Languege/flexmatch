// pubsub
// @author LanguageY++2013 2023/5/16 17:11
// @company soulgame
package pubsub

import (
	"github.com/Languege/flexmatch/service/match/proto/open"
	"github.com/Languege/flexmatch/common/logger"
)

type eventHandler func(ev *open.MatchEvent) error

type EventHandlers map[open.MatchEventType]eventHandler

func(s *EventHandlers) RegisterEventHandler(evType open.MatchEventType, handler eventHandler) {
	if _, ok :=  (*s)[evType];ok {
		logger.DPanicf("%s handler has been registered previously", evType.String())
	}
	(*s)[evType] = handler
}

func (s *EventHandlers) Receive(ev *open.MatchEvent) error {
	defer func() {
		if r := recover();r != nil {
			logger.DPanicf("match event receive recover %s", r)
		}
	}()
	handler, ok := (*s)[ev.MatchEventType]
	if !ok {
		logger.Debugf("%s handler not registered", ev.MatchEventType.String())
		return nil
	}

	return handler(ev)
}
