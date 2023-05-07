// entities
// @author LanguageY++2013 2022/11/9 14:37
// @company soulgame
package entities

import (
	"github.com/Languege/flexmatch/service/match/proto/open"
	"github.com/Languege/flexmatch/service/match/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//初始订阅设置
var _subscribers  = []matchEventSubscriber{}

func init() {
	RegisterSubscribe(matchEventPrint)
}

func RegisterSubscribe(sb matchEventSubscriber) {
	_subscribers = append(_subscribers, sb)
}

type matchEventSubscriber func(topic string, ev *open.MatchEvent)

type matchEventSubscribeManager struct {
	topic string
	subscribers []matchEventSubscriber
}

func newMatchEventSubscribeManager(topic string) (m *matchEventSubscribeManager){
	m = &matchEventSubscribeManager{topic:topic}

	m.subscribers = append(m.subscribers, _subscribers...)

	return m
}

func(m *matchEventSubscribeManager) Register(sb matchEventSubscriber) {
	m.subscribers = append(m.subscribers, sb)
}

func(m *matchEventSubscribeManager) MatchEventInput(ev *open.MatchEvent) {
	for _, subscriber := range m.subscribers {
		subscriber(m.topic, ev)
	}
}

type matchEventMarshal struct {
	open.MatchEvent
}

func newMatchEventMarshal(ev *open.MatchEvent) *matchEventMarshal {
	return &matchEventMarshal{
		MatchEvent:*ev,
	}
}

func(m *matchEventMarshal) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt64("AcceptanceTimeout", m.AcceptanceTimeout)

	return nil
}

func matchEventPrint(topic string, ev *open.MatchEvent) {
	logger.Infow("", zap.String("topic", topic), zap.String("MatchEventType", ev.MatchEventType.String()),
		zap.Object("ev", newMatchEventMarshal(ev)))
}
