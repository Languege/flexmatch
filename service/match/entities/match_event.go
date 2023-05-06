// entities
// @author LanguageY++2013 2022/11/9 14:37
// @company soulgame
package entities

import (
	"encoding/json"
	"github.com/Languege/flexmatch/service/match/proto/open"
	"log"
	"github.com/Shopify/sarama"
	"github.com/google/uuid"
	kafka_wrapper "github.com/Languege/flexmatch/service/match/wrappers/kafka"
)

type matchEventSubscriber func(topic string, ev *open.MatchEvent)

type matchEventSubscribeManager struct {
	topic string
	subscribers []matchEventSubscriber
}

func newMatchEventSubscribeManager(topic string) (m *matchEventSubscribeManager){
	m = &matchEventSubscribeManager{topic:topic}

	m.Register(matchEventPrint)
	m.Register(matchEventPushKafka)


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

func matchEventPrint(topic string, ev *open.MatchEvent) {
	data, _ := json.Marshal(ev)
	log.Printf("%s\n", string(data))
}

var kafkaProducer = kafka_wrapper.NewAsyncProducer()

func matchEventPushKafka(topic string, ev *open.MatchEvent) {
	key := ev.MatchId
	if key == "" {
		key = uuid.NewString()
	}
	data, _ := json.Marshal(ev)
	message := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(data),
	}

	kafka_wrapper.MessageAdaptor(message)

	kafkaProducer.Input() <- message
}