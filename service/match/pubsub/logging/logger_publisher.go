// logging
// @author LanguageY++2013 2023/5/11 17:17
// @company soulgame
package logging

import (
	"github.com/Languege/flexmatch/common/logger"
	"github.com/Languege/flexmatch/service/match/proto/open"
	"go.uber.org/zap"
)

type LoggerPublisher struct {
}

func NewLoggerPublisher()  *LoggerPublisher {
	return &LoggerPublisher{}
}

func (l LoggerPublisher) Name() string {
	return "logger"
}

func (l LoggerPublisher) Send(topic string, ev *open.MatchEvent) error {
	logger.Infow(ev.MatchEventType.String(), zap.String("topic", topic),
		zap.String("evEncodeType", "protobuf/text"),
		zap.String("ev", ev.String()))

	return nil
}