// entities
// @author LanguageY++2013 2022/11/9 14:37
// @company soulgame
package entities

import (
	"github.com/Languege/flexmatch/service/match/proto/open"
	"encoding/json"
	"log"
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
	for _, subscriber := range _subscribers {
		subscriber(m.topic, ev)
	}
}

func matchEventPrint(topic string, ev *open.MatchEvent) {
	data, _ := json.Marshal(ev)
	log.Printf("%s\n", string(data))
}
