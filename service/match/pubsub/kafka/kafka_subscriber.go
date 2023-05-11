// kafka
// @author LanguageY++2013 2023/5/11 09:54
// @company soulgame
package kafka

import (
	"encoding/json"
	"fmt"
	"github.com/Languege/flexmatch/common/logger"
	"github.com/Languege/flexmatch/service/match/proto/open"
	"github.com/Shopify/sarama"
	"go.uber.org/zap"
	kafka_wrapper "github.com/Languege/flexmatch/common/wrappers/kafka"
)
type eventHandler func(ev *open.MatchEvent) error

type KafkaSubscriber struct {
	brokerList    []string
	topics        []string
	group         string
	eventHandlers map[open.MatchEventType]eventHandler
}

func NewKafkaSubscriber(brokerList []string, topics []string, group string) (*KafkaSubscriber, error) {
	s := &KafkaSubscriber{
		brokerList: brokerList,
		topics:     topics,
		group:      group,
		eventHandlers: map[open.MatchEventType]eventHandler{},
	}

	return s, nil
}



func (s *KafkaSubscriber) Setup(ses sarama.ConsumerGroupSession) error {
	logger.Infof("KafkaSubscriber setup, %s starting...", ses.MemberID())

	return nil
}

func (s *KafkaSubscriber) Cleanup(ses sarama.ConsumerGroupSession) error {
	logger.Info("KafkaSubscriber cleanup, %s stopping...", ses.MemberID())

	return nil
}

func (s *KafkaSubscriber) ConsumeClaim(ses sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) (retErr error) {
	//recover
	defer func() {
		if r := recover(); r != nil {
			retErr = fmt.Errorf("%v", r)
			logger.Errorf("KafkaSubscribe ConsumerClaim Err %s", retErr)
			return
		}
	}()

	logger.Infof("start consumer from topic %s, partitions %d, initialOffset %d", claim.Topic(), claim.Partition(), claim.InitialOffset())
	ch := claim.Messages()
	for {
		msg, ok := <-ch
		if !ok {
			return fmt.Errorf("stop consumer due to topic %s partition %d messages chan closed", claim.Topic(), claim.Partition())
		}

		ev := &open.MatchEvent{}
		err := json.Unmarshal(msg.Value, ev)
		if err != nil {
			logger.Errorf("json unmarshal  '%s' err %s", string(msg.Value), err)
			ses.MarkMessage(msg, "")
			continue
		}

		if err = s.Receive(ev); err != nil {
			logger.Errorw(fmt.Sprintf("handle match event err %s ", err), zap.Any("ev", ev))
		}

		ses.MarkMessage(msg, "")
	}
}

func (s KafkaSubscriber) Name() string {
	return "kafka"
}

func (s *KafkaSubscriber) Receive(ev *open.MatchEvent) error {
	handler, ok := s.eventHandlers[ev.MatchEventType]
	if !ok {
		logger.Debugf("%s handler not registered", ev.MatchEventType.String())
		return nil
	}

	return handler(ev)
}

func(s *KafkaSubscriber) Start() {
	kafka_wrapper.RegisterConsumerFromBeginning(s, s.brokerList, s.topics, s.group)
}

func(s *KafkaSubscriber) RegisterEventHandler(evType open.MatchEventType, handler eventHandler) {
	if _, ok :=  s.eventHandlers[evType];ok {
		logger.DPanicf("%s handler has been registered previously", evType.String())
	}
	s.eventHandlers[evType] = handler
}
