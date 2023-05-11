// entities
// @author LanguageY++2013 2022/11/9 14:37
// @company soulgame
package entities

import (
	"github.com/Languege/flexmatch/service/match/pubsub"
	"sync"
	logger_pubsub "github.com/Languege/flexmatch/service/match/pubsub/logging"
)

//匹配事件发布， 默认LoggerPublisher

var (
	publisher pubsub.Publisher = logger_pubsub.NewLoggerPublisher()
	publisherInit sync.Once
)

func SetPublisher(pub pubsub.Publisher) {
	publisherInit.Do(func() {
		publisher = pub
	})
}